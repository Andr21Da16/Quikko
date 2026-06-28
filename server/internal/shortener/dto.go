package shortener

import (
	"log/slog"
	"strings"
	"time"
)

type CreateURLRequest struct {
	// max=2048: longitud máxima de URL ampliamente aceptada.
	OriginalURL string  `json:"originalUrl" validate:"required,url,max=2048"`
	CustomAlias *string `json:"customAlias,omitempty" validate:"omitempty,min=3,max=30,alphanum"`
}

type URLResponse struct {
	ID            string `json:"id"`
	ShortCode     string `json:"shortCode"`
	ShortURL      string `json:"shortUrl"` // base_url + shortCode
	OriginalURL   string `json:"originalUrl"`
	IsCustomAlias bool   `json:"isCustomAlias"`
	IsActive      bool   `json:"isActive"`
	CreatedAt     string `json:"createdAt"`
	// TotalClicks es el contador aproximado de clics del dominio (ver domain.go). Se
	// expone para que el frontend (Mis URLs / Analytics) muestre clics por URL sin una
	// query a InfluxDB por fila; la verdad fina sigue estando en InfluxDB.
	TotalClicks int64 `json:"totalClicks"`
	// QRCodeBase64 es el QR de shortUrl como data URI ("data:image/png;base64,...").
	// Se genera on-demand en cada respuesta; no se persiste.
	QRCodeBase64 string `json:"qrCodeBase64"`
}

type CheckAliasResponse struct {
	Alias     string `json:"alias"`
	Available bool   `json:"available"`
}

type ToggleActiveRequest struct {
	IsActive bool `json:"isActive"`
}

// toURLResponse traduce el modelo de dominio a su DTO de salida, construyendo el
// shortUrl absoluto a partir de la base pública configurada y su código QR.
//
// El QR codifica exactamente el mismo shortUrl que se devuelve en el DTO. Si la
// generación fallara (improbable para una URL normal), se loguea y se deja el
// campo vacío: nunca se rompe la respuesta de crear/listar por un fallo de QR.
func toURLResponse(u *ShortURL, baseURL string) URLResponse {
	shortURL := strings.TrimRight(baseURL, "/") + "/" + u.ShortCode

	qr, err := GenerateQRBase64(shortURL)
	if err != nil {
		slog.Error("no se pudo generar el QR de la URL", "shortCode", u.ShortCode, "err", err)
	}

	return URLResponse{
		ID:            u.ID,
		ShortCode:     u.ShortCode,
		ShortURL:      shortURL,
		OriginalURL:   u.OriginalURL,
		IsCustomAlias: u.IsCustomAlias,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
		TotalClicks:   u.TotalClicks,
		QRCodeBase64:  qr,
	}
}
