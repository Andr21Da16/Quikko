// Package auth implementa registro, login y emisión/renovación de JWT.
// Expone un userID autenticado (hex de ObjectID) que el resto de dominios
// consume como propietario de sus recursos.
package auth

import (
	"errors"
	"time"
)

// Plan es el plan de suscripción del usuario. Agent 7 lo usa como capa transversal:
// shortener limita las URLs activas del plan Free y analytics restringe el rango
// histórico. El plan NO se embebe en el JWT (se consulta en DB en cada operación
// relevante) para que un cambio de plan surta efecto inmediato sin relogin.
type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
)

// User es el modelo de dominio. Contrato estable: Agent 2 (Shortener) referencia
// al propietario de cada URL por su ID (string, hex de ObjectID).
//
// PasswordHash lleva json:"-" para que nunca se filtre si el struct llegara a
// serializarse a JSON por error; la serialización de salida usa los DTOs.
type User struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Email        string    `bson:"email" json:"email"`
	PasswordHash string    `bson:"passwordHash" json:"-"`
	Plan         Plan      `bson:"plan" json:"plan"` // default PlanFree al registrarse
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
}

// AccountSummary es el agregado de cuenta que alimenta el header/sidebar del
// dashboard. Junta datos del User con métricas derivadas de
// sus URLs (vía AccountURLStore). ActiveURLsLimit es nil cuando el plan es Pro
// (ilimitado); para Free lleva el cupo configurado.
type AccountSummary struct {
	Email           string
	Plan            Plan
	CreatedAt       time.Time
	ActiveURLsCount int64
	ActiveURLsLimit *int
	TotalClicks     int64
}

// Errores de dominio. El handler/error_handler los traduce a códigos HTTP.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	// ErrPlanLimitExceeded lo devuelve shortener al superar el cupo de URLs activas
	// del plan Free. Vive aquí (con el modelo Plan) y shortener registra su mapeo
	// HTTP con el límite configurado interpolado.
	ErrPlanLimitExceeded = errors.New("plan limit exceeded")
)
