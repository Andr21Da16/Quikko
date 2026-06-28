package shortener

import (
	"errors"
	"testing"
)

func TestValidateOriginalURL(t *testing.T) {
	const ownDomain = "https://quikko.io"

	rejected := []struct {
		name string
		url  string
	}{
		{"localhost", "http://localhost:8080/algo"},
		{"ip privada 192.168", "http://192.168.1.1/algo"},
		{"ip privada 10.x", "http://10.0.0.5/x"},
		{"loopback 127", "http://127.0.0.1/x"},
		{"unspecified 0.0.0.0", "http://0.0.0.0/x"},
		{"link-local 169.254", "http://169.254.1.1/x"},
		{"esquema ftp", "ftp://ejemplo.com"},
		{"esquema javascript", "javascript:alert(1)"},
		{"esquema data", "data:text/html,hola"},
		{"mismo dominio", "https://quikko.io/xyz"},
	}
	for _, tc := range rejected {
		t.Run("rechaza "+tc.name, func(t *testing.T) {
			err := ValidateOriginalURL(tc.url, ownDomain)
			if !errors.Is(err, ErrInvalidURL) {
				t.Fatalf("esperaba ErrInvalidURL para %q, obtuvo %v", tc.url, err)
			}
			if err.Error() == "" {
				t.Fatal("el error debe llevar un motivo específico legible")
			}
		})
	}

	accepted := []string{
		"https://google.com",
		"http://example.com/path?q=1",
		"https://sub.dominio.com/a/b/c",
	}
	for _, u := range accepted {
		t.Run("acepta "+u, func(t *testing.T) {
			if err := ValidateOriginalURL(u, ownDomain); err != nil {
				t.Fatalf("esperaba éxito para %q, obtuvo %v", u, err)
			}
		})
	}
}
