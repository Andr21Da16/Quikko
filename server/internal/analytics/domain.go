// Mapeo a InfluxDB: measurement "clicks"; tags indexados shortCode, ownerId,
// country, deviceType, browser; field count=1 (sumable); timestamp del clic.
package analytics

import (
	"errors"
	"time"
)

// ClickEvent es el modelo de dominio del clic dentro de analytics. Tiene el mismo
// shape que redirect.ClickEvent: RecordClick convierte uno en otro sin
// acoplar este paquete a la firma exacta del otro dominio en su capa interna.
type ClickEvent struct {
	ShortCode  string
	OwnerID    string
	Country    string
	DeviceType string
	Browser    string
	Timestamp  time.Time
}

// ClickStats es el agregado de métricas que consume el dashboard.
type ClickStats struct {
	TotalClicks     int64            `json:"totalClicks"`
	ClicksByCountry map[string]int64 `json:"clicksByCountry"`
	ClicksByDevice  map[string]int64 `json:"clicksByDevice"`
	ClicksByBrowser map[string]int64 `json:"clicksByBrowser"`
	ClicksOverTime  []TimeBucket     `json:"clicksOverTime"` // serie para gráfico de línea
}

// TimeBucket es un punto de la serie temporal (un bucket de aggregateWindow).
type TimeBucket struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
}

// TimeRange es el rango de consulta permitido. Se valida en el Service antes de
// tocar InfluxDB: nunca se interpola un valor arbitrario en la query Flux.
type TimeRange string

const (
	Range24h TimeRange = "24h"
	Range7d  TimeRange = "7d"
	Range30d TimeRange = "30d"
)

// Errores de dominio. El handler los traduce a códigos HTTP vía el mapeo central.
var (
	// ErrInvalidRange: el rango pedido no está en el set permitido.
	ErrInvalidRange = errors.New("rango de tiempo inválido")
	// ErrForbidden: la URL no pertenece al solicitante (o no existe). Mismo error
	// para ambos casos para no filtrar existencia (igual criterio que shortener).
	ErrForbidden = errors.New("sin acceso a las métricas de esta url")
	// ErrRangeNotAllowedForPlan: un usuario Free pidió un rango histórico (7d/30d)
	// reservado al plan Pro (Agent 7).
	ErrRangeNotAllowedForPlan = errors.New("rango histórico no disponible para el plan actual")
)

// ParseTimeRange normaliza y valida el rango. Un valor vacío usa el default 24h;
// cualquier valor fuera del set devuelve ErrInvalidRange.
func ParseTimeRange(s string) (TimeRange, error) {
	switch s {
	case "":
		return Range24h, nil
	case string(Range24h):
		return Range24h, nil
	case string(Range7d):
		return Range7d, nil
	case string(Range30d):
		return Range30d, nil
	default:
		return "", ErrInvalidRange
	}
}

// fluxStart devuelve la duración para la cláusula range(start:) de Flux (ej. -7d).
func (r TimeRange) fluxStart() string {
	return "-" + string(r)
}

// window devuelve la granularidad del aggregateWindow de la serie temporal:
// buckets horarios en 24h, diarios en 7d/30d.
func (r TimeRange) window() string {
	if r == Range24h {
		return "1h"
	}
	return "1d"
}
