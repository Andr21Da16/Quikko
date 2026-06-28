package analytics

// StatsQueryParams son los query params de GET /analytics/stats. shortCode es
// opcional (su ausencia = vista overview de todas las URLs del usuario); range
// se valida también aquí en el binding, además del Service.
type StatsQueryParams struct {
	ShortCode *string `query:"shortCode"`
	Range     string  `query:"range" validate:"omitempty,oneof=24h 7d 30d"` // default "24h"
}

// StatsResponse es el payload de éxito: el rango efectivamente aplicado y el agregado.
type StatsResponse struct {
	Range string     `json:"range"`
	Stats ClickStats `json:"stats"`
}
