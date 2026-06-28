package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestEnvelopeContract(t *testing.T) {
	t.Run("éxito simple omite meta", func(t *testing.T) {
		env := Envelope{
			Success: true,
			Data:    map[string]string{"shortCode": "xYz12A", "originalUrl": "https://google.com"},
			Error:   nil,
		}
		got := marshal(t, env)
		want := `{"success":true,"data":{"originalUrl":"https://google.com","shortCode":"xYz12A"},"error":null}`
		if got != want {
			t.Fatalf("envelope de éxito no coincide.\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("éxito paginado incluye meta", func(t *testing.T) {
		env := Envelope{
			Success: true,
			Data:    []map[string]int{{"clicks": 42}},
			Error:   nil,
			Meta:    &Meta{Page: 1, Limit: 20, Total: 134, TotalPages: 7},
		}
		got := marshal(t, env)
		want := `{"success":true,"data":[{"clicks":42}],"error":null,"meta":{"page":1,"limit":20,"total":134,"totalPages":7}}`
		if got != want {
			t.Fatalf("envelope paginado no coincide.\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("error pone data null y omite meta", func(t *testing.T) {
		env := Envelope{
			Success: false,
			Data:    nil,
			Error:   &APIError{Code: "ALIAS_TAKEN", Message: "El alias 'mi-promo' ya está en uso. Elige otro."},
		}
		got := marshal(t, env)
		want := `{"success":false,"data":null,"error":{"code":"ALIAS_TAKEN","message":"El alias 'mi-promo' ya está en uso. Elige otro."}}`
		if got != want {
			t.Fatalf("envelope de error no coincide.\n got: %s\nwant: %s", got, want)
		}
	})
}

func TestMapDomainError(t *testing.T) {
	errAliasTaken := errors.New("alias taken")
	RegisterErrorMapping(errAliasTaken, 409, "ALIAS_TAKEN", "alias en uso")

	t.Run("error registrado se mapea", func(t *testing.T) {
		status, code, _ := MapDomainError(errAliasTaken)
		if status != 409 || code != "ALIAS_TAKEN" {
			t.Fatalf("esperaba 409/ALIAS_TAKEN, obtuvo %d/%s", status, code)
		}
	})

	t.Run("error envuelto con %w se mapea", func(t *testing.T) {
		status, code, _ := MapDomainError(fmt.Errorf("contexto del service: %w", errAliasTaken))
		if status != 409 || code != "ALIAS_TAKEN" {
			t.Fatalf("esperaba 409/ALIAS_TAKEN tras unwrap, obtuvo %d/%s", status, code)
		}
	})

	t.Run("error desconocido cae a INTERNAL_ERROR 500", func(t *testing.T) {
		status, code, _ := MapDomainError(errors.New("vaya"))
		if status != 500 || code != "INTERNAL_ERROR" {
			t.Fatalf("esperaba 500/INTERNAL_ERROR, obtuvo %d/%s", status, code)
		}
	})
}

func marshal(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal falló: %v", err)
	}
	return string(b)
}
