package shortener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"urlshortener/internal/auth"
	"urlshortener/pkg/response"
)

// aliasPattern: alfanumérico y guiones, 3-30 caracteres (validación de negocio
// del alias personalizado, complementaria a la del DTO).
var aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9-]{3,30}$`)

func init() {
	response.RegisterErrorMapping(ErrAliasTaken, http.StatusConflict, "ALIAS_TAKEN",
		"El alias ya está en uso. Elige otro.")
	response.RegisterErrorMapping(ErrURLNotFound, http.StatusNotFound, "URL_NOT_FOUND",
		"La URL no existe.")
	response.RegisterErrorMapping(ErrNotOwner, http.StatusForbidden, "FORBIDDEN",
		"No tienes permiso sobre este recurso.")
	response.RegisterErrorMapping(ErrCodeGenFailed, http.StatusInternalServerError, "INTERNAL_ERROR",
		"No se pudo generar un código único. Inténtalo de nuevo.")
	response.RegisterErrorMapping(ErrURLInactive, http.StatusGone, "URL_INACTIVE",
		"Esta URL ha sido desactivada.")

	response.RegisterErrorMapping(ErrInvalidURL, http.StatusBadRequest, "INVALID_URL",
		"La URL proporcionada no es válida o no está permitida.")
}

// ShortenerService define la lógica de negocio del dominio.
type ShortenerService interface {
	CreateShortURL(ctx context.Context, ownerID, originalURL string, customAlias *string) (*ShortURL, error)
	CheckAliasAvailability(ctx context.Context, alias string) (available bool, err error)
	ListUserURLs(ctx context.Context, ownerID string, filter ListFilter, page, limit int) (urls []*ShortURL, total int64, err error)
	// GetOwnedURLByCode resuelve una URL del usuario por su shortCode. Alimenta la
	// página de detalle del frontend, que necesita los metadatos + QR de
	// una sola URL sin listar todas. Como loadOwned, devuelve ErrNotOwner tanto si no
	// existe como si no es del usuario (no filtra existencia).
	GetOwnedURLByCode(ctx context.Context, ownerID, shortCode string) (*ShortURL, error)
	ToggleActive(ctx context.Context, ownerID, urlID string, isActive bool) error
	DeleteURL(ctx context.Context, ownerID, urlID string) error
}

type shortenerService struct {
	repo            URLRepository
	cache           CacheRepository
	users           auth.UserRepository // para leer el plan del usuario (Agent 7)
	freePlanMaxURLs int                 // cupo de URLs activas del plan Free
	baseURL         string              // dominio propio del servicio, para la validación de URL (Agent 9)
}

func NewShortenerService(repo URLRepository, cache CacheRepository, users auth.UserRepository, freePlanMaxURLs int, baseURL string) ShortenerService {
	return &shortenerService{repo: repo, cache: cache, users: users, freePlanMaxURLs: freePlanMaxURLs, baseURL: baseURL}
}

func (s *shortenerService) ensureActiveQuota(ctx context.Context, ownerID string) error {
	user, err := s.users.FindByID(ctx, ownerID)
	if err != nil {
		return fmt.Errorf("shortener: no se pudo verificar el plan del usuario: %w", err)
	}
	if user.Plan != auth.PlanFree {
		return nil
	}
	active, err := s.repo.CountActiveByOwner(ctx, ownerID)
	if err != nil {
		return err
	}
	if active >= int64(s.freePlanMaxURLs) {
		return auth.ErrPlanLimitExceeded
	}
	return nil
}

func (s *shortenerService) CreateShortURL(ctx context.Context, ownerID, originalURL string, customAlias *string) (*ShortURL, error) {

	if err := ValidateOriginalURL(originalURL, s.baseURL); err != nil {
		return nil, err
	}

	// Cupo del plan: si está lleno, no se genera código ni se toca Redis.
	if err := s.ensureActiveQuota(ctx, ownerID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	url := &ShortURL{
		OriginalURL: originalURL,
		OwnerID:     ownerID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if customAlias != nil {
		alias := *customAlias
		if !aliasPattern.MatchString(alias) {
			return nil, ErrAliasTaken // alias inválido se trata como no disponible
		}
		taken, err := s.repo.ExistsByCode(ctx, alias)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, ErrAliasTaken // sin reintento: el usuario eligió este alias
		}
		url.ShortCode = alias
		url.IsCustomAlias = true
		if err := s.repo.Create(ctx, url); err != nil {
			return nil, err // posible ErrAliasTaken por carrera
		}
	} else {
		if err := s.createWithGeneratedCode(ctx, url); err != nil {
			return nil, err
		}
	}

	// Poblar el cache no es bloqueante para la creación: si Redis falla, Mongo
	// sigue siendo la verdad y Agent 3 repobla on-demand.
	if err := s.cache.Set(ctx, url.ShortCode, url.OriginalURL); err != nil {
		slog.Error("no se pudo poblar el cache tras crear URL", "shortCode", url.ShortCode, "err", err)
	}

	return url, nil
}

// createWithGeneratedCode genera un código y reintenta ante colisión hasta
// maxRetries, devolviendo ErrCodeGenFailed si se agotan.
func (s *shortenerService) createWithGeneratedCode(ctx context.Context, url *ShortURL) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := GenerateCode(defaultCodeLength)
		if err != nil {
			return err
		}

		exists, err := s.repo.ExistsByCode(ctx, code)
		if err != nil {
			return err
		}
		if exists {
			continue // colisión: reintentar con otro código
		}

		url.ShortCode = code
		url.IsCustomAlias = false
		err = s.repo.Create(ctx, url)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrAliasTaken) {
			continue // carrera: otro proceso tomó el código entre el check y el insert
		}
		return err
	}

	slog.Warn("agotados los reintentos de generación de código: el espacio de códigos puede estar saturándose",
		"maxRetries", maxRetries, "length", defaultCodeLength)
	return ErrCodeGenFailed
}

func (s *shortenerService) CheckAliasAvailability(ctx context.Context, alias string) (bool, error) {
	if !aliasPattern.MatchString(alias) {
		return false, nil // un alias con formato inválido no está disponible
	}
	exists, err := s.repo.ExistsByCode(ctx, alias)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *shortenerService) ListUserURLs(ctx context.Context, ownerID string, filter ListFilter, page, limit int) ([]*ShortURL, int64, error) {
	return s.repo.FindByOwner(ctx, ownerID, filter, page, limit)
}

func (s *shortenerService) GetOwnedURLByCode(ctx context.Context, ownerID, shortCode string) (*ShortURL, error) {
	url, err := s.repo.FindByCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, ErrURLNotFound) {
			return nil, ErrNotOwner // no filtra existencia (misma decisión que loadOwned)
		}
		return nil, fmt.Errorf("shortener: error cargando URL por código: %w", err)
	}
	if url.OwnerID != ownerID {
		return nil, ErrNotOwner
	}
	return url, nil
}

func (s *shortenerService) ToggleActive(ctx context.Context, ownerID, urlID string, isActive bool) error {
	url, err := s.loadOwned(ctx, ownerID, urlID)
	if err != nil {
		return err
	}

	// Reactivar una URL desactivada consume cupo igual que crear una nueva. Solo se
	// valida en la transición inactiva -> activa: si ya estaba activa, el conteo no
	// cambia y no debe rechazarse (evita falsos positivos al estar justo en el límite).
	if isActive && !url.IsActive {
		if err := s.ensureActiveQuota(ctx, ownerID); err != nil {
			return err
		}
	}

	url.IsActive = isActive
	url.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, url); err != nil {
		return err
	}

	// Sincronizar Redis para que Agent 3 refleje el cambio de inmediato.
	if isActive {
		if err := s.cache.Set(ctx, url.ShortCode, url.OriginalURL); err != nil {
			slog.Error("no se pudo repoblar el cache al activar URL", "shortCode", url.ShortCode, "err", err)
		}
	} else {
		if err := s.cache.Delete(ctx, url.ShortCode); err != nil {
			slog.Error("no se pudo eliminar del cache al desactivar URL", "shortCode", url.ShortCode, "err", err)
		}
	}
	return nil
}

func (s *shortenerService) DeleteURL(ctx context.Context, ownerID, urlID string) error {
	url, err := s.loadOwned(ctx, ownerID, urlID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, urlID, ownerID); err != nil {
		return err
	}
	if err := s.cache.Delete(ctx, url.ShortCode); err != nil {
		slog.Error("no se pudo eliminar del cache al borrar URL", "shortCode", url.ShortCode, "err", err)
	}
	return nil
}

// loadOwned carga una URL y verifica ownership. Tanto si no existe como si no es
// del usuario, devuelve ErrNotOwner: así no se filtra si el id existe (misma
// decisión de seguridad que en login).
func (s *shortenerService) loadOwned(ctx context.Context, ownerID, urlID string) (*ShortURL, error) {
	url, err := s.repo.FindByID(ctx, urlID)
	if err != nil {
		if errors.Is(err, ErrURLNotFound) {
			return nil, ErrNotOwner
		}
		return nil, fmt.Errorf("shortener: error cargando URL: %w", err)
	}
	if url.OwnerID != ownerID {
		return nil, ErrNotOwner
	}
	return url, nil
}
