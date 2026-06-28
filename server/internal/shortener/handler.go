package shortener

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"urlshortener/config"
	"urlshortener/internal/auth"
	appmw "urlshortener/internal/platform/middleware"
	"urlshortener/pkg/response"
)

// Handler traduce HTTP <-> service para el dominio shortener.
type Handler struct {
	svc     ShortenerService
	baseURL string
}

func NewHandler(svc ShortenerService, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: baseURL}
}

// RegisterRoutes arma repos -> service -> handler y engancha las rutas bajo
// /api/v1/urls, todas protegidas por auth. Devuelve error si falla el bootstrap
// de índices de Mongo, para abortar el arranque.
func RegisterRoutes(api *echo.Group, db *mongo.Database, rdb *redis.Client, cfg *config.Config) error {
	repo, err := NewURLRepository(db)
	if err != nil {
		return err
	}

	userRepo, err := auth.NewUserRepository(db)
	if err != nil {
		return err
	}
	cache := NewCacheRepository(rdb)
	svc := NewShortenerService(repo, cache, userRepo, cfg.FreePlanMaxActiveURLs, cfg.BaseURL)
	h := NewHandler(svc, cfg.BaseURL)

	// Mapeo del error de cupo con el límite real interpolado.
	response.RegisterErrorMapping(auth.ErrPlanLimitExceeded, http.StatusForbidden, "PLAN_LIMIT_EXCEEDED",
		fmt.Sprintf("Tu plan Free permite hasta %d URLs activas. Desactiva alguna o mejora tu plan.", cfg.FreePlanMaxActiveURLs))

	// Rate limit de creación: por usuario autenticado (no por IP), N por minuto.
	createLimiter := appmw.RateLimit(rdb, "rl:create", cfg.RateLimitCreatePerMin, time.Minute, userIDExtractor)

	g := api.Group("/urls", appmw.Auth(cfg.JWTSecret))
	g.POST("", h.Create, createLimiter)
	g.GET("", h.List)
	g.GET("/check-alias", h.CheckAlias)
	// Detalle por shortCode. Segmento estático "code" para no chocar con el nombre de
	// param ":id" usado por las rutas PATCH/DELETE (Echo comparte el árbol de rutas).
	g.GET("/code/:shortCode", h.Get)
	g.PATCH("/:id/active", h.ToggleActive)
	g.DELETE("/:id", h.Delete)
	return nil
}

func (h *Handler) Create(c echo.Context) error {
	ownerID := currentUserID(c)

	var req CreateURLRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	url, err := h.svc.CreateShortURL(c.Request().Context(), ownerID, req.OriginalURL, req.CustomAlias)
	if err != nil {
		// INVALID_URL lleva el motivo específico (esquema/local/privada/mismo dominio)
		// como mensaje al cliente; el resto se mapea por el error handler global.
		if errors.Is(err, ErrInvalidURL) {
			return response.Fail(c, http.StatusBadRequest, "INVALID_URL", err.Error())
		}
		return err // ErrAliasTaken/ErrCodeGenFailed mapeados por el error handler global
	}
	return response.Success(c, http.StatusCreated, toURLResponse(url, h.baseURL))
}

func (h *Handler) List(c echo.Context) error {
	ownerID := currentUserID(c)
	page, limit := parsePagination(c)

	filter := ListFilter{Search: strings.TrimSpace(c.QueryParam("search"))}
	if v := c.QueryParam("isActive"); v != "" {
		if v != "true" && v != "false" {
			return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR",
				"El parámetro 'isActive' debe ser 'true' o 'false'.")
		}
		active := v == "true"
		filter.IsActive = &active
	}

	urls, total, err := h.svc.ListUserURLs(c.Request().Context(), ownerID, filter, page, limit)
	if err != nil {
		return err
	}

	items := make([]URLResponse, len(urls))
	for i, u := range urls {
		items[i] = toURLResponse(u, h.baseURL)
	}

	return response.SuccessPaginated(c, http.StatusOK, items, response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages(total, limit),
	})
}

// Get devuelve los metadatos + QR de una sola URL del usuario, por su shortCode.
// Una URL ajena o inexistente devuelve FORBIDDEN (no filtra existencia).
func (h *Handler) Get(c echo.Context) error {
	ownerID := currentUserID(c)
	shortCode := c.Param("shortCode")

	url, err := h.svc.GetOwnedURLByCode(c.Request().Context(), ownerID, shortCode)
	if err != nil {
		return err // ErrNotOwner -> 403 FORBIDDEN vía error handler global
	}
	return response.Success(c, http.StatusOK, toURLResponse(url, h.baseURL))
}

func (h *Handler) CheckAlias(c echo.Context) error {
	alias := c.QueryParam("alias")
	if alias == "" {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "El parámetro 'alias' es obligatorio.")
	}

	available, err := h.svc.CheckAliasAvailability(c.Request().Context(), alias)
	if err != nil {
		return err
	}
	return response.Success(c, http.StatusOK, CheckAliasResponse{Alias: alias, Available: available})
}

func (h *Handler) ToggleActive(c echo.Context) error {
	ownerID := currentUserID(c)
	urlID := c.Param("id")

	var req ToggleActiveRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}

	if err := h.svc.ToggleActive(c.Request().Context(), ownerID, urlID, req.IsActive); err != nil {
		return err // ErrNotOwner -> 403 vía error handler global
	}
	return response.Success(c, http.StatusOK, nil)
}

func (h *Handler) Delete(c echo.Context) error {
	ownerID := currentUserID(c)
	urlID := c.Param("id")

	if err := h.svc.DeleteURL(c.Request().Context(), ownerID, urlID); err != nil {
		return err // ErrNotOwner -> 403 vía error handler global
	}
	return response.Success(c, http.StatusOK, nil)
}

// --- helpers ---

// currentUserID lee el userID inyectado por el auth_middleware.
func currentUserID(c echo.Context) string {
	id, _ := c.Get(appmw.ContextUserIDKey).(string)
	return id
}

// userIDExtractor deriva la clave de rate limit a partir del usuario autenticado.
func userIDExtractor(c echo.Context) string {
	return currentUserID(c)
}

// parsePagination lee ?page/?limit con defaults (1/20) y clamp de limit a 100.
func parsePagination(c echo.Context) (page, limit int) {
	page, limit = 1, 20
	if v, err := strconv.Atoi(c.QueryParam("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.QueryParam("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func totalPages(total int64, limit int) int {
	if limit <= 0 || total == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}
