package redirect

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"urlshortener/config"
	appmw "urlshortener/internal/platform/middleware"
	"urlshortener/internal/shortener"
)

// Rutas del frontend a las que el redirect manda los errores de negocio.
const (
	frontendPathNotFound = "/link-no-encontrado"
	frontendPathInactive = "/link-inactivo"
)

// Handler sirve el endpoint público de redirección.
type Handler struct {
	svc RedirectService
	// frontendURL es la base del frontend para redirigir los errores de negocio a sus
	// páginas con marca (sin barra final, normalizada en NewHandler).
	frontendURL string
}

func NewHandler(svc RedirectService, frontendURL string) *Handler {
	return &Handler{svc: svc, frontendURL: strings.TrimRight(frontendURL, "/")}
}

// RegisterRoutes monta `GET /:code` en la RAÍZ del router (no bajo /api/v1) para
// que las URLs cortas sean `https://base/<code>`. Echo prioriza rutas estáticas,
// así que `/api/v1/...` y `/health` no colisionan con el parámetro `/:code`.
func RegisterRoutes(e *echo.Echo, db *mongo.Database, rdb *redis.Client, cfg *config.Config, recorder ClickRecorder) error {
	repo, err := shortener.NewURLRepository(db)
	if err != nil {
		return err
	}
	cache := shortener.NewCacheRepository(rdb)
	// GeoIP vía servicio externo (ipapi.co) con cache en Redis: reutiliza el
	// mismo cliente Redis compartido; si Redis cae, degrada a consultar el servicio.
	geo := NewCountryResolver(rdb, cfg.GeoIPHTTPTimeout)
	svc := NewRedirectService(repo, cache, geo, recorder)
	h := NewHandler(svc, cfg.FrontendURL)

	// Rate limit por IP: ventana de 1s con el máximo configurado.
	limiter := appmw.RateLimit(rdb, "rl:redirect", cfg.RateLimitRedirectPerSec, time.Second, appmw.ClientIPExtractor)

	e.GET("/:code", h.Redirect, limiter)
	return nil
}

// Redirect resuelve el código y responde un 302. El registro del clic se dispara
// async y nunca bloquea ni puede alterar la respuesta.
//
// Errores (Agent 33): GET /:code es el único endpoint que un humano abre directo en el
// navegador, así que los errores de NEGOCIO esperados redirigen (302) a páginas con marca
// del frontend (Agent 34) en vez de mostrar JSON crudo:
//   - ErrURLNotFound  -> {FRONTEND_URL}/link-no-encontrado
//   - ErrURLInactive  -> {FRONTEND_URL}/link-inactivo
//
// Un fallo interno genuino (5xx) NO se enmascara: cae al error handler global con el
// envelope JSON estándar, para poder diagnosticarlo (no es un flujo de negocio esperado).
func (h *Handler) Redirect(c echo.Context) error {
	code := c.Param("code")

	originalURL, err := h.svc.Resolve(c.Request().Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrURLNotFound):
			return c.Redirect(http.StatusFound, h.frontendURL+frontendPathNotFound)
		case errors.Is(err, shortener.ErrURLInactive):
			return c.Redirect(http.StatusFound, h.frontendURL+frontendPathInactive)
		default:
			return err // 5xx genuino -> envelope JSON estándar (diagnosticable)
		}
	}

	// Capturamos los datos del request ANTES de responder (el contexto del request
	// no sobrevive al handler) y disparamos el registro async — no se espera.
	h.svc.RecordClickAsync(code, c.RealIP(), c.Request().UserAgent())

	return c.Redirect(http.StatusFound, originalURL)
}
