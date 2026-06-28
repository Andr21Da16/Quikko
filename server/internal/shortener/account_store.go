package shortener

import (
	"context"

	"urlshortener/internal/auth"
)

// accountURLStore adapta los repos de shortener (Mongo + cache Redis) a la interfaz
// auth.AccountURLStore, que auth consume para el resumen y el borrado en cascada de
// cuenta. Vive en shortener —que ya importa auth— porque auth no
// puede importar shortener sin crear un ciclo de import. main.go construye el
// adaptador y lo inyecta en auth.RegisterRoutes.
type accountURLStore struct {
	repo  URLRepository
	cache CacheRepository
}

// NewAccountURLStore devuelve un auth.AccountURLStore respaldado por los repos de
// shortener. Reutiliza el mismo esquema de clave de cache de redirección (no inventa
// otro), igual que redirect.
func NewAccountURLStore(repo URLRepository, cache CacheRepository) auth.AccountURLStore {
	return &accountURLStore{repo: repo, cache: cache}
}

func (s *accountURLStore) CountActiveByOwner(ctx context.Context, ownerID string) (int64, error) {
	return s.repo.CountActiveByOwner(ctx, ownerID)
}

func (s *accountURLStore) SumClicksByOwner(ctx context.Context, ownerID string) (int64, error) {
	return s.repo.SumClicksByOwner(ctx, ownerID)
}

func (s *accountURLStore) ShortCodesByOwner(ctx context.Context, ownerID string) ([]string, error) {
	return s.repo.FindShortCodesByOwner(ctx, ownerID)
}

func (s *accountURLStore) DeleteAllByOwner(ctx context.Context, ownerID string) error {
	return s.repo.DeleteAllByOwner(ctx, ownerID)
}

func (s *accountURLStore) PurgeCache(ctx context.Context, shortCode string) error {
	return s.cache.Delete(ctx, shortCode)
}
