package shortener

import "context"

// URLRepository es la persistencia de metadatos en Mongo (fuente de verdad).
type URLRepository interface {
	Create(ctx context.Context, url *ShortURL) error
	FindByCode(ctx context.Context, code string) (*ShortURL, error)

	FindByID(ctx context.Context, id string) (*ShortURL, error)

	FindByOwner(ctx context.Context, ownerID string, filter ListFilter, page, limit int) (urls []*ShortURL, total int64, err error)

	CountActiveByOwner(ctx context.Context, ownerID string) (int64, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
	Update(ctx context.Context, url *ShortURL) error
	Delete(ctx context.Context, id, ownerID string) error

	IncrementClicks(ctx context.Context, shortCode string) error

	SumClicksByOwner(ctx context.Context, ownerID string) (int64, error)

	FindShortCodesByOwner(ctx context.Context, ownerID string) ([]string, error)

	DeleteAllByOwner(ctx context.Context, ownerID string) error
}

// CacheRepository es el cache de redirección en Redis (shortCode -> originalURL).
type CacheRepository interface {
	Set(ctx context.Context, code, originalURL string) error
	Get(ctx context.Context, code string) (string, error)
	Delete(ctx context.Context, code string) error
}
