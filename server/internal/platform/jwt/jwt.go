// Package jwt firma y valida los tokens del proyecto. Lo usan el dominio auth
// (al emitir tokens) y el auth_middleware (al validarlos).
package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Tipos de token soportados, embebidos en el claim "type".
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

// Errores tipados para que el middleware distinga expiración de invalidez.
var (
	ErrTokenExpired = errors.New("jwt: token expirado")
	ErrTokenInvalid = errors.New("jwt: token inválido")
)

// Claims es el payload firmado. Embebe RegisteredClaims para exp/iat estándar.
type Claims struct {
	UserID string `json:"userID"`
	Type   string `json:"type"`
	jwtlib.RegisteredClaims
}

// GenerateAccessToken firma un token de acceso de corta duración.
func GenerateAccessToken(userID, secret string, ttl time.Duration) (string, error) {
	return generate(userID, TypeAccess, secret, ttl)
}

// GenerateRefreshToken firma un token de refresco de larga duración.
func GenerateRefreshToken(userID, secret string, ttl time.Duration) (string, error) {
	return generate(userID, TypeRefresh, secret, ttl)
}

func generate(userID, tokenType, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("jwt: no se pudo firmar el token: %w", err)
	}
	return signed, nil
}

// ValidateToken verifica firma y expiración. Devuelve ErrTokenExpired o
// ErrTokenInvalid (envueltos) para que el llamador mapee el código HTTP correcto.
func ValidateToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwtlib.ParseWithClaims(tokenString, claims, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: método de firma inesperado %v", ErrTokenInvalid, t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
