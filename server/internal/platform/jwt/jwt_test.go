package jwt

import (
	"errors"
	"testing"
	"time"
)

const testSecret = "secreto-de-prueba"

func TestValidateToken_Valido(t *testing.T) {
	token, err := GenerateAccessToken("user-123", testSecret, time.Minute)
	if err != nil {
		t.Fatalf("generación falló: %v", err)
	}
	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("validación falló: %v", err)
	}
	if claims.UserID != "user-123" || claims.Type != TypeAccess {
		t.Fatalf("claims inesperados: %+v", claims)
	}
}

func TestValidateToken_Expirado(t *testing.T) {
	token, err := GenerateAccessToken("user-123", testSecret, -time.Minute) // ya expirado
	if err != nil {
		t.Fatalf("generación falló: %v", err)
	}
	_, err = ValidateToken(token, testSecret)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("esperaba ErrTokenExpired, obtuvo %v", err)
	}
}

func TestValidateToken_Malformado(t *testing.T) {
	_, err := ValidateToken("esto.no.es.un.jwt", testSecret)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("esperaba ErrTokenInvalid, obtuvo %v", err)
	}
}

func TestValidateToken_FirmaIncorrecta(t *testing.T) {
	token, _ := GenerateAccessToken("user-123", testSecret, time.Minute)
	_, err := ValidateToken(token, "otro-secreto")
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("esperaba ErrTokenInvalid, obtuvo %v", err)
	}
}
