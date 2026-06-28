package redirect

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// countryUnknown es el valor por defecto cuando no se puede resolver el país.
// Nunca se devuelve un error al flujo de redirección por un fallo de GeoIP.
const countryUnknown = "unknown"

// geoIPCacheKeyPrefix namespacea las entradas IP→país en Redis (clave geoip:{ip}).
const geoIPCacheKeyPrefix = "geoip:"

// geoIPCacheTTL: 48h. Se cachean también los "unknown" (fallos): así una IP problemática
// no dispara una llamada externa en cada visita dentro de la ventana.
const geoIPCacheTTL = 48 * time.Hour

// geoIPCacheOpTimeout acota cada operación de cache en Redis. Es corto a propósito: si
// Redis está lento/caído, preferimos saltar el cache y resolver, no colgar la goroutine.
const geoIPCacheOpTimeout = 2 * time.Second

// geoIPAPIBaseURL es la base del endpoint de texto plano de ipapi.co. Es var (no const)
// solo para poder apuntarla a un servidor de prueba en los tests; en runtime no se cambia.
var geoIPAPIBaseURL = "https://ipapi.co/"

// CountryResolver resuelve un código de país ISO a partir de una IP.
type CountryResolver interface {
	ResolveCountry(ip string) string
}

// noopResolver siempre devuelve "unknown". Se usa en los tests y como degradación segura.
type noopResolver struct{}

func (noopResolver) ResolveCountry(string) string { return countryUnknown }

// externalGeoIPResolver resuelve el país consultando ipapi.co, con cache en Redis.
//
// DESVIACIÓN ARQUITECTÓNICA DOCUMENTADA: el proyecto evita registrarse en
// MaxMind, así que se reemplaza el .mmdb local por un servicio HTTP externo (ipapi.co, sin
// API key). Esto introduce una llamada de red dentro de la goroutine async del redirect
// (RecordClickAsync) — NUNCA en el camino síncrono del 302. Se mitigan sus tres riesgos
// (latencia variable, rate limit del servicio gratuito, nuevo punto de fallo de red):
//   - cache en Redis (TTL 48h): solo la 1ª vez que se ve una IP se paga la llamada externa;
//   - timeout explícito en el http.Client: la goroutine nunca queda colgada indefinidamente;
//   - cualquier fallo (timeout, red, respuesta inesperada, rate limit) degrada a "unknown",
//     igual que el comportamiento previo cuando faltaba el .mmdb — nunca rompe el registro.
type externalGeoIPResolver struct {
	rdb    *redis.Client // cliente Redis compartido; si está caído, se salta el cache
	client *http.Client  // con Timeout explícito (config GEOIP_HTTP_TIMEOUT_SECONDS)
}

// NewCountryResolver construye el resolver de país basado en ipapi.co con cache en Redis.
// `rdb` es el cliente Redis compartido (el mismo de la cache de redirección); si Redis está
// caído en runtime, la resolución degrada a llamar siempre al servicio externo (sin cache),
// sin error fatal. `httpTimeout` acota cada llamada externa.
func NewCountryResolver(rdb *redis.Client, httpTimeout time.Duration) CountryResolver {
	return &externalGeoIPResolver{
		rdb:    rdb,
		client: &http.Client{Timeout: httpTimeout},
	}
}

func (r *externalGeoIPResolver) ResolveCountry(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return countryUnknown
	}
	// IPs privadas/loopback/no-públicas: nunca gastamos una llamada externa (ipapi.co
	// tampoco podría resolverlas). Mismo precedente de chequeo que shortener.ValidateOriginalURL.
	if !isPublicIP(parsed) {
		return countryUnknown
	}

	// 1) Cache: si Redis tiene la IP (incluido un "unknown" cacheado), se usa directo.
	if country, ok := r.getCached(ip); ok {
		return country
	}

	// 2) Servicio externo. Cualquier fallo → "unknown".
	country := r.fetchFromAPI(ip)

	// 3) Cachear el resultado (país real o "unknown") con TTL 48h. Si Redis falla, se
	//    loguea y se sigue: degradar, no caer.
	r.setCached(ip, country)

	return country
}

// isPublicIP descarta IPs que no tiene sentido (ni se puede) geolocalizar externamente.
func isPublicIP(ip net.IP) bool {
	return !(ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast())
}

// getCached devuelve el país cacheado y true si existe. Ante Redis ausente/caído o un
// miss, devuelve ("", false) y el caller consulta el servicio externo.
func (r *externalGeoIPResolver) getCached(ip string) (string, bool) {
	if r.rdb == nil {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), geoIPCacheOpTimeout)
	defer cancel()

	val, err := r.rdb.Get(ctx, geoIPCacheKeyPrefix+ip).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			// Error real de Redis (no un simple miss): degradamos a consultar el servicio.
			slog.Warn("geoip: error leyendo cache de país, se ignora el cache", "ip", ip, "err", err)
		}
		return "", false
	}
	return val, true
}

// setCached guarda IP→país en Redis con TTL de 48h. Un fallo (incluido Redis caído) se
// loguea y no rompe nada.
func (r *externalGeoIPResolver) setCached(ip, country string) {
	if r.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), geoIPCacheOpTimeout)
	defer cancel()

	if err := r.rdb.Set(ctx, geoIPCacheKeyPrefix+ip, country, geoIPCacheTTL).Err(); err != nil {
		slog.Warn("geoip: no se pudo cachear el país, se continúa", "ip", ip, "err", err)
	}
}

// fetchFromAPI consulta https://ipapi.co/{ip}/country/ (endpoint de TEXTO PLANO que
// devuelve solo el código ISO de país). Se eligió el de texto plano sobre el JSON completo
// porque solo se necesita el país: menos bytes y parsing trivial. Cualquier anomalía
// (status no-200, body inesperado, rate limit, error de red, timeout) → "unknown".
func (r *externalGeoIPResolver) fetchFromAPI(ip string) string {
	url := geoIPAPIBaseURL + ip + "/country/"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Warn("geoip: no se pudo construir la request externa, país 'unknown'", "ip", ip, "err", err)
		return countryUnknown
	}
	// ipapi.co rechaza con 403 las requests sin User-Agent.
	req.Header.Set("User-Agent", "quikko-url-shortener/1.0")
	req.Header.Set("Accept", "text/plain")

	resp, err := r.client.Do(req)
	if err != nil {
		// Incluye el caso de timeout (Client.Timeout): corta la espera y degrada.
		slog.Warn("geoip: fallo consultando ipapi.co, país 'unknown'", "ip", ip, "err", err)
		return countryUnknown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("geoip: ipapi.co respondió status no-200, país 'unknown'", "ip", ip, "status", resp.StatusCode)
		return countryUnknown
	}

	// El código ISO son 2 bytes; leemos acotado para no confiar en el tamaño del body.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		slog.Warn("geoip: error leyendo respuesta de ipapi.co, país 'unknown'", "ip", ip, "err", err)
		return countryUnknown
	}

	country := strings.TrimSpace(string(body))
	// ipapi.co devuelve mensajes de texto (ej. "Undefined", "RateLimited", "Sign up...")
	// ante IPs no resolubles o límites de uso; solo aceptamos un ISO-3166-1 alpha-2 válido.
	if !isValidCountryCode(country) {
		slog.Warn("geoip: respuesta inesperada de ipapi.co, país 'unknown'", "ip", ip, "resp", country)
		return countryUnknown
	}
	return country
}

// isValidCountryCode valida un código ISO-3166-1 alpha-2 (2 letras mayúsculas).
func isValidCountryCode(s string) bool {
	if len(s) != 2 {
		return false
	}
	for _, c := range s {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}
