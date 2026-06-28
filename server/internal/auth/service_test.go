package auth

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"urlshortener/internal/platform/jwt"
)

// mockRepo es un UserRepository en memoria para tests de service.
type mockRepo struct {
	byEmail map[string]*User
	byID    map[string]*User
	nextID  int
}

func newMockRepo() *mockRepo {
	return &mockRepo{byEmail: map[string]*User{}, byID: map[string]*User{}}
}

func (m *mockRepo) Create(_ context.Context, user *User) error {
	if _, ok := m.byEmail[user.Email]; ok {
		return ErrEmailTaken
	}
	m.nextID++
	// IDs hex de 24 chars como ObjectID real, para que ValidateToken/GetByID funcionen.
	user.ID = padHexID(m.nextID)
	cp := *user
	m.byEmail[user.Email] = &cp
	m.byID[user.ID] = &cp
	return nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepo) UpdatePlan(_ context.Context, id string, plan Plan) error {
	u, ok := m.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	u.Plan = plan
	return nil
}

func (m *mockRepo) UpdatePassword(_ context.Context, id, passwordHash string) error {
	u, ok := m.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	u, ok := m.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(m.byID, id)
	delete(m.byEmail, u.Email)
	return nil
}

func padHexID(n int) string {
	const hexLen = 24
	s := ""
	for i := 0; i < hexLen; i++ {
		s += "0"
	}
	digits := "0123456789abcdef"
	// Coloca n en hex al final.
	out := []byte(s)
	i := hexLen - 1
	for n > 0 && i >= 0 {
		out[i] = digits[n%16]
		n /= 16
		i--
	}
	return string(out)
}

// mockURLStore implementa AccountURLStore en memoria para los tests de cuenta.
type mockURLStore struct {
	activeCount int64
	totalClicks int64
	shortCodes  []string
	purged      []string
	deletedAll  []string
}

func (m *mockURLStore) CountActiveByOwner(context.Context, string) (int64, error) {
	return m.activeCount, nil
}
func (m *mockURLStore) SumClicksByOwner(context.Context, string) (int64, error) {
	return m.totalClicks, nil
}
func (m *mockURLStore) ShortCodesByOwner(context.Context, string) ([]string, error) {
	return m.shortCodes, nil
}
func (m *mockURLStore) DeleteAllByOwner(_ context.Context, ownerID string) error {
	m.deletedAll = append(m.deletedAll, ownerID)
	return nil
}
func (m *mockURLStore) PurgeCache(_ context.Context, shortCode string) error {
	m.purged = append(m.purged, shortCode)
	return nil
}

func newTestService(repo UserRepository) AuthService {
	return NewAuthService(repo, &mockURLStore{}, "test-secret", 15*time.Minute, 7*24*time.Hour, 5)
}

func newTestServiceWithStore(repo UserRepository, store AccountURLStore) AuthService {
	return NewAuthService(repo, store, "test-secret", 15*time.Minute, 7*24*time.Hour, 5)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.Register(ctx, "a@b.com", "password123"); err != nil {
		t.Fatalf("primer registro debería ok, obtuvo %v", err)
	}
	_, err := svc.Register(ctx, "a@b.com", "otraPassword")
	if err != ErrEmailTaken {
		t.Fatalf("esperaba ErrEmailTaken, obtuvo %v", err)
	}
}

// TestRegister_DefaultsToFreePlan: todo usuario nuevo arranca en Free (Agent 7).
func TestRegister_DefaultsToFreePlan(t *testing.T) {
	svc := newTestService(newMockRepo())
	user, err := svc.Register(context.Background(), "p@x.com", "password123")
	if err != nil {
		t.Fatalf("registro falló: %v", err)
	}
	if user.Plan != PlanFree {
		t.Fatalf("esperaba plan free por defecto, obtuvo %q", user.Plan)
	}
}

// TestUpdatePlan: cambiar el plan se refleja al releer el usuario (Agent 7).
func TestUpdatePlan(t *testing.T) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	user, err := svc.Register(ctx, "up@x.com", "password123")
	if err != nil {
		t.Fatalf("registro falló: %v", err)
	}

	updated, err := svc.UpdatePlan(ctx, user.ID, PlanPro)
	if err != nil {
		t.Fatalf("cambio de plan falló: %v", err)
	}
	if updated.Plan != PlanPro {
		t.Fatalf("esperaba plan pro tras el cambio, obtuvo %q", updated.Plan)
	}
	// Releer confirma persistencia.
	reread, _ := svc.GetByID(ctx, user.ID)
	if reread.Plan != PlanPro {
		t.Fatalf("el cambio de plan no persistió, GetByID devolvió %q", reread.Plan)
	}
}

// TestLogin_SameErrorForBadPasswordAndUnknownEmail es el criterio de seguridad:
// no se debe revelar si falló el email o la password.
func TestLogin_SameErrorForBadPasswordAndUnknownEmail(t *testing.T) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	if _, err := svc.Register(ctx, "user@x.com", "correctPassword"); err != nil {
		t.Fatalf("registro falló: %v", err)
	}

	_, _, errBadPass := svc.Login(ctx, "user@x.com", "wrongPassword")
	_, _, errNoUser := svc.Login(ctx, "missing@x.com", "whatever123")

	if errBadPass != ErrInvalidCredentials {
		t.Fatalf("password incorrecta: esperaba ErrInvalidCredentials, obtuvo %v", errBadPass)
	}
	if errNoUser != ErrInvalidCredentials {
		t.Fatalf("email inexistente: esperaba ErrInvalidCredentials, obtuvo %v", errNoUser)
	}
	if errBadPass != errNoUser {
		t.Fatalf("los errores deben ser idénticos: %v vs %v", errBadPass, errNoUser)
	}
}

func TestLogin_Success_IssuesValidTokens(t *testing.T) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	if _, err := svc.Register(ctx, "ok@x.com", "password123"); err != nil {
		t.Fatalf("registro falló: %v", err)
	}

	access, refresh, err := svc.Login(ctx, "ok@x.com", "password123")
	if err != nil {
		t.Fatalf("login falló: %v", err)
	}

	ac, err := jwt.ValidateToken(access, "test-secret")
	if err != nil || ac.Type != jwt.TypeAccess {
		t.Fatalf("access token inválido: %v, type=%s", err, ac.Type)
	}
	rc, err := jwt.ValidateToken(refresh, "test-secret")
	if err != nil || rc.Type != jwt.TypeRefresh {
		t.Fatalf("refresh token inválido: %v, type=%s", err, rc.Type)
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	if _, err := svc.Register(context.Background(), "h@x.com", "superSecret1"); err != nil {
		t.Fatalf("registro falló: %v", err)
	}
	stored := repo.byEmail["h@x.com"]
	if stored.PasswordHash == "superSecret1" {
		t.Fatal("la password se guardó en texto plano")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("superSecret1")); err != nil {
		t.Fatalf("el hash almacenado no corresponde a la password: %v", err)
	}
}

func TestRefreshToken(t *testing.T) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	if _, err := svc.Register(ctx, "r@x.com", "password123"); err != nil {
		t.Fatalf("registro falló: %v", err)
	}
	_, refresh, _ := svc.Login(ctx, "r@x.com", "password123")

	t.Run("refresh válido emite nuevo access", func(t *testing.T) {
		access, err := svc.RefreshToken(ctx, refresh)
		if err != nil {
			t.Fatalf("refresh falló: %v", err)
		}
		c, err := jwt.ValidateToken(access, "test-secret")
		if err != nil || c.Type != jwt.TypeAccess {
			t.Fatalf("nuevo access token inválido: %v", err)
		}
	})

	t.Run("un access token no sirve como refresh", func(t *testing.T) {
		access, _, _ := svc.Login(ctx, "r@x.com", "password123")
		if _, err := svc.RefreshToken(ctx, access); err != jwt.ErrTokenInvalid {
			t.Fatalf("esperaba ErrTokenInvalid usando access como refresh, obtuvo %v", err)
		}
	})

	t.Run("token basura es inválido", func(t *testing.T) {
		if _, err := svc.RefreshToken(ctx, "no-soy-un-token"); err == nil {
			t.Fatal("esperaba error con token basura")
		}
	})
}

// --- Agent 9, Parte A: gestión de cuenta ---

func TestChangePassword(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	ctx := context.Background()
	user, err := svc.Register(ctx, "cp@x.com", "oldPassword1")
	if err != nil {
		t.Fatalf("registro falló: %v", err)
	}

	t.Run("password actual incorrecta devuelve ErrInvalidCredentials", func(t *testing.T) {
		if err := svc.ChangePassword(ctx, user.ID, "wrongCurrent", "newPassword1"); err != ErrInvalidCredentials {
			t.Fatalf("esperaba ErrInvalidCredentials, obtuvo %v", err)
		}
	})

	t.Run("password actual correcta actualiza el hash", func(t *testing.T) {
		if err := svc.ChangePassword(ctx, user.ID, "oldPassword1", "newPassword1"); err != nil {
			t.Fatalf("cambio falló: %v", err)
		}
		// El nuevo hash corresponde a la nueva password y NO a la vieja.
		stored := repo.byID[user.ID]
		if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("newPassword1")) != nil {
			t.Fatal("el hash no corresponde a la nueva password")
		}
		if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("oldPassword1")) == nil {
			t.Fatal("la password vieja no debería seguir siendo válida")
		}
	})
}

func TestDeleteAccount(t *testing.T) {
	repo := newMockRepo()
	store := &mockURLStore{shortCodes: []string{"abc123", "def456"}}
	svc := newTestServiceWithStore(repo, store)
	ctx := context.Background()
	user, err := svc.Register(ctx, "del@x.com", "password123")
	if err != nil {
		t.Fatalf("registro falló: %v", err)
	}

	t.Run("password incorrecta no borra nada", func(t *testing.T) {
		if err := svc.DeleteAccount(ctx, user.ID, "wrong"); err != ErrInvalidCredentials {
			t.Fatalf("esperaba ErrInvalidCredentials, obtuvo %v", err)
		}
		if _, err := svc.GetByID(ctx, user.ID); err != nil {
			t.Fatal("la cuenta no debió borrarse con password incorrecta")
		}
	})

	t.Run("password correcta orquesta el borrado en cascada", func(t *testing.T) {
		if err := svc.DeleteAccount(ctx, user.ID, "password123"); err != nil {
			t.Fatalf("borrado falló: %v", err)
		}
		// Cada shortCode se purgó del cache.
		if len(store.purged) != 2 {
			t.Fatalf("esperaba 2 purgas de cache, hubo %v", store.purged)
		}
		// Las URLs del usuario se borraron en Mongo.
		if len(store.deletedAll) != 1 || store.deletedAll[0] != user.ID {
			t.Fatalf("esperaba DeleteAllByOwner(%q), hubo %v", user.ID, store.deletedAll)
		}
		// El usuario ya no existe.
		if _, err := svc.GetByID(ctx, user.ID); err != ErrUserNotFound {
			t.Fatalf("esperaba ErrUserNotFound tras borrar, obtuvo %v", err)
		}
	})
}

func TestGetAccountSummary(t *testing.T) {
	repo := newMockRepo()
	ctx := context.Background()

	t.Run("plan Free expone el límite y suma de clics", func(t *testing.T) {
		store := &mockURLStore{activeCount: 3, totalClicks: 42}
		svc := newTestServiceWithStore(repo, store)
		user, err := svc.Register(ctx, "free@x.com", "password123")
		if err != nil {
			t.Fatalf("registro falló: %v", err)
		}

		summary, err := svc.GetAccountSummary(ctx, user.ID)
		if err != nil {
			t.Fatalf("resumen falló: %v", err)
		}
		if summary.ActiveURLsCount != 3 || summary.TotalClicks != 42 {
			t.Fatalf("agregados incorrectos: %+v", summary)
		}
		if summary.ActiveURLsLimit == nil || *summary.ActiveURLsLimit != 5 {
			t.Fatalf("Free debe exponer el límite 5, obtuvo %v", summary.ActiveURLsLimit)
		}
	})

	t.Run("plan Pro no tiene límite (nil)", func(t *testing.T) {
		store := &mockURLStore{activeCount: 999, totalClicks: 1000}
		svc := newTestServiceWithStore(repo, store)
		user, err := svc.Register(ctx, "pro@x.com", "password123")
		if err != nil {
			t.Fatalf("registro falló: %v", err)
		}
		if _, err := svc.UpdatePlan(ctx, user.ID, PlanPro); err != nil {
			t.Fatalf("cambio a Pro falló: %v", err)
		}

		summary, err := svc.GetAccountSummary(ctx, user.ID)
		if err != nil {
			t.Fatalf("resumen falló: %v", err)
		}
		if summary.ActiveURLsLimit != nil {
			t.Fatalf("Pro debe ser ilimitado (nil), obtuvo %v", *summary.ActiveURLsLimit)
		}
	})
}
