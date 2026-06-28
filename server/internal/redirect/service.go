package redirect

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"urlshortener/internal/shortener"
)

type ClickEvent struct {
	ShortCode  string
	OwnerID    string
	Country    string
	DeviceType string
	Browser    string
	Timestamp  time.Time
}

type ClickRecorder interface {
	RecordClick(ctx context.Context, event ClickEvent) error
}

// NoopClickRecorder descarta los eventos.
type NoopClickRecorder struct{}

func (NoopClickRecorder) RecordClick(context.Context, ClickEvent) error { return nil }

// RedirectService resuelve códigos a URLs y dispara el registro async del clic.
type RedirectService interface {
	// Resolve devuelve la URL original para un código (Redis primero, fallback a
	// Mongo). Es la ruta caliente: debe ser lo más rápida posible.
	Resolve(ctx context.Context, code string) (originalURL string, err error)
	// RecordClickAsync construye y registra el clic en una goroutine propia; no
	// bloquea ni es esperada por el handler. Nunca afecta la respuesta al usuario.
	RecordClickAsync(code, ip, userAgent string)
}

type redirectService struct {
	repo     shortener.URLRepository
	cache    shortener.CacheRepository
	geo      CountryResolver
	recorder ClickRecorder
}

// NewRedirectService inyecta los repos de Agent 2 (reutilizados, no recreados a
// nivel de interfaz), el resolver de país y el ClickRecorder.
func NewRedirectService(repo shortener.URLRepository, cache shortener.CacheRepository, geo CountryResolver, recorder ClickRecorder) RedirectService {
	return &redirectService{repo: repo, cache: cache, geo: geo, recorder: recorder}
}

func (s *redirectService) Resolve(ctx context.Context, code string) (string, error) {
	// 1) Camino feliz: Redis tiene el valor -> devolver sin tocar Mongo.
	if url, err := s.cache.Get(ctx, code); err == nil {
		return url, nil
	} else if !errors.Is(err, shortener.ErrCacheMiss) {
		// Error real de Redis (no un simple miss): seguimos a Mongo igualmente,
		// pero lo registramos porque indica degradación de la cache.
		slog.Error("redirect: error leyendo cache, usando fallback a Mongo", "code", code, "err", err)
	}

	// 2) Fallback: fuente de verdad en Mongo.
	su, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return "", err // ErrURLNotFound se propaga
	}
	if !su.IsActive {
		return "", shortener.ErrURLInactive
	}

	// 3) Repoblar Redis para que el próximo hit sea rápido (no bloqueante si falla).
	if err := s.cache.Set(ctx, code, su.OriginalURL); err != nil {
		slog.Error("redirect: no se pudo repoblar la cache", "code", code, "err", err)
	}
	return su.OriginalURL, nil
}

func (s *redirectService) RecordClickAsync(code, ip, userAgent string) {
	go func() {
		// GeoIP primero: gestiona su PROPIO timeout (HTTP externo + cache Redis),
		// independiente del contexto de las escrituras a Mongo/Influx. Si lo hiciéramos
		// con el mismo ctx de 5s y la resolución externa tardara, ese deadline se
		// consumiría y el clic no se registraría — justo lo contrario a lo deseado
		// (el clic debe registrarse igual, con país "unknown", ante un geoip lento).
		country := s.geo.ResolveCountry(ip)
		device, browser := ParseUserAgent(userAgent)

		// Contexto propio para las operaciones de DB/analítica. Se crea DESPUÉS del
		// geoip para que su deadline empiece a contar aquí. El del request muere cuando
		// el handler responde el 302, por eso usamos context.Background().
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// El ownerID no está en la cache (solo guarda originalURL), así que lo
		// resolvemos aquí, en el camino async. Si falla, registramos el clic igual
		// con ownerID vacío en vez de perderlo.
		ownerID := ""
		if su, err := s.repo.FindByCode(ctx, code); err != nil {
			slog.Warn("redirect: no se pudo resolver ownerID para el clic", "code", code, "err", err)
		} else {
			ownerID = su.OwnerID
		}

		// Contador aproximado de clics en Mongo: mismo principio
		// fire-and-forget que el resto de esta goroutine — un fallo se loguea pero
		// nunca afecta la respuesta (el usuario ya fue redirigido).
		if err := s.repo.IncrementClicks(ctx, code); err != nil {
			slog.Error("redirect: no se pudo incrementar el contador de clics", "code", code, "err", err)
		}

		event := ClickEvent{
			ShortCode:  code,
			OwnerID:    ownerID,
			Country:    country,
			DeviceType: device,
			Browser:    browser,
			Timestamp:  time.Now().UTC(),
		}

		if err := s.recorder.RecordClick(ctx, event); err != nil {
			// Un fallo de analítica nunca afecta al usuario (ya fue redirigido).
			slog.Error("redirect: fallo registrando clic", "code", code, "err", err)
		}
	}()
}
