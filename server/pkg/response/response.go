package response

import (
	"sync"

	"github.com/labstack/echo/v4"
)

// Envelope es la estructura única de toda respuesta HTTP.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *APIError   `json:"error"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError describe un error de negocio expuesto al cliente.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta acompaña a las respuestas paginadas.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// Success responde un payload simple: success=true, error=null.
func Success(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, Envelope{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// SuccessPaginated responde una colección con su bloque meta de paginación.
func SuccessPaginated(c echo.Context, status int, data interface{}, meta Meta) error {
	return c.JSON(status, Envelope{
		Success: true,
		Data:    data,
		Error:   nil,
		Meta:    &meta,
	})
}

// Fail responde un error: success=false, data=null, error poblado.
func Fail(c echo.Context, status int, code, message string) error {
	return c.JSON(status, Envelope{
		Success: false,
		Data:    nil,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

type errorMapping struct {
	status  int
	code    string
	message string
}

var (
	mu       sync.RWMutex
	registry = make(map[error]errorMapping)
)

// RegisterErrorMapping asocia un error de dominio centinela con su respuesta HTTP.
// Es idempotente y seguro para concurrencia.
func RegisterErrorMapping(err error, status int, code, message string) {
	mu.Lock()
	defer mu.Unlock()
	registry[err] = errorMapping{status: status, code: code, message: message}
}

// MapDomainError traduce un error a su (status, code, message). Si el error no
// está registrado (o sus errores envueltos con %w tampoco), devuelve un
// INTERNAL_ERROR 500 genérico que no filtra detalles internos.
func MapDomainError(err error) (status int, code string, message string) {
	mu.RLock()
	defer mu.RUnlock()

	// Recorremos la cadena de errores envueltos (errors.Unwrap) para soportar
	// fmt.Errorf("...: %w", ErrDominio).
	for e := err; e != nil; e = unwrap(e) {
		if m, ok := registry[e]; ok {
			return m.status, m.code, m.message
		}
	}
	return 500, "INTERNAL_ERROR", "Ocurrió un error interno. Inténtalo de nuevo más tarde."
}

// unwrap es un wrapper local sobre errors.Unwrap para evitar el import en el
// hot-path de arriba y mantener el bucle legible.
func unwrap(err error) error {
	u, ok := err.(interface{ Unwrap() error })
	if !ok {
		return nil
	}
	return u.Unwrap()
}
