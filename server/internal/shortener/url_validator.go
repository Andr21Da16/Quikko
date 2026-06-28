package shortener

import (
	"net"
	"net/url"
	"strings"
)

type invalidURLError struct {
	reason string
}

func (e *invalidURLError) Error() string { return e.reason }
func (e *invalidURLError) Unwrap() error { return ErrInvalidURL }

func invalidURL(reason string) error {
	return &invalidURLError{reason: reason}
}

// ValidateOriginalURL refuerza la validación del DTO (`validate:"url"`) con chequeos
// que las tags no pueden expresar: esquema permitido, hosts locales/privados y el
// dominio propio del servicio (para prevenir loops de redirección infinitos).
//
// ownDomain es la BASE_URL configurada (ej. "https://quikko.io"); se compara solo
// su host. Devuelve un error que envuelve ErrInvalidURL con el motivo específico.
func ValidateOriginalURL(rawURL, ownDomain string) error {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return invalidURL("La URL no tiene un formato válido.")
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return invalidURL("Solo se permiten URLs con esquema http o https.")
	}

	host := u.Hostname()
	if host == "" {
		return invalidURL("La URL debe incluir un host válido.")
	}
	lowerHost := strings.ToLower(host)

	if lowerHost == "localhost" {
		return invalidURL("No se permiten URLs que apunten a localhost.")
	}

	// Si el host es una IP literal, rechazar loopback, privadas, no especificadas
	// (0.0.0.0) y link-local (169.254.0.0/16). net.IP.IsPrivate cubre 10/8, 172.16/12
	// y 192.168/16 sin listas manuales.
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			return invalidURL("No se permiten URLs que apunten a direcciones IP privadas o locales.")
		}
	}

	// Evitar acortar una URL del propio servicio (loop de redirección).
	if own := ownDomainHost(ownDomain); own != "" && lowerHost == own {
		return invalidURL("No puedes acortar una URL de este mismo servicio.")
	}

	return nil
}

// ownDomainHost extrae el host (sin puerto) de la BASE_URL configurada. Devuelve ""
// si no se puede parsear, en cuyo caso simplemente no se aplica el chequeo de mismo
// dominio (degradación segura: el resto de validaciones sigue vigente).
func ownDomainHost(ownDomain string) string {
	if ownDomain == "" {
		return ""
	}
	u, err := url.Parse(strings.TrimSpace(ownDomain))
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
