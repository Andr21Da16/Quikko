package shortener

import (
	"encoding/base64"
	"strings"
	"testing"
)

// pngMagic son los primeros bytes de todo archivo PNG.
var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

func TestGenerateQRBase64_PrefixAndPNG(t *testing.T) {
	const content = "https://sho.rt/abc123"

	got, err := GenerateQRBase64(content)
	if err != nil {
		t.Fatalf("generación de QR falló: %v", err)
	}
	if got == "" {
		t.Fatal("el QR no debe estar vacío")
	}
	if !strings.HasPrefix(got, qrDataPrefix) {
		t.Fatalf("el QR debe empezar con %q, fue %q...", qrDataPrefix, got[:min(len(got), 32)])
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(got, qrDataPrefix))
	if err != nil {
		t.Fatalf("la parte base64 no decodifica: %v", err)
	}
	if len(raw) < len(pngMagic) || string(raw[:len(pngMagic)]) != string(pngMagic) {
		t.Fatal("el contenido decodificado no es un PNG válido")
	}
}

// Determinista: el mismo contenido produce el mismo QR; contenidos distintos, QRs
// distintos (garantiza que se codifica el contenido dado, no algo fijo).
func TestGenerateQRBase64_DeterministicAndContentSensitive(t *testing.T) {
	a1, _ := GenerateQRBase64("https://sho.rt/aaa")
	a2, _ := GenerateQRBase64("https://sho.rt/aaa")
	b, _ := GenerateQRBase64("https://sho.rt/bbb")

	if a1 != a2 {
		t.Fatal("el mismo contenido debe producir el mismo QR")
	}
	if a1 == b {
		t.Fatal("contenidos distintos deben producir QRs distintos")
	}
}

// toURLResponse debe poblar el QR codificando exactamente el shortUrl devuelto.
func TestToURLResponse_QRMatchesShortURL(t *testing.T) {
	u := &ShortURL{ID: "id1", ShortCode: "abc123", OriginalURL: "https://go.dev", IsActive: true}
	resp := toURLResponse(u, "https://sho.rt/")

	if resp.ShortURL != "https://sho.rt/abc123" {
		t.Fatalf("shortUrl inesperado: %q", resp.ShortURL)
	}
	expected, _ := GenerateQRBase64(resp.ShortURL)
	if resp.QRCodeBase64 != expected {
		t.Fatal("el QRCodeBase64 debe codificar exactamente el shortUrl de la respuesta")
	}
}
