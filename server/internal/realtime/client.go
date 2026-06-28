package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"urlshortener/internal/shortener"
)

const (
	// writeWait es el plazo máximo para completar una escritura al socket.
	writeWait = 10 * time.Second
	// pongWait es cuánto esperamos un pong del cliente antes de dar por muerta la
	// conexión; pingPeriod debe ser menor para enviar pings a tiempo.
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize acota los mensajes entrantes (subscribe/unsubscribe son pequeños).
	maxMessageSize = 1024
	// sendBufferSize es el buffer del canal de salida por cliente. Al llenarse, el
	// hub descarta mensajes (ver Hub.Broadcast) en vez de bloquear.
	sendBufferSize = 16
)

// Client representa una conexión WebSocket viva. Tiene una goroutine de lectura y
// una de escritura; el canal send las desacopla del hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	urls   shortener.URLRepository // para validar ownership en subscribe
	userID string
	send   chan WSMessage

	mu    sync.Mutex      // protege rooms
	rooms map[string]bool // rooms a los que está unido, para limpiar al desconectar

	closeOnce sync.Once
	done      chan struct{} // se cierra una vez al desconectar
}

func newClient(hub *Hub, conn *websocket.Conn, urls shortener.URLRepository, userID string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		urls:   urls,
		userID: userID,
		send:   make(chan WSMessage, sendBufferSize),
		rooms:  make(map[string]bool),
		done:   make(chan struct{}),
	}
}

// joinRoom registra al cliente en el room y lo recuerda para limpiarlo al cerrar.
func (c *Client) joinRoom(room string) {
	c.mu.Lock()
	c.rooms[room] = true
	c.mu.Unlock()
	c.hub.Register(c, room)
}

// leaveRoom saca al cliente del room.
func (c *Client) leaveRoom(room string) {
	c.mu.Lock()
	delete(c.rooms, room)
	c.mu.Unlock()
	c.hub.Unregister(c, room)
}

// cleanup desregistra al cliente de TODOS sus rooms y cierra la conexión. Se
// ejecuta una sola vez (sync.Once), la llamen readPump o writePump. El orden
// importa: primero se quita de los rooms (así Broadcast deja de apuntarle) y solo
// después se cierra `done`/`conn`; el canal send nunca se cierra, para que un
// Broadcast concurrente jamás escriba en un canal cerrado (panic).
func (c *Client) cleanup() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		rooms := make([]string, 0, len(c.rooms))
		for r := range c.rooms {
			rooms = append(rooms, r)
		}
		c.rooms = make(map[string]bool)
		c.mu.Unlock()

		for _, r := range rooms {
			c.hub.Unregister(c, r)
		}

		close(c.done)
		c.conn.Close()
	})
}

// readPump consume mensajes entrantes (subscribe/unsubscribe). Al primer error de
// lectura (cliente se fue, red caída) limpia y termina.
func (c *Client) readPump() {
	defer c.cleanup()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("realtime: cierre inesperado de WS", "userID", c.userID, "err", err)
			}
			return
		}
		c.handleMessage(data)
	}
}

func (c *Client) handleMessage(data []byte) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.trySend(newMessage(MsgTypeError, ErrorPayload{Code: "VALIDATION_ERROR", Message: "Mensaje WebSocket malformado."}))
		return
	}

	switch msg.Type {
	case MsgTypeSubscribe:
		c.handleSubscribe(msg.Payload)
	case MsgTypeUnsubscribe:
		c.handleUnsubscribe(msg.Payload)
	default:
		c.trySend(newMessage(MsgTypeError, ErrorPayload{Code: "VALIDATION_ERROR", Message: "Tipo de mensaje no soportado."}))
	}
}

// handleSubscribe valida ownership ANTES de unir al cliente al room url:{code}:
// un usuario no puede escuchar clics de URLs ajenas. Mismo error tanto si la URL
// no existe como si es de otro dueño (no se filtra existencia).
func (c *Client) handleSubscribe(payload json.RawMessage) {
	var p SubscribePayload
	if err := json.Unmarshal(payload, &p); err != nil || p.ShortCode == "" {
		c.trySend(newMessage(MsgTypeError, ErrorPayload{Code: "VALIDATION_ERROR", Message: "Se requiere un shortCode."}))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url, err := c.urls.FindByCode(ctx, p.ShortCode)
	if err != nil || url.OwnerID != c.userID {
		c.trySend(newMessage(MsgTypeError, ErrorPayload{Code: "FORBIDDEN", Message: "No tienes acceso a esta URL."}))
		return
	}

	c.joinRoom("url:" + p.ShortCode)
}

func (c *Client) handleUnsubscribe(payload json.RawMessage) {
	var p SubscribePayload
	if err := json.Unmarshal(payload, &p); err != nil || p.ShortCode == "" {
		return
	}
	c.leaveRoom("url:" + p.ShortCode)
}

// trySend encola un mensaje hacia el propio cliente sin bloquear la goroutine de
// lectura (si la de escritura está lenta/muerta, se descarta o se corta por done).
func (c *Client) trySend(msg WSMessage) {
	select {
	case c.send <- msg:
	case <-c.done:
	default:
		slog.Warn("realtime: no se pudo enviar mensaje al cliente, buffer lleno", "userID", c.userID)
	}
}

// writePump escribe los mensajes encolados al socket y envía pings periódicos para
// mantener viva la conexión y detectar desconexiones. Es el ÚNICO escritor del
// socket (gorilla exige un solo escritor concurrente).
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.cleanup()
	}()

	for {
		select {
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			// readPump ya inició el cierre: intentamos un close frame limpio (best-effort).
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}
