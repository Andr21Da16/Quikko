package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"urlshortener/config"
	"urlshortener/internal/platform/jwt"
	appmw "urlshortener/internal/platform/middleware"
	"urlshortener/pkg/response"
)

// Handler traduce HTTP <-> service para el dominio auth.
type Handler struct {
	svc       AuthService
	jwtSecret string
}

func NewHandler(svc AuthService, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

// RegisterRoutes arma repo -> service -> handler y engancha las rutas de auth
// bajo /api/v1/auth. Devuelve error si la infraestructura (índice Mongo) falla,
// para que main.go aborte el arranque.
//
// urls es el adaptador sobre las URLs del usuario (provisto por shortener vía
// main.go) que el service necesita para el resumen y el borrado en cascada de
// cuenta (Agent 9). Auth no importa shortener (evita ciclo): consume la interfaz.
func RegisterRoutes(api *echo.Group, db *mongo.Database, rdb *redis.Client, cfg *config.Config, urls AccountURLStore) error {
	repo, err := NewUserRepository(db)
	if err != nil {
		return err
	}
	svc := NewAuthService(repo, urls, cfg.JWTSecret, cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL, cfg.FreePlanMaxActiveURLs)
	h := NewHandler(svc, cfg.JWTSecret)

	// Rate limit anti brute-force por IP por minuto. Login y registro usan contadores
	// SEPARADOS (claves Redis y env vars distintas): un flood de registros no debe agotar
	// el cupo de logins legítimos ni viceversa, y login —el más sensible a fuerza bruta de
	// contraseñas— se ajusta de forma independiente (Agent 25).
	registerLimiter := appmw.RateLimit(rdb, "rl:auth", cfg.AuthRateLimitPerMin, time.Minute, appmw.ClientIPExtractor)
	loginLimiter := appmw.RateLimit(rdb, "rl:login", cfg.LoginRateLimitPerMin, time.Minute, appmw.ClientIPExtractor)
	protected := appmw.Auth(cfg.JWTSecret)

	g := api.Group("/auth")
	g.POST("/register", h.Register, registerLimiter)
	g.POST("/login", h.Login, loginLimiter)
	g.POST("/refresh", h.Refresh)
	g.GET("/me", h.Me, protected)
	g.PATCH("/me/plan", h.UpdatePlan, protected)
	g.PATCH("/me/password", h.ChangePassword, protected)
	g.DELETE("/me", h.DeleteAccount, protected)
	g.GET("/me/summary", h.Summary, protected)
	return nil
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	ctx := c.Request().Context()
	user, err := h.svc.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			return response.Fail(c, http.StatusConflict, "AUTH_EMAIL_TAKEN", "El email ya está registrado.")
		}
		return err // 500 vía error handler global
	}

	access, refresh, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusCreated, AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         toUserDTO(user),
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	ctx := c.Request().Context()
	access, refresh, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Email o contraseña incorrectos.")
		}
		return err
	}

	user, err := h.userFromToken(ctx, access)
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         toUserDTO(user),
	})
}

func (h *Handler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	access, err := h.svc.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "El refresh token ha expirado. Inicia sesión nuevamente.")
		}
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El refresh token es inválido.")
	}

	return response.Success(c, http.StatusOK, AuthResponse{AccessToken: access})
}

func (h *Handler) Me(c echo.Context) error {
	userID, _ := c.Get(appmw.ContextUserIDKey).(string)
	if userID == "" {
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El token es inválido.")
	}

	user, err := h.svc.GetByID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El usuario ya no existe.")
		}
		return err
	}

	return response.Success(c, http.StatusOK, toUserDTO(user))
}

// UpdatePlan cambia el plan del propio usuario autenticado.
//
// TODO: en producción esto requeriría validación de pago o rol admin. Por ahora
// cualquier usuario autenticado puede cambiar su propio plan: es una herramienta
// de prueba mientras no hay pasarela de pago.
func (h *Handler) UpdatePlan(c echo.Context) error {
	userID, _ := c.Get(appmw.ContextUserIDKey).(string)
	if userID == "" {
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El token es inválido.")
	}

	var req UpdatePlanRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	user, err := h.svc.UpdatePlan(c.Request().Context(), userID, Plan(req.Plan))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El usuario ya no existe.")
		}
		return err
	}

	return response.Success(c, http.StatusOK, toUserDTO(user))
}

// ChangePassword cambia la password del usuario autenticado, exigiendo la actual.
func (h *Handler) ChangePassword(c echo.Context) error {
	userID, _ := c.Get(appmw.ContextUserIDKey).(string)
	if userID == "" {
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El token es inválido.")
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	err := h.svc.ChangePassword(c.Request().Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "La contraseña actual es incorrecta.")
		}
		if errors.Is(err, ErrUserNotFound) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El usuario ya no existe.")
		}
		return err
	}
	return response.Success(c, http.StatusOK, nil)
}

// DeleteAccount elimina la cuenta del usuario y todos sus datos asociados, exigiendo
// la password actual para confirmar identidad.
func (h *Handler) DeleteAccount(c echo.Context) error {
	userID, _ := c.Get(appmw.ContextUserIDKey).(string)
	if userID == "" {
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El token es inválido.")
	}

	var req DeleteAccountRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de la solicitud inválido.")
	}
	if err := c.Validate(&req); err != nil {
		return response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	err := h.svc.DeleteAccount(c.Request().Context(), userID, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "La contraseña es incorrecta.")
		}
		if errors.Is(err, ErrUserNotFound) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El usuario ya no existe.")
		}
		return err
	}
	return response.Success(c, http.StatusOK, nil)
}

// Summary devuelve el resumen de cuenta (plan, límites, total de clics, fecha de
// registro) para alimentar el header/sidebar del dashboard.
func (h *Handler) Summary(c echo.Context) error {
	userID, _ := c.Get(appmw.ContextUserIDKey).(string)
	if userID == "" {
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El token es inválido.")
	}

	summary, err := h.svc.GetAccountSummary(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El usuario ya no existe.")
		}
		return err
	}
	return response.Success(c, http.StatusOK, toAccountSummaryResponse(summary))
}

// userFromToken resuelve el userID embebido en un access token recién emitido
// y carga el usuario, para poblar el campo User de AuthResponse en el login.
func (h *Handler) userFromToken(ctx context.Context, accessToken string) (*User, error) {
	claims, err := jwt.ValidateToken(accessToken, h.jwtSecret)
	if err != nil {
		return nil, err
	}
	return h.svc.GetByID(ctx, claims.UserID)
}
