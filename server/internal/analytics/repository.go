package analytics

import "context"

// ClickRepository es la persistencia de clics en InfluxDB (serie temporal).
type ClickRepository interface {
	// Write encola el evento vía el Write API asíncrono de InfluxDB. No bloquea:
	// devuelve apenas el punto queda en el buffer interno del cliente.
	Write(ctx context.Context, event ClickEvent) error
	// QueryStats agrega los clics del rango por país/dispositivo/navegador y arma
	// la serie temporal. shortCode vacío agrega TODAS las URLs de ownerID (overview).
	QueryStats(ctx context.Context, shortCode, ownerID string, rango TimeRange) (*ClickStats, error)
}
