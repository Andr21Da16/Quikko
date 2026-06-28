package shortener

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	// qrSize es el lado en píxeles del PNG generado (fijo, no configurable por ahora).
	qrSize = 256
	// qrDataPrefix hace el valor usable directo en un <img src="..."> sin procesar.
	qrDataPrefix = "data:image/png;base64,"
)

// GenerateQRBase64 genera el PNG de un código QR que codifica `content` (la URL
// corta completa) y lo devuelve como data URI base64 listo para el frontend:
// "data:image/png;base64,iVBORw0KG...". Es barato (microsegundos), por eso se
// genera on-demand en cada respuesta y no se persiste ni se cachea.
func GenerateQRBase64(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, qrSize)
	if err != nil {
		return "", fmt.Errorf("shortener: no se pudo generar el QR: %w", err)
	}
	return qrDataPrefix + base64.StdEncoding.EncodeToString(png), nil
}
