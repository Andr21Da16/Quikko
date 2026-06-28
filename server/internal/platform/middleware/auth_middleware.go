// Package middleware contiene los middlewares transversales de Echo
// (autenticación JWT, rate limiting y el error handler global).
package middleware

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"

	"urlshortener/internal/platform/jwt"
	"urlshortener/pkg/response"
)

// ContextUserIDKey es la clave bajo la cual se inyecta el userID autenticado.
const ContextUserIDKey = "userID"

// Auth devuelve un middleware que exige un header "Authorization: Bearer <token>"
// válido. Si lo es, inyecta el userID en el contexto; si no, corta con 401.
func Auth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return response.Fail(c, 401, "AUTH_TOKEN_INVALID", "Falta el header Authorization.")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				return response.Fail(c, 401, "AUTH_TOKEN_INVALID", "El header Authorization debe tener formato 'Bearer <token>'.")
			}

			claims, err := jwt.ValidateToken(parts[1], secret)
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					return response.Fail(c, 401, "AUTH_TOKEN_EXPIRED", "El token ha expirado. Inicia sesión nuevamente.")
				}
				return response.Fail(c, 401, "AUTH_TOKEN_INVALID", "El token es inválido.")
			}

			c.Set(ContextUserIDKey, claims.UserID)
			return next(c)
		}
	}
}
