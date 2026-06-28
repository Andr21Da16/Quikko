package realtime

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"urlshortener/config"
	"urlshortener/internal/platform/jwt"
	"urlshortener/internal/shortener"
	"urlshortener/pkg/response"
)

// Handler hace el upgrade HTTP -> WebSocket y arranca las goroutines del cliente.
type Handler struct {
	hub       *Hub
	urls      shortener.URLRepository
	jwtSecret string
	upgrader  websocket.Upgrader
}

func NewHandler(hub *Hub, urls shortener.URLRepository, jwtSecret string) *Handler {
	return &Handler{
		hub:       hub,
		urls:      urls,
		jwtSecret: jwtSecret,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Sin frontend dedicado aún: se acepta cualquier origen. Endurecer con
			// una allowlist cuando exista el dashboard.
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

// RegisterRoutes monta GET /ws en la RAÍZ (no bajo /api/v1) para que el cliente
// se conecte a `wss://host/ws?token=...`, igual de coherente que /:code: Echo da
// prioridad a las rutas estáticas, así que /ws no colisiona con /:code.
//

func RegisterRoutes(e *echo.Echo, hub *Hub, db *mongo.Database, cfg *config.Config) error {
	urlRepo, err := shortener.NewURLRepository(db)
	if err != nil {
		return err
	}
	h := NewHandler(hub, urlRepo, cfg.JWTSecret)
	e.GET("/ws", h.ServeWS)
	return nil
}

// ServeWS valida el JWT (query param `token`) y, solo si es válido, hace el
// upgrade. Un token ausente/ inválido se rechaza con 401 ANTES del upgrade (mejor
// UX que aceptar y cerrar). Tras conectar, el cliente queda suscrito de forma
// automática al room de su usuario.
func (h *Handler) ServeWS(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Falta el token de acceso.")
	}

	claims, err := jwt.ValidateToken(token, h.jwtSecret)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "El token ha expirado. Inicia sesión nuevamente.")
		}
		return response.Fail(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "El token es inválido.")
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		// El upgrader ya escribió la respuesta de error; solo registramos.
		slog.Error("realtime: fallo en el upgrade a WebSocket", "userID", claims.UserID, "err", err)
		return nil
	}

	client := newClient(h.hub, conn, h.urls, claims.UserID)
	// Suscripción automática al canal del propio usuario (sin subscribe explícito).
	client.joinRoom("user:" + claims.UserID)

	go client.writePump()
	go client.readPump()

	return nil
}
