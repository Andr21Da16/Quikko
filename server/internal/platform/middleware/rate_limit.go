package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"urlshortener/pkg/response"
)

// RateLimit es un middleware genérico de fixed-window sobre Redis.
//
//   - keyPrefix:    namespace de la clave (ej. "rl:redirect").
//   - maxRequests:  máximo de requests permitidos por ventana.
//   - window:       duración de la ventana (ej. 1*time.Second, 1*time.Minute).
//   - keyExtractor: deriva el identificador del cliente (IP, userID, shortCode...).
//
// Usa INCR + EXPIRE: el primer request de la ventana fija el TTL; a partir del
// request N+1 dentro de la ventana, corta con 429.
func RateLimit(
	rdb *redis.Client,
	keyPrefix string,
	maxRequests int,
	window time.Duration,
	keyExtractor func(echo.Context) string,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := fmt.Sprintf("%s:%s", keyPrefix, keyExtractor(c))

			ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
			defer cancel()

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				// Fail-open: si Redis no responde, no bloqueamos tráfico legítimo;
				// solo lo registramos para observabilidad.
				slog.Error("rate limit: error consultando Redis, permitiendo request", "key", key, "err", err)
				return next(c)
			}

			// En el primer request de la ventana, fijamos el TTL.
			if count == 1 {
				if err := rdb.Expire(ctx, key, window).Err(); err != nil {
					slog.Error("rate limit: no se pudo fijar EXPIRE", "key", key, "err", err)
				}
			}

			if count > int64(maxRequests) {
				return response.Fail(c, 429, "RATE_LIMIT_EXCEEDED", "Demasiadas solicitudes. Inténtalo de nuevo en unos momentos.")
			}

			return next(c)
		}
	}
}

// ClientIPExtractor es un keyExtractor común basado en la IP del cliente.
func ClientIPExtractor(c echo.Context) string {
	return c.RealIP()
}
