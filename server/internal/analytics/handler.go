package analytics

import (
	"fmt"
	"net/http"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"urlshortener/config"
	"urlshortener/internal/auth"
	appmw "urlshortener/internal/platform/middleware"
	"urlshortener/internal/shortener"
	"urlshortener/pkg/response"
)

// Handler traduce HTTP <-> service para el dominio analytics.
type Handler struct {
	svc AnalyticsService
}

func NewHandler(svc AnalyticsService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes arma repos -> service -> handler y engancha GET /analytics/stats
// bajo /api/v1, protegido por auth. Devuelve el AnalyticsService construido para
// que main.go lo inyecte en redirect como ClickRecorder (reemplazando al
// NoopClickRecorder): así un único service es a la vez escritor y consultor.
//

func RegisterRoutes(api *echo.Group, db *mongo.Database, influxClient influxdb2.Client, cfg *config.Config, publisher EventPublisher) (AnalyticsService, error) {
	urlRepo, err := shortener.NewURLRepository(db)
	if err != nil {
		return nil, err
	}
	// UserRepository de Agent 1, reutilizado para leer el plan del usuario (Agent 7).
	userRepo, err := auth.NewUserRepository(db)
	if err != nil {
		return nil, err
	}
	clickRepo := NewClickRepository(influxClient, cfg.InfluxOrg, cfg.InfluxBucket)
	svc := NewAnalyticsService(clickRepo, urlRepo, userRepo, publisher)
	h := NewHandler(svc)

	g := api.Group("/analytics", appmw.Auth(cfg.JWTSecret))
	g.GET("/stats", h.Stats)
	g.GET("/stats/export", h.ExportCSV)
	return svc, nil
}

// Stats responde las métricas agregadas del usuario autenticado. Con shortCode
// devuelve las de esa URL (403 FORBIDDEN si no es suya); sin él, el overview de
// todas sus URLs.
func (h *Handler) Stats(c echo.Context) error {
	ownerID := currentUserID(c)

	var params StatsQueryParams
	if err := c.Bind(&params); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Parámetros de consulta inválidos.")
	}
	if err := c.Validate(&params); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	stats, err := h.svc.GetStats(c.Request().Context(), ownerID, params.ShortCode, params.Range)
	if err != nil {
		return err // ErrInvalidRange -> 400, ErrForbidden -> 403 (mapeo global)
	}

	rango := params.Range
	if rango == "" {
		rango = string(Range24h)
	}
	return response.Success(c, http.StatusOK, StatsResponse{Range: rango, Stats: *stats})
}

// ExportCSV descarga las mismas métricas que Stats como archivo CSV. Mismos query
// params y validaciones (ownership, plan/rango).
//
// Excepción documentada a la sección 4 del documento maestro: el camino feliz NO
// usa el envelope JSON estándar (un CSV no es JSON) — responde el archivo crudo con
// Content-Type text/csv y Content-Disposition de descarga. Los errores SÍ usan el
// envelope estándar (vía el error handler global o response.Fail).
func (h *Handler) ExportCSV(c echo.Context) error {
	ownerID := currentUserID(c)

	var params StatsQueryParams
	if err := c.Bind(&params); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Parámetros de consulta inválidos.")
	}
	if err := c.Validate(&params); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	csvBytes, err := h.svc.ExportStatsCSV(c.Request().Context(), ownerID, params.ShortCode, params.Range)
	if err != nil {
		return err // ErrForbidden/ErrRangeNotAllowedForPlan/... -> envelope JSON (mapeo global)
	}

	rango := params.Range
	if rango == "" {
		rango = string(Range24h)
	}
	label := "overview"
	if params.ShortCode != nil && *params.ShortCode != "" {
		label = *params.ShortCode
	}
	filename := fmt.Sprintf("stats-%s-%s.csv", label, rango)

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename=%q`, filename))
	return c.Blob(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

// currentUserID lee el userID inyectado por el auth_middleware.
func currentUserID(c echo.Context) string {
	id, _ := c.Get(appmw.ContextUserIDKey).(string)
	return id
}
