package realtime

import (
	"encoding/json"
	"testing"
	"time"

	"urlshortener/internal/analytics"
)

// newTestClient construye un Client mínimo válido para probar el Hub sin un socket
// real: Broadcast solo toca send y userID.
func newTestClient(userID string, buffer int) *Client {
	return &Client{userID: userID, send: make(chan WSMessage, buffer)}
}

func drain(c *Client) []WSMessage {
	var msgs []WSMessage
	for {
		select {
		case m := <-c.send:
			msgs = append(msgs, m)
		default:
			return msgs
		}
	}
}

// Criterio: un evento publicado para un userID llega al cliente sin subscribe
// explícito (el room de usuario es automático), y también al room de la URL.
func TestPublishClickEvent_ReachesUserAndURLRooms(t *testing.T) {
	hub := NewHub()

	owner := newTestClient("owner1", 4)
	hub.Register(owner, "user:owner1")

	subscriber := newTestClient("owner1", 4)
	hub.Register(subscriber, "url:abc123")

	hub.PublishClickEvent(analytics.ClickEvent{
		ShortCode:  "abc123",
		OwnerID:    "owner1",
		Country:    "PE",
		DeviceType: "mobile",
		Browser:    "Chrome",
		Timestamp:  time.Now().UTC(),
	})

	for _, tc := range []struct {
		name string
		c    *Client
	}{{"room de usuario", owner}, {"room de url", subscriber}} {
		msgs := drain(tc.c)
		if len(msgs) != 1 {
			t.Fatalf("%s: esperaba 1 mensaje, hubo %d", tc.name, len(msgs))
		}
		if msgs[0].Type != MsgTypeClickEvent {
			t.Fatalf("%s: tipo inesperado %q", tc.name, msgs[0].Type)
		}
		var p ClickEventPayload
		if err := json.Unmarshal(msgs[0].Payload, &p); err != nil {
			t.Fatalf("%s: payload no deserializa: %v", tc.name, err)
		}
		if p.ShortCode != "abc123" || p.Country != "PE" {
			t.Fatalf("%s: payload inesperado %+v", tc.name, p)
		}
	}
}

// Criterio: un cliente NO unido a un room de URL no recibe sus eventos.
func TestBroadcast_NonMemberDoesNotReceive(t *testing.T) {
	hub := NewHub()

	outsider := newTestClient("intruder", 4)
	hub.Register(outsider, "user:intruder") // solo en su propio room

	hub.Broadcast("url:abc123", newMessage(MsgTypeClickEvent, ClickEventPayload{ShortCode: "abc123"}))

	if msgs := drain(outsider); len(msgs) != 0 {
		t.Fatalf("un no-miembro no debe recibir eventos del room, recibió %d", len(msgs))
	}
}

// Criterio: tras Unregister, publicar no entrega al cliente ni hace panic.
func TestUnregister_NoDeliveryNoPanic(t *testing.T) {
	hub := NewHub()

	c := newTestClient("owner1", 4)
	hub.Register(c, "user:owner1")
	hub.Unregister(c, "user:owner1")

	hub.PublishClickEvent(analytics.ClickEvent{ShortCode: "x", OwnerID: "owner1", Timestamp: time.Now()})

	if msgs := drain(c); len(msgs) != 0 {
		t.Fatalf("tras desregistrar no debe recibir eventos, recibió %d", len(msgs))
	}
	// El room vacío debe haberse eliminado del hub.
	hub.mu.RLock()
	_, exists := hub.rooms["user:owner1"]
	hub.mu.RUnlock()
	if exists {
		t.Fatal("un room vacío debe eliminarse del hub")
	}
}

// Criterio: buffer lleno => Broadcast descarta el mensaje nuevo sin bloquear ni
// hacer panic. Con 1000 eventos seguidos, no se bloquea y solo se conservan los
// que cupieron en el buffer.
func TestBroadcast_FullBufferDropsWithoutBlocking(t *testing.T) {
	hub := NewHub()

	const buffer = 8
	c := newTestClient("owner1", buffer)
	hub.Register(c, "user:owner1")

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			hub.PublishClickEvent(analytics.ClickEvent{ShortCode: "x", OwnerID: "owner1", Timestamp: time.Now()})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Broadcast se bloqueó con el buffer lleno: debería descartar, no esperar")
	}

	if got := len(drain(c)); got != buffer {
		t.Fatalf("esperaba exactamente %d mensajes retenidos (buffer), hubo %d", buffer, got)
	}
}
