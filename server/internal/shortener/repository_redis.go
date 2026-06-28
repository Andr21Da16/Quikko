package shortener

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// cacheKeyPrefix namespacea las claves de redirección en Redis. Contrato implícito
// (Redirect): la clave de lookup es cacheKeyPrefix + shortCode.
const cacheKeyPrefix = "shorturl:"

// cacheTTL: decisión documentada. Usamos un TTL largo (30 días) en vez de claves
// eternas, para que entradas de URLs poco usadas caduquen solas y Redis no crezca
// indefinidamente. Mongo es la fuente de verdad; si la URL se borra en Mongo,
// el cache caducará solo. El TTL no es crítico: si Redis falla, el service hace
// fallback a Mongo y re-cachea la URL al primer clic.
const cacheTTL = 30 * 24 * time.Hour

var ErrCacheMiss = errors.New("shortener: cache miss")

type redisCacheRepository struct {
	client *redis.Client
}

// NewCacheRepository construye el cache de redirección sobre Redis.
func NewCacheRepository(client *redis.Client) CacheRepository {
	return &redisCacheRepository{client: client}
}

func (r *redisCacheRepository) key(code string) string {
	return cacheKeyPrefix + code
}

func (r *redisCacheRepository) Set(ctx context.Context, code, originalURL string) error {
	if err := r.client.Set(ctx, r.key(code), originalURL, cacheTTL).Err(); err != nil {
		return fmt.Errorf("shortener: error escribiendo en cache: %w", err)
	}
	return nil
}

func (r *redisCacheRepository) Get(ctx context.Context, code string) (string, error) {
	val, err := r.client.Get(ctx, r.key(code)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrCacheMiss
		}
		return "", fmt.Errorf("shortener: error leyendo de cache: %w", err)
	}
	return val, nil
}

func (r *redisCacheRepository) Delete(ctx context.Context, code string) error {
	if err := r.client.Del(ctx, r.key(code)).Err(); err != nil {
		return fmt.Errorf("shortener: error eliminando de cache: %w", err)
	}
	return nil
}
