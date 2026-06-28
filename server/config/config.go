// Package config carga y valida toda la configuración del servicio desde
// variables de entorno. Es el ÚNICO lugar del proyecto que lee os.Getenv;
// el resto del código recibe un *Config inyectado.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	Env     string
	BaseURL string // base pública para construir el shortUrl (ej. https://sho.rt)

	// FrontendURL: base del frontend Next.js. La usa el redirect para mandar
	// los errores de negocio de GET /:code (link inexistente/inactivo) a páginas con marca
	// del frontend en vez de devolver JSON crudo a un navegador humano. Distinta de BaseURL
	// (que es el dominio del backend/short-links).
	FrontendURL string

	// CORS: orígenes permitidos para el frontend. Lista explícita,
	// NUNCA "*": si en el futuro se habilitan credentials/cookies, un wildcard con
	// credentials es rechazado por los navegadores y abriría un agujero de seguridad.
	// Hoy el auth va por header (no cookies), pero se deja la lista explícita lista
	// para producción.
	AllowedOrigins []string

	// Mongo
	MongoURI    string
	MongoDBName string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// InfluxDB
	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string

	// JWT
	JWTSecret          string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration

	// GeoIP: resolución de país vía servicio externo (ipapi.co) con cache en
	// Redis, en vez del .mmdb local de MaxMind (que requería registro). Timeout explícito
	// de cada llamada HTTP externa para no colgar la goroutine async del redirect.
	GeoIPHTTPTimeout time.Duration

	// Rate limiting
	RateLimitRedirectPerSec int
	RateLimitCreatePerMin   int
	AuthRateLimitPerMin     int // registro (por IP)
	LoginRateLimitPerMin    int // login (por IP), contador propio: el endpoint más sensible a fuerza bruta (Agent 25)

	// Planes
	FreePlanMaxActiveURLs int // tope de URLs activas simultáneas para el plan Free
}

// Load lee el archivo .env (si existe) y construye un *Config validado.
// Devuelve error si falta alguna variable obligatoria o si un valor tipado
// (int/duration) es inválido.
func Load() (*Config, error) {
	// En desarrollo cargamos .env; en producción las vars vienen del entorno.
	// Un .env ausente no es error: las variables pueden estar ya en el entorno.
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		MongoURI:    os.Getenv("MONGO_URI"),
		MongoDBName: getEnv("MONGO_DB_NAME", "url_shortener"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),

		InfluxURL:    getEnv("INFLUX_URL", "http://localhost:8086"),
		InfluxToken:  os.Getenv("INFLUX_TOKEN"),
		InfluxOrg:    getEnv("INFLUX_ORG", "url-shortener"),
		InfluxBucket: getEnv("INFLUX_BUCKET", "clicks"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		AllowedOrigins: getEnvList("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
	}

	var err error

	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.JWTAccessTokenTTL, err = getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return nil, err
	}
	if cfg.JWTRefreshTokenTTL, err = getEnvDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour); err != nil {
		return nil, err
	}
	if cfg.RateLimitRedirectPerSec, err = getEnvInt("RATE_LIMIT_REDIRECT_PER_SEC", 100); err != nil {
		return nil, err
	}
	if cfg.RateLimitCreatePerMin, err = getEnvInt("RATE_LIMIT_CREATE_PER_MIN", 20); err != nil {
		return nil, err
	}
	if cfg.AuthRateLimitPerMin, err = getEnvInt("AUTH_RATE_LIMIT_PER_MIN", 5); err != nil {
		return nil, err
	}
	if cfg.LoginRateLimitPerMin, err = getEnvInt("RATE_LIMIT_LOGIN_PER_MIN", 5); err != nil {
		return nil, err
	}
	if cfg.FreePlanMaxActiveURLs, err = getEnvInt("FREE_PLAN_MAX_ACTIVE_URLS", 5); err != nil {
		return nil, err
	}

	var geoIPTimeoutSecs int
	if geoIPTimeoutSecs, err = getEnvInt("GEOIP_HTTP_TIMEOUT_SECONDS", 8); err != nil {
		return nil, err
	}
	cfg.GeoIPHTTPTimeout = time.Duration(geoIPTimeoutSecs) * time.Second

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate verifica que las variables obligatorias no estén vacías.
func (c *Config) validate() error {
	required := map[string]string{
		"MONGO_URI":  c.MongoURI,
		"JWT_SECRET": c.JWTSecret,
	}
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("config: variable de entorno obligatoria %q no está definida", name)
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvList parsea una variable separada por comas en un slice, recortando espacios
// y descartando entradas vacías. Si la variable no está definida, usa el fallback.
func getEnvList(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: %q debe ser un entero válido, se obtuvo %q: %w", key, v, err)
	}
	return n, nil
}

// getEnvDuration parsea valores tipo "15m", "7d", "30s". Como time.ParseDuration
// no soporta "d" (días), se traduce manualmente a horas.
func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	if len(v) > 1 && v[len(v)-1] == 'd' {
		days, err := strconv.Atoi(v[:len(v)-1])
		if err != nil {
			return 0, fmt.Errorf("config: %q tiene formato de días inválido %q: %w", key, v, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %q debe ser una duración válida (ej. 15m, 7d), se obtuvo %q: %w", key, v, err)
	}
	return d, nil
}
