package middleware

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"urlshortener/pkg/response"
)

// ErrorHandler es el echo.HTTPErrorHandler global. Captura cualquier error que
// un handler devuelva sin haber escrito ya una respuesta y produce un envelope
// uniforme, sin filtrar detalles internos (stack traces van solo al log).
func ErrorHandler(err error, c echo.Context) {
	// Si la respuesta ya se envió (ej. el handler ya usó response.Fail), no la
	// pisamos; Echo solo nos llama de nuevo y duplicaría el body.
	if c.Response().Committed {
		return
	}

	// echo.HTTPError nos da un status semántico (404 de ruta no encontrada, etc.).
	if he, ok := err.(*echo.HTTPError); ok {
		code := codeForStatus(he.Code)
		message := http.StatusText(he.Code)
		if m, ok := he.Message.(string); ok && m != "" {
			message = m
		}
		_ = response.Fail(c, he.Code, code, message)
		return
	}

	// Cualquier error de dominio registrado se traduce; el resto cae a 500.
	status, code, message := response.MapDomainError(err)
	if status >= http.StatusInternalServerError {
		// Solo logueamos el detalle real de los 5xx; los 4xx son esperables.
		slog.Error("error no manejado", "path", c.Request().URL.Path, "method", c.Request().Method, "err", err)
	}
	_ = response.Fail(c, status, code, message)
}

// codeForStatus mapea los status HTTP genéricos de Echo a nuestros códigos de error.
func codeForStatus(status int) string {
	switch status {
	case http.StatusNotFound:
		return "URL_NOT_FOUND"
	case http.StatusUnauthorized:
		return "AUTH_TOKEN_INVALID"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusBadRequest:
		return "VALIDATION_ERROR"
	case http.StatusRequestEntityTooLarge:
		return "PAYLOAD_TOO_LARGE"
	default:
		return "INTERNAL_ERROR"
	}
}
