// Package redis gestiona el cliente compartido de Redis.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const connectTimeout = 10 * time.Second

// Connect abre el cliente Redis y verifica con PING. Falla rápido en boot.
func Connect(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping fallido a %q (db %d): %w", addr, db, err)
	}

	slog.Info("conexión a Redis establecida", "addr", addr, "db", db)
	return client, nil
}
