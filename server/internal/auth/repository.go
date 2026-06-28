package auth

import "context"

// UserRepository es la interfaz de persistencia que consume el service.
// El service nunca conoce la implementación concreta (Mongo).
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	// UpdatePlan cambia el plan del usuario. Devuelve ErrUserNotFound si
	// el id no corresponde a ningún usuario.
	UpdatePlan(ctx context.Context, id string, plan Plan) error
	// UpdatePassword reemplaza el hash de password del usuario.
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	// Delete elimina al usuario de la fuente de verdad . El
	// borrado en cascada de sus URLs/cache lo orquesta el service vía AccountURLStore.
	Delete(ctx context.Context, id string) error
}

// AccountURLStore es lo que auth necesita de las URLs de un usuario para gestionar
// su cuenta (resumen y borrado en cascada). Se declara aquí.
type AccountURLStore interface {
	// CountActiveByOwner cuenta las URLs activas del usuario (para el cupo del resumen).
	CountActiveByOwner(ctx context.Context, ownerID string) (int64, error)
	// SumClicksByOwner suma el TotalClicks de todas las URLs del usuario.
	SumClicksByOwner(ctx context.Context, ownerID string) (int64, error)
	// ShortCodesByOwner devuelve los shortCodes del usuario (para purgar el cache).
	ShortCodesByOwner(ctx context.Context, ownerID string) ([]string, error)
	// DeleteAllByOwner elimina todas las URLs del usuario en la fuente de verdad.
	DeleteAllByOwner(ctx context.Context, ownerID string) error
	// PurgeCache elimina una entrada del cache de redirección por shortCode.
	PurgeCache(ctx context.Context, shortCode string) error
}
