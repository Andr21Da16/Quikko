package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"urlshortener/config"
	"urlshortener/internal/analytics"
	"urlshortener/internal/auth"
	"urlshortener/internal/platform/influx"
	appmw "urlshortener/internal/platform/middleware"
	mongodb "urlshortener/internal/platform/mongo"
	redisclient "urlshortener/internal/platform/redis"
	"urlshortener/internal/realtime"
	"urlshortener/internal/redirect"
	"urlshortener/internal/shortener"
)

type Dependencies struct {
	Config *config.Config
	Mongo  *mongo.Database
	Redis  *redis.Client
	Influx influxdb2.Client
}

// swaggerUIHTML es la página de Swagger UI servida en GET /docs. Carga los assets
// desde un CDN y apunta al contrato estático en /docs/openapi.yaml.
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <title>Quikko API — Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/docs/openapi.yaml',
      dom_id: '#swagger-ui',
    });
  </script>
</body>
</html>`

// customValidator integra go-playground/validator con el binding de Echo, de
// modo que los handlers puedan llamar c.Validate(dto) sobre sus DTOs.
type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("no se pudo cargar la configuración", "err", err)
		os.Exit(1)
	}

	// --- Conexiones a infraestructura: fail-fast en boot ---
	mongoDB, err := mongodb.Connect(cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		slog.Error("fallo al conectar a MongoDB", "err", err)
		os.Exit(1)
	}

	redisClient, err := redisclient.Connect(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Error("fallo al conectar a Redis", "err", err)
		os.Exit(1)
	}

	influxClient, err := influx.Connect(cfg.InfluxURL, cfg.InfluxToken, cfg.InfluxOrg)
	if err != nil {
		slog.Error("fallo al conectar a InfluxDB", "err", err)
		os.Exit(1)
	}

	deps := &Dependencies{
		Config: cfg,
		Mongo:  mongoDB,
		Redis:  redisClient,
		Influx: influxClient,
	}

	// --- Echo: middlewares globales + error handler personalizado ---
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = appmw.ErrorHandler
	e.Validator = &customValidator{validator: validator.New()}

	e.Use(echomw.Recover())
	e.Use(echomw.Logger())

	// Headers de seguridad HTTP. Se aplican a TODAS las respuestas salvo /docs.
	//   - nosniff: el navegador no adivina el Content-Type.
	//   - X-Frame-Options: DENY + CSP frame-ancestors 'none': el dashboard nunca se embebe
	//     en iframes de terceros (anti-clickjacking).
	//   - Referrer-Policy: limita la fuga del referrer hacia destinos cross-origin.
	//   - CSP default-src 'none': defensa en profundidad para respuestas JSON (que no se
	//     renderizan como documento). NO se agrega HSTS: el proyecto corre en HTTP local;
	//     HSTS es tarea del deploy real con HTTPS (fuera de alcance).
	// Estos headers NO interfieren con el redirect (GET /:code): un 302 lleva headers +
	// body vacío y el navegador sigue el Location igual; Referrer-Policy incluso es deseable
	// ahí. La ÚNICA ruta excluida es /docs (Swagger UI), que carga CSS/JS de un CDN externo
	// e inline scripts incompatibles con una CSP estricta default-src 'none'.
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/docs")
		},
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'none'; frame-ancestors 'none'",
	}))

	// Límite de tamaño de body. Los payloads legítimos del proyecto son
	// diminutos (URLs, emails, passwords); 1MB es holgado. Un body mayor se rechaza con
	// 413 (vía el HTTPErrorHandler global), no con 500 ni colgando el proceso.
	e.Use(echomw.BodyLimit("1M"))

	// CORS configurable por entorno. AllowOrigins es una lista
	// explícita (ALLOWED_ORIGINS), NUNCA "*": un origen no listado no recibe los
	// headers que permitirían al navegador hacer la petición. Si algún día se
	// habilitan cookies/credentials, mantener la lista explícita es obligatorio.
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
	}))

	// Healthcheck básico de infraestructura (útil para readiness probes).
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// --- Documentación de la API ---
	// Sirve el contrato OpenAPI estático y una Swagger UI (desde CDN, sin dependencias
	// Go nuevas ni anotaciones swaggo). El YAML se resuelve relativo al cwd, que en
	// desarrollo es server/ (go run cmd/api/main.go). Rutas estáticas: Echo las
	// prioriza, así que no colisionan con GET /:code de redirect.
	e.File("/docs/openapi.yaml", "docs/openapi.yaml")
	e.GET("/docs", func(c echo.Context) error {
		return c.HTML(http.StatusOK, swaggerUIHTML)
	})

	// Cada dominio expone su propio RegisterRoutes(api *echo.Group, deps *Dependencies)
	// y se invoca aquí. Por ejemplo:
	//
	//   auth.RegisterRoutes(api, deps)
	//   shortener.RegisterRoutes(api, deps)
	//
	// Hasta que existan, el grupo queda listo y vacío.
	api := e.Group("/api/v1")
	if err := registerRoutes(e, api, deps); err != nil {
		slog.Error("fallo registrando rutas de los dominios", "err", err)
		os.Exit(1)
	}

	// --- Arranque + graceful shutdown ---
	go func() {
		addr := ":" + cfg.Port
		slog.Info("servidor escuchando", "addr", addr, "env", cfg.Env)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("el servidor se detuvo inesperadamente", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("señal de apagado recibida, cerrando ordenadamente...")

	shutdown(e, deps)
	slog.Info("apagado completo")
}

func registerRoutes(e *echo.Echo, api *echo.Group, deps *Dependencies) error {
	// Adaptador de las URLs del usuario que auth necesita para el resumen y el
	// borrado en cascada de cuenta. Se construye con los repos públicos de
	// shortener (idempotentes: la creación de índices es no-op si ya existen, mismo
	// patrón que redirect). auth no puede importar shortener (crearía un ciclo), así
	// que shortener provee el adaptador y aquí, en el wiring, se inyecta.
	urlRepo, err := shortener.NewURLRepository(deps.Mongo)
	if err != nil {
		return err
	}
	accountStore := shortener.NewAccountURLStore(urlRepo, shortener.NewCacheRepository(deps.Redis))

	if err := auth.RegisterRoutes(api, deps.Mongo, deps.Redis, deps.Config, accountStore); err != nil {
		return err
	}
	if err := shortener.RegisterRoutes(api, deps.Mongo, deps.Redis, deps.Config); err != nil {
		return err
	}
	// Realtime: el Hub se crea primero porque es a la vez el
	// EventPublisher que Analytics inyecta para republicar cada clic por WebSocket.
	hub := realtime.NewHub()

	// Analytics se construye ANTES que redirect: su service es el
	// ClickRecorder real que se inyecta en redirect, reemplazando al
	// NoopClickRecorder. Recibe el Hub como EventPublisher para el push en tiempo real.
	recorder, err := analytics.RegisterRoutes(api, deps.Mongo, deps.Influx, deps.Config, hub)
	if err != nil {
		return err
	}
	// Redirect monta GET /:code en la raíz (no en /api/v1) y emite cada clic al
	// ClickRecorder inyectado (Analytics).
	if err := redirect.RegisterRoutes(e, deps.Mongo, deps.Redis, deps.Config, recorder); err != nil {
		return err
	}
	// Realtime monta GET /ws en la raíz (token por query param, validado en el handler).
	if err := realtime.RegisterRoutes(e, hub, deps.Mongo, deps.Config); err != nil {
		return err
	}
	return nil
}

// shutdown cierra el servidor HTTP y luego las conexiones de infraestructura.
func shutdown(e *echo.Echo, deps *Dependencies) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error apagando el servidor HTTP", "err", err)
	}

	if err := deps.Mongo.Client().Disconnect(ctx); err != nil {
		slog.Error("error desconectando MongoDB", "err", err)
	}
	if err := deps.Redis.Close(); err != nil {
		slog.Error("error cerrando Redis", "err", err)
	}
	deps.Influx.Close()
}
