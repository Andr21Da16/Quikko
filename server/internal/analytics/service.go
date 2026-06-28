package analytics

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"urlshortener/internal/auth"
	"urlshortener/internal/redirect"
	"urlshortener/internal/shortener"
	"urlshortener/pkg/response"
)

func init() {
	response.RegisterErrorMapping(ErrInvalidRange, http.StatusBadRequest, "VALIDATION_ERROR",
		"El rango debe ser uno de: 24h, 7d, 30d.")
	response.RegisterErrorMapping(ErrForbidden, http.StatusForbidden, "FORBIDDEN",
		"No tienes permiso sobre las métricas de esta URL.")
	response.RegisterErrorMapping(ErrRangeNotAllowedForPlan, http.StatusForbidden, "PLAN_RANGE_NOT_ALLOWED",
		"El histórico de 7 y 30 días está disponible solo en el plan Pro. Tu plan actual permite hasta 24h.")
}

// EventPublisher es el puente hacia el hub de WebSocket. Se inyecta por
// constructor para no importar el paquete realtime y mantener limpia la dirección
// de dependencias. Es fire-and-forget sin error: un fallo de WS jamás debe romper
// el flujo de analytics.
type EventPublisher interface {
	PublishClickEvent(event ClickEvent)
}

// NoopEventPublisher descarta los eventos. Placeholder mientras Agent 5 no exista.
type NoopEventPublisher struct{}

func (NoopEventPublisher) PublishClickEvent(ClickEvent) {}

// AnalyticsService es la lógica de negocio del dominio. Embebe redirect.ClickRecorder
// para que su implementación pueda inyectarse directamente donde Agent 3 espera un
// ClickRecorder (en main.go, reemplazando al NoopClickRecorder).
type AnalyticsService interface {
	redirect.ClickRecorder // RecordClick(ctx, redirect.ClickEvent) error
	// GetStats devuelve las métricas agregadas del usuario. Si shortCode != nil,
	// valida ownership (403 FORBIDDEN si no es suya); si es nil, agrega todas sus URLs.
	GetStats(ctx context.Context, ownerID string, shortCode *string, rango string) (*ClickStats, error)
	// ExportStatsCSV devuelve las mismas métricas que GetStats serializadas a CSV.
	// Reutiliza GetStats, así que aplica idénticas validaciones (ownership, plan/rango).
	ExportStatsCSV(ctx context.Context, ownerID string, shortCode *string, rango string) ([]byte, error)
}

type analyticsService struct {
	repo      ClickRepository
	urls      shortener.URLRepository // dependencia legítima para validar ownership
	users     auth.UserRepository     // para leer el plan del usuario (Agent 7)
	publisher EventPublisher
}

func NewAnalyticsService(repo ClickRepository, urls shortener.URLRepository, users auth.UserRepository, publisher EventPublisher) AnalyticsService {
	return &analyticsService{repo: repo, urls: urls, users: users, publisher: publisher}
}

// RecordClick implementa redirect.ClickRecorder.
func (s *analyticsService) RecordClick(ctx context.Context, event redirect.ClickEvent) error {
	// Mismo shape de struct: conversión directa, sin acoplar el modelo interno a
	// la firma de redirect más allá de este punto de traducción.
	ev := ClickEvent(event)

	if err := s.repo.Write(ctx, ev); err != nil {
		return err
	}
	s.publisher.PublishClickEvent(ev)
	return nil
}

func (s *analyticsService) GetStats(ctx context.Context, ownerID string, shortCode *string, rango string) (*ClickStats, error) {
	tr, err := ParseTimeRange(rango)
	if err != nil {
		return nil, err // ErrInvalidRange -> 400, sin tocar InfluxDB
	}

	// Restricción por plan: Free solo accede a 24h. El plan se consulta
	// en DB en cada request para que un cambio surta efecto inmediato (sin relogin).
	user, err := s.users.FindByID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("analytics: no se pudo verificar el plan del usuario: %w", err)
	}
	if user.Plan == auth.PlanFree && tr != Range24h {
		return nil, ErrRangeNotAllowedForPlan // sin tocar InfluxDB
	}

	code := ""
	if shortCode != nil && *shortCode != "" {
		code = *shortCode
		url, err := s.urls.FindByCode(ctx, code)
		if err != nil {
			// No filtramos existencia: una URL inexistente se trata como ajena.
			if errors.Is(err, shortener.ErrURLNotFound) {
				return nil, ErrForbidden
			}
			return nil, fmt.Errorf("analytics: validando ownership: %w", err)
		}
		if url.OwnerID != ownerID {
			return nil, ErrForbidden
		}
	}

	return s.repo.QueryStats(ctx, code, ownerID, tr)
}

// ExportStatsCSV reutiliza GetStats (misma validación de ownership y plan/rango)
// y serializa el resultado a CSV. Si GetStats falla (ErrForbidden, ErrInvalidRange,
// ErrRangeNotAllowedForPlan...) ese error se propaga tal cual: el handler lo
// traduce al envelope JSON estándar, no se genera ningún CSV.
func (s *analyticsService) ExportStatsCSV(ctx context.Context, ownerID string, shortCode *string, rango string) ([]byte, error) {
	stats, err := s.GetStats(ctx, ownerID, shortCode, rango)
	if err != nil {
		return nil, err
	}
	// rango ya fue validado por GetStats; el error de ParseTimeRange aquí es imposible.
	tr, _ := ParseTimeRange(rango)
	return statsToCSV(stats, tr), nil
}
