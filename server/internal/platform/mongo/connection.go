// Package mongo gestiona el cliente compartido de MongoDB.
package mongo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// connectTimeout limita el tiempo de conexión + ping en el arranque.
const connectTimeout = 10 * time.Second

// Connect abre el cliente Mongo, verifica con un ping y devuelve la *mongo.Database.
// Falla rápido en boot si la conexión no es válida.
func Connect(uri, dbName string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo: no se pudo conectar a %q: %w", uri, err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo: ping fallido a %q: %w", uri, err)
	}

	slog.Info("conexión a MongoDB establecida", "uri", uri, "db", dbName)
	return client.Database(dbName), nil
}
