// Package shortener implementa el CRUD de URLs cortas: creación (autogenerada o
// con alias), gestión (listar, activar/desactivar, eliminar) y el mantenimiento
// del cache de redirección en Redis.
//

package shortener

import (
	"errors"
	"time"
)

// ShortURL es el modelo de dominio de una URL acortada.
type ShortURL struct {
	ID            string    `bson:"_id,omitempty"`
	ShortCode     string    `bson:"shortCode"` // único, indexado
	OriginalURL   string    `bson:"originalUrl"`
	OwnerID       string    `bson:"ownerId"` // referencia a auth.User.ID
	IsCustomAlias bool      `bson:"isCustomAlias"`
	IsActive      bool      `bson:"isActive"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
	// TotalClicks es un contador aproximado de clics, incrementado de forma atómica
	// ($inc) en su goroutine async de registro de clic.
	// Default 0 al crear. Es "aproximado" porque el incremento es fire-and-forget:
	// un fallo se loguea pero no se reintenta (la fuente fina de verdad es InfluxDB).
	TotalClicks int64 `bson:"totalClicks"`
}

// ListFilter agrupa los filtros opcionales del listado de URLs de un usuario.
// Se resuelve en la query de Mongo (no en memoria) para que escale con el plan Pro (sin
// tope de URLs).
type ListFilter struct {
	// Search: coincidencia parcial case-insensitive sobre shortCode u originalUrl.
	// Vacío = sin búsqueda. Se trata como texto literal (regexp.QuoteMeta) en el repo.
	Search string
	// IsActive: nil = todas; &true = solo activas; &false = solo inactivas.
	IsActive *bool
}

// Errores de dominio. El handler los traduce a códigos HTTP.
var (
	ErrAliasTaken    = errors.New("alias already taken")
	ErrURLNotFound   = errors.New("url not found")
	ErrNotOwner      = errors.New("user is not the owner of this url")
	ErrCodeGenFailed = errors.New("failed to generate unique short code after max retries")

	ErrURLInactive = errors.New("url is inactive")

	ErrInvalidURL = errors.New("invalid or disallowed original url")
)
