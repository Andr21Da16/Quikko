package redirect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	appmw "urlshortener/internal/platform/middleware"
	"urlshortener/internal/shortener"
)

type mockRepo struct {
	mu          sync.Mutex
	findByCode  func(code string) (*shortener.ShortURL, error)
	mongoCalled bool
	incremented chan string // recibe el shortCode en cada IncrementClicks
}

func (m *mockRepo) FindByCode(_ context.Context, code string) (*shortener.ShortURL, error) {
	m.mu.Lock()
	m.mongoCalled = true
	m.mu.Unlock()
	if m.findByCode != nil {
		return m.findByCode(code)
	}
	return nil, shortener.ErrURLNotFound
}
func (m *mockRepo) wasMongoCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mongoCalled
}

// Métodos no usados por redirect (satisfacen la interfaz URLRepository).
func (m *mockRepo) Create(context.Context, *shortener.ShortURL) error { return nil }
func (m *mockRepo) FindByID(context.Context, string) (*shortener.ShortURL, error) {
	return nil, shortener.ErrURLNotFound
}
func (m *mockRepo) FindByOwner(context.Context, string, shortener.ListFilter, int, int) ([]*shortener.ShortURL, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) CountActiveByOwner(context.Context, string) (int64, error) { return 0, nil }
func (m *mockRepo) ExistsByCode(context.Context, string) (bool, error)        { return false, nil }
func (m *mockRepo) Update(context.Context, *shortener.ShortURL) error         { return nil }
func (m *mockRepo) Delete(context.Context, string, string) error              { return nil }
func (m *mockRepo) SumClicksByOwner(context.Context, string) (int64, error)   { return 0, nil }
func (m *mockRepo) FindShortCodesByOwner(context.Context, string) ([]string, error) {
	return nil, nil
}
func (m *mockRepo) DeleteAllByOwner(context.Context, string) error { return nil }
func (m *mockRepo) IncrementClicks(_ context.Context, shortCode string) error {
	if m.incremented != nil {
		select {
		case m.incremented <- shortCode:
		default:
		}
	}
	return nil
}

type mockCache struct {
	mu     sync.Mutex
	store  map[string]string
	getErr error
}

func newMockCache() *mockCache { return &mockCache{store: map[string]string{}} }

func (m *mockCache) Get(_ context.Context, code string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return "", m.getErr
	}
	if v, ok := m.store[code]; ok {
		return v, nil
	}
	return "", shortener.ErrCacheMiss
}
func (m *mockCache) Set(_ context.Context, code, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[code] = url
	return nil
}
func (m *mockCache) Delete(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, code)
	return nil
}

type mockRecorder struct {
	err    error
	called chan ClickEvent
}

func newMockRecorder(err error) *mockRecorder {
	return &mockRecorder{err: err, called: make(chan ClickEvent, 1)}
}
func (m *mockRecorder) RecordClick(_ context.Context, e ClickEvent) error {
	select {
	case m.called <- e:
	default:
	}
	return m.err
}

func activeURL(code string) *shortener.ShortURL {
	return &shortener.ShortURL{ShortCode: code, OriginalURL: "https://go.dev", OwnerID: "owner1", IsActive: true}
}

// --- Tests de servicio---

func TestResolve_CacheHit_DoesNotTouchMongo(t *testing.T) {
	repo := &mockRepo{findByCode: func(string) (*shortener.ShortURL, error) {
		t.Fatal("Mongo no debe consultarse en un cache hit")
		return nil, nil
	}}
	cache := newMockCache()
	cache.store["abc"] = "https://cached.example"
	svc := NewRedirectService(repo, cache, noopResolver{}, NoopClickRecorder{})

	url, err := svc.Resolve(context.Background(), "abc")
	if err != nil || url != "https://cached.example" {
		t.Fatalf("esperaba la URL cacheada, obtuvo %q / %v", url, err)
	}
	if repo.wasMongoCalled() {
		t.Fatal("no debió llamarse a Mongo")
	}
}

func TestResolve_CacheMiss_FallsBackAndRepopulates(t *testing.T) {
	repo := &mockRepo{findByCode: func(c string) (*shortener.ShortURL, error) { return activeURL(c), nil }}
	cache := newMockCache() // vacío => miss
	svc := NewRedirectService(repo, cache, noopResolver{}, NoopClickRecorder{})

	url, err := svc.Resolve(context.Background(), "abc")
	if err != nil || url != "https://go.dev" {
		t.Fatalf("esperaba fallback a Mongo, obtuvo %q / %v", url, err)
	}
	if !repo.wasMongoCalled() {
		t.Fatal("debió consultarse Mongo en el miss")
	}
	// Redis debe quedar repoblado: un Get posterior devuelve el valor.
	if got, err := cache.Get(context.Background(), "abc"); err != nil || got != "https://go.dev" {
		t.Fatalf("Redis no quedó repoblado: %q / %v", got, err)
	}
}

func TestResolve_NotFound(t *testing.T) {
	repo := &mockRepo{findByCode: func(string) (*shortener.ShortURL, error) { return nil, shortener.ErrURLNotFound }}
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, NoopClickRecorder{})

	if _, err := svc.Resolve(context.Background(), "nope"); !errors.Is(err, shortener.ErrURLNotFound) {
		t.Fatalf("esperaba ErrURLNotFound, obtuvo %v", err)
	}
}

func TestResolve_Inactive(t *testing.T) {
	repo := &mockRepo{findByCode: func(c string) (*shortener.ShortURL, error) {
		u := activeURL(c)
		u.IsActive = false
		return u, nil
	}}
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, NoopClickRecorder{})

	if _, err := svc.Resolve(context.Background(), "abc"); !errors.Is(err, shortener.ErrURLInactive) {
		t.Fatalf("esperaba ErrURLInactive, obtuvo %v", err)
	}
}

// --- Tests de handler (status codes + clic async) ---

const testFrontendURL = "http://frontend.test"

func newTestServer(svc RedirectService) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = appmw.ErrorHandler
	h := NewHandler(svc, testFrontendURL)
	e.GET("/:code", h.Redirect)
	return e
}

func TestHandler_Success_302(t *testing.T) {
	repo := &mockRepo{findByCode: func(c string) (*shortener.ShortURL, error) { return activeURL(c), nil }}
	rec := newMockRecorder(nil)
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, rec)
	e := newTestServer(svc)

	res := httptest.NewRecorder()
	e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/abc", nil))

	if res.Code != http.StatusFound {
		t.Fatalf("esperaba 302, obtuvo %d", res.Code)
	}
	if loc := res.Header().Get("Location"); loc != "https://go.dev" {
		t.Fatalf("Location incorrecto: %q", loc)
	}
	waitClick(t, rec)
}

func TestHandler_NotFound_RedirectsToFrontend(t *testing.T) {
	repo := &mockRepo{findByCode: func(string) (*shortener.ShortURL, error) { return nil, shortener.ErrURLNotFound }}
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, NoopClickRecorder{})
	e := newTestServer(svc)

	res := httptest.NewRecorder()
	e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/nope", nil))

	if res.Code != http.StatusFound {
		t.Fatalf("esperaba 302, obtuvo %d (body: %s)", res.Code, res.Body.String())
	}
	if loc := res.Header().Get("Location"); loc != testFrontendURL+"/link-no-encontrado" {
		t.Fatalf("Location incorrecto: %q", loc)
	}
}

// una URL inactiva redirige (302) a la página de marca del frontend.
func TestHandler_Inactive_RedirectsToFrontend(t *testing.T) {
	repo := &mockRepo{findByCode: func(c string) (*shortener.ShortURL, error) {
		u := activeURL(c)
		u.IsActive = false
		return u, nil
	}}
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, NoopClickRecorder{})
	e := newTestServer(svc)

	res := httptest.NewRecorder()
	e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/abc", nil))

	if res.Code != http.StatusFound {
		t.Fatalf("esperaba 302, obtuvo %d (body: %s)", res.Code, res.Body.String())
	}
	if loc := res.Header().Get("Location"); loc != testFrontendURL+"/link-inactivo" {
		t.Fatalf("Location incorrecto: %q", loc)
	}
}

// un fallo interno genuino (5xx) NO se enmascara como página de frontend; sigue
// respondiendo el envelope JSON estándar para poder diagnosticarlo.
func TestHandler_GenericError_StaysJSON(t *testing.T) {
	boom := errors.New("mongo caído")
	repo := &mockRepo{findByCode: func(string) (*shortener.ShortURL, error) { return nil, boom }}
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, NoopClickRecorder{})
	e := newTestServer(svc)

	res := httptest.NewRecorder()
	e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/abc", nil))

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuvo %d (body: %s)", res.Code, res.Body.String())
	}
	if loc := res.Header().Get("Location"); loc != "" {
		t.Fatalf("un 5xx no debe redirigir; Location = %q", loc)
	}
	if ct := res.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("esperaba envelope JSON, Content-Type = %q", ct)
	}
}

// Criterio: un fallo del ClickRecorder no afecta el redirect.
func TestHandler_RecorderFailure_StillRedirects(t *testing.T) {
	repo := &mockRepo{findByCode: func(c string) (*shortener.ShortURL, error) { return activeURL(c), nil }}
	rec := newMockRecorder(errors.New("influx caído"))
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, rec)
	e := newTestServer(svc)

	res := httptest.NewRecorder()
	e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/abc", nil))

	if res.Code != http.StatusFound {
		t.Fatalf("el fallo de analítica no debe afectar el redirect: esperaba 302, obtuvo %d", res.Code)
	}
	if loc := res.Header().Get("Location"); loc != "https://go.dev" {
		t.Fatalf("Location incorrecto: %q", loc)
	}
	waitClick(t, rec) // el recorder se invocó (y su error fue tragado)
}

// cada redirect exitoso dispara IncrementClicks async.
func TestHandler_Success_IncrementsClickCounter(t *testing.T) {
	repo := &mockRepo{
		findByCode:  func(c string) (*shortener.ShortURL, error) { return activeURL(c), nil },
		incremented: make(chan string, 1),
	}
	rec := newMockRecorder(nil)
	svc := NewRedirectService(repo, newMockCache(), noopResolver{}, rec)
	e := newTestServer(svc)

	res := httptest.NewRecorder()
	e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/abc", nil))

	if res.Code != http.StatusFound {
		t.Fatalf("esperaba 302, obtuvo %d", res.Code)
	}
	select {
	case code := <-repo.incremented:
		if code != "abc" {
			t.Fatalf("IncrementClicks recibió %q, esperaba \"abc\"", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("el contador de clics nunca se incrementó (async)")
	}
}

func waitClick(t *testing.T, rec *mockRecorder) {
	t.Helper()
	select {
	case <-rec.called:
	case <-time.After(2 * time.Second):
		t.Fatal("el registro async del clic nunca se invocó")
	}
}
