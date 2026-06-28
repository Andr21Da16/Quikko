package realtime

import (
	"log/slog"
	"sync"

	"urlshortener/internal/analytics"
)

// Hub mantiene los rooms y difunde mensajes. Es el único objeto compartido entre
// todas las conexiones; su mutex protege el mapa de rooms ante acceso concurrente

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]bool // key: "user:{id}" | "url:{code}"
}

// NewHub crea un hub vacío. Es a la vez el EventPublisher que Analytics inyecta en redirect para republicar cada clic por WebSocket.
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*Client]bool)}
}

// Register añade al cliente al room (creándolo si no existía).
func (h *Hub) Register(c *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.rooms[room]
	if !ok {
		clients = make(map[*Client]bool)
		h.rooms[room] = clients
	}
	clients[c] = true
}

// Unregister quita al cliente del room y elimina el room si queda vacío (libera
// memoria del hub para que no se acumulen rooms muertos).
func (h *Hub) Unregister(c *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[room]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
}

// Broadcast envía un mensaje a todos los clientes de un room. Es NO BLOQUEANTE:
// el envío al canal de cada cliente usa select/default, de modo que un cliente
// lento (buffer lleno) no traba la difusión al resto ni al caller.
//
// Política ante buffer lleno: se DESCARTA este mensaje para ese cliente concreto
// (el nuevo, no el más antiguo: el viejo ya está encolado y se entregará). Se
// prefiere perder un evento de métrica de un cliente saturado antes que bloquear
// el hub entero o crecer memoria sin límite.
func (h *Hub) Broadcast(room string, msg WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.rooms[room] {
		select {
		case c.send <- msg:
		default:
			slog.Warn("realtime: buffer del cliente lleno, evento descartado", "room", room, "userID", c.userID)
		}
	}
}

// PublishClickEvent implementa analytics.EventPublisher. Fire-and-forget: nunca
// devuelve error ni bloquea al caller (Agent 4) más allá de unos envíos no
// bloqueantes a canales. Difunde a los dos rooms relevantes: el del dueño de la
// URL y el de la URL puntual.
func (h *Hub) PublishClickEvent(event analytics.ClickEvent) {
	msg := newMessage(MsgTypeClickEvent, ClickEventPayload{
		ShortCode:  event.ShortCode,
		Country:    event.Country,
		DeviceType: event.DeviceType,
		Browser:    event.Browser,
		Timestamp:  event.Timestamp,
	})
	h.Broadcast("user:"+event.OwnerID, msg)
	h.Broadcast("url:"+event.ShortCode, msg)
}
