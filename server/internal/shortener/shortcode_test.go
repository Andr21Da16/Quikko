package shortener

import (
	"regexp"
	"testing"
)

var base62Re = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func TestGenerateCode_LengthAndAlphabet(t *testing.T) {
	for _, length := range []int{6, 7, 10} {
		code, err := GenerateCode(length)
		if err != nil {
			t.Fatalf("GenerateCode(%d) error: %v", length, err)
		}
		if len(code) != length {
			t.Fatalf("esperaba longitud %d, obtuvo %d (%q)", length, len(code), code)
		}
		if !base62Re.MatchString(code) {
			t.Fatalf("código fuera del alfabeto base62: %q", code)
		}
	}
}

func TestGenerateCode_DefaultLength(t *testing.T) {
	code, err := GenerateCode(0)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(code) != defaultCodeLength {
		t.Fatalf("esperaba longitud por defecto %d, obtuvo %d", defaultCodeLength, len(code))
	}
}

// TestGenerateCode_NoCollisionsIn1000 cubre el criterio de aceptación: generar
// 1000 códigos no debe producir colisiones (a 7 chars base62 la probabilidad es
// ínfima; si fallara indicaría un bug en la fuente de entropía).
func TestGenerateCode_NoCollisionsIn1000(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		code, err := GenerateCode(defaultCodeLength)
		if err != nil {
			t.Fatalf("error en iteración %d: %v", i, err)
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("colisión detectada en iteración %d: %q", i, code)
		}
		seen[code] = struct{}{}
	}
	if len(seen) != n {
		t.Fatalf("esperaba %d códigos únicos, obtuvo %d", n, len(seen))
	}
}
