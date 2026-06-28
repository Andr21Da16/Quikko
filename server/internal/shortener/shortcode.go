package shortener

import (
	"crypto/rand"
	"fmt"
)

// base62Alphabet es el conjunto de caracteres usado para los códigos cortos.
const base62Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// defaultCodeLength es la longitud por defecto de un código autogenerado.
// 6 caracteres base62 ≈ 56.800 millones de combinaciones.
const defaultCodeLength = 7

// maxRetries es el número de reintentos ante colisión antes de rendirse.
const maxRetries = 5

// GenerateCode produce un código corto base62 criptográficamente aleatorio de
// la longitud indicada. Usa crypto/rand (no math/rand) para evitar secuencias
// predecibles. Si length <= 0 se usa defaultCodeLength.
func GenerateCode(length int) (string, error) {
	if length <= 0 {
		length = defaultCodeLength
	}

	// Leemos `length` bytes aleatorios y mapeamos cada uno al alfabeto. Para
	// evitar sesgo de módulo, descartamos los bytes que caen en la "cola" que no
	// es múltiplo exacto del tamaño del alfabeto, releyendo según haga falta.
	n := byte(len(base62Alphabet))
	maxUnbiased := byte(256 - (256 % int(n))) // mayor múltiplo de n que cabe en un byte

	out := make([]byte, length)
	filled := 0
	buf := make([]byte, length)

	for filled < length {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("shortener: error leyendo entropía: %w", err)
		}
		for _, b := range buf {
			if b >= maxUnbiased {
				continue // descartar para no sesgar la distribución
			}
			out[filled] = base62Alphabet[b%n]
			filled++
			if filled == length {
				break
			}
		}
	}

	return string(out), nil
}
