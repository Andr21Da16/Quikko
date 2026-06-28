package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"urlshortener/internal/platform/jwt"
	"urlshortener/pkg/response"
)

// bcryptCost equilibra seguridad y latencia de login/registro.
const bcryptCost = 12

func init() {
	// Registro central de mapeo de errores de dominio -> HTTP
	response.RegisterErrorMapping(ErrEmailTaken, http.StatusConflict, "AUTH_EMAIL_TAKEN",
		"El email ya está registrado.")
	response.RegisterErrorMapping(ErrInvalidCredentials, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS",
		"Email o contraseña incorrectos.")
}

// AuthService define la lógica de negocio de autenticación. GetByID es necesario
// para el endpoint protegido GET /me (resuelve el userID del JWT a un usuario).
type AuthService interface {
	Register(ctx context.Context, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, err error)
	GetByID(ctx context.Context, id string) (*User, error)
	// UpdatePlan cambia el plan del propio usuario y devuelve el usuario ya actualizado.
	UpdatePlan(ctx context.Context, id string, plan Plan) (*User, error)
	// ChangePassword verifica la password actual y la reemplaza (Agent 9, Parte A.1).
	// Devuelve ErrInvalidCredentials si la actual no coincide (mismo error que login).
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	// DeleteAccount verifica la password y elimina la cuenta y todos sus datos
	// asociados (URLs en Mongo + entradas de cache), en cascada (Agent 9, Parte A.2).
	DeleteAccount(ctx context.Context, userID, password string) error
	// GetAccountSummary junta plan, límites y métricas para el dashboard (Parte A.3).
	GetAccountSummary(ctx context.Context, userID string) (*AccountSummary, error)
}

type authService struct {
	repo            UserRepository
	urls            AccountURLStore // URLs del usuario para resumen/borrado
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	freePlanMaxURLs int // cupo del plan Free, para exponerlo en el resumen de cuenta
}

// NewAuthService construye el service con sus dependencias inyectadas. urls es el
// adaptador sobre las URLs del usuario (provisto por shortener vía main.go) para el
// resumen y el borrado en cascada de cuenta.
func NewAuthService(repo UserRepository, urls AccountURLStore, jwtSecret string, accessTTL, refreshTTL time.Duration, freePlanMaxURLs int) AuthService {
	return &authService{
		repo:            repo,
		urls:            urls,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		freePlanMaxURLs: freePlanMaxURLs,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("auth: error hasheando password: %w", err)
	}

	user := &User{
		Email:        email,
		PasswordHash: string(hash),
		Plan:         PlanFree, // todo usuario nuevo arranca en el plan Free
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err // ErrEmailTaken se propaga tal cual
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// No revelamos si el fallo fue email inexistente o password incorrecta.
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return s.issueTokens(user.ID)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := jwt.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return "", err // jwt.ErrTokenExpired / jwt.ErrTokenInvalid se propagan
	}
	if claims.Type != jwt.TypeRefresh {
		return "", jwt.ErrTokenInvalid
	}

	// Opcional: confirmar que el usuario sigue existiendo antes de renovar.
	if _, err := s.repo.FindByID(ctx, claims.UserID); err != nil {
		return "", jwt.ErrTokenInvalid
	}

	access, err := jwt.GenerateAccessToken(claims.UserID, s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("auth: error generando access token: %w", err)
	}
	return access, nil
}

func (s *authService) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *authService) UpdatePlan(ctx context.Context, id string, plan Plan) (*User, error) {
	if err := s.repo.UpdatePlan(ctx, id, plan); err != nil {
		return nil, err // ErrUserNotFound se propaga
	}
	// Devolvemos el usuario ya actualizado para que el handler lo refleje en la respuesta.
	return s.repo.FindByID(ctx, id)
}

func (s *authService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err // ErrUserNotFound se propaga
	}
	// La password actual debe coincidir; si no, mismo error que un login fallido.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("auth: error hasheando la nueva password: %w", err)
	}
	return s.repo.UpdatePassword(ctx, userID, string(hash))
}

// DeleteAccount verifica identidad y orquesta el borrado en cascada:
// purga el cache de redirección de cada URL, borra las URLs en Mongo y, por último,
// el usuario. Alcance consciente: los clics históricos en InfluxDB NO se eliminan
// (quedan huérfanos); un borrado de series por ownerId queda fuera de alcance ahora.
func (s *authService) DeleteAccount(ctx context.Context, userID, password string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err // ErrUserNotFound se propaga
	}
	// Verificar password antes de borrar: evita que una sesión robada elimine la
	// cuenta sin confirmar identidad.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	// 1) Purgar el cache de redirección de cada URL del usuario. Un fallo aquí se
	//    loguea pero no aborta: Redis tiene TTL y Mongo es la fuente de verdad.
	codes, err := s.urls.ShortCodesByOwner(ctx, userID)
	if err != nil {
		return fmt.Errorf("auth: no se pudieron listar las URLs del usuario para borrar la cuenta: %w", err)
	}
	for _, code := range codes {
		if err := s.urls.PurgeCache(ctx, code); err != nil {
			slog.Error("auth: no se pudo purgar el cache al eliminar la cuenta", "shortCode", code, "err", err)
		}
	}

	// 2) Borrar todas las URLs del usuario en Mongo.
	if err := s.urls.DeleteAllByOwner(ctx, userID); err != nil {
		return fmt.Errorf("auth: no se pudieron eliminar las URLs del usuario: %w", err)
	}

	// 3) Borrar el usuario.
	return s.repo.Delete(ctx, userID)
}

func (s *authService) GetAccountSummary(ctx context.Context, userID string) (*AccountSummary, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err // ErrUserNotFound se propaga
	}
	activeCount, err := s.urls.CountActiveByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: error contando URLs activas para el resumen: %w", err)
	}
	totalClicks, err := s.urls.SumClicksByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: error sumando clics para el resumen: %w", err)
	}

	// El límite solo aplica al plan Free; Pro es ilimitado (nil).
	var limit *int
	if user.Plan == PlanFree {
		l := s.freePlanMaxURLs
		limit = &l
	}

	return &AccountSummary{
		Email:           user.Email,
		Plan:            user.Plan,
		CreatedAt:       user.CreatedAt,
		ActiveURLsCount: activeCount,
		ActiveURLsLimit: limit,
		TotalClicks:     totalClicks,
	}, nil
}

// issueTokens genera el par access/refresh para un userID ya autenticado.
func (s *authService) issueTokens(userID string) (string, string, error) {
	access, err := jwt.GenerateAccessToken(userID, s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("auth: error generando access token: %w", err)
	}
	refresh, err := jwt.GenerateRefreshToken(userID, s.jwtSecret, s.refreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("auth: error generando refresh token: %w", err)
	}
	return access, refresh, nil
}
