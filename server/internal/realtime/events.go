// Package realtime implementa el hub de WebSocket que empuja eventos de clic al
// dashboard sin polling. Patrón de "rooms": cada conexión se suscribe automática
// al canal de su usuario (user:{userID}, derivado del JWT) y opcionalmente a
// canales de URLs propias (url:{shortCode}, previa validación de ownership).
//
// Implementa analytics.EventPublisher: recibe cada ClickEvent ya
// procesado y lo difunde a los rooms correspondientes de forma no bloqueante.
package realtime

import (
	"encoding/json"
	"log/slog"
	"time"
)

// WSMessageType discrimina el tipo de mensaje del protocolo WebSocket.
type WSMessageType string

const (
	MsgTypeSubscribe   WSMessageType = "subscribe"   // cliente -> server
	MsgTypeUnsubscribe WSMessageType = "unsubscribe" // cliente -> server
	MsgTypeClickEvent  WSMessageType = "click_event" // server -> cliente
	MsgTypeError       WSMessageType = "error"       // server -> cliente
)

// WSMessage es el sobre de todo mensaje que cruza el socket. Payload se deja como
// json.RawMessage para no re-serializar al difundir: se marshalea una vez y se
// reenvía tal cual a todos los clientes del room.
type WSMessage struct {
	Type    WSMessageType   `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SubscribePayload acompaña a subscribe/unsubscribe (cliente -> server).
type SubscribePayload struct {
	ShortCode string `json:"shortCode"`
}

// ClickEventPayload es lo que recibe el cliente por cada clic (server -> cliente).
type ClickEventPayload struct {
	ShortCode  string    `json:"shortCode"`
	Country    string    `json:"country"`
	DeviceType string    `json:"deviceType"`
	Browser    string    `json:"browser"`
	Timestamp  time.Time `json:"timestamp"`
}

// ErrorPayload comunica un error al cliente (ej. suscripción a URL ajena).
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// newMessage serializa un payload dentro de un WSMessage. El marshal de nuestros
// payloads estáticos no debería fallar; si lo hace, se loguea y se devuelve un
// mensaje con payload vacío en vez de romper el flujo de difusión.
func newMessage(t WSMessageType, payload interface{}) WSMessage {
	raw, err := json.Marshal(payload)
	if err != nil {
		slog.Error("realtime: no se pudo serializar el payload del mensaje WS", "type", t, "err", err)
		return WSMessage{Type: t}
	}
	return WSMessage{Type: t, Payload: raw}
}
