package auth

import "time"

// DTOs de entrada/salida del dominio auth. La validación usa go-playground/validator
// integrado con el binding de Echo (c.Validate).

// Máximos de longitud: email = 254 (RFC 5321); password = 72, que es el
// límite real de bytes que bcrypt considera (los bytes extra se ignoran al hashear, así
// que rechazarlos es más honesto y evita malgastar trabajo de hashing con inputs enormes).
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,max=72"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// UpdatePlanRequest es el body de PATCH /me/plan.
type UpdatePlanRequest struct {
	Plan string `json:"plan" validate:"required,oneof=free pro"`
}

// ChangePasswordRequest es el body de PATCH /me/password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,max=72"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=72"`
}

// DeleteAccountRequest es el body de DELETE /me. Exige la
// password actual para confirmar identidad antes del borrado en cascada.
type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

// AccountSummaryResponse es la salida de GET /me/summary.
// ActiveUrlsLimit es null cuando el plan es Pro (ilimitado).
type AccountSummaryResponse struct {
	Email           string `json:"email"`
	Plan            string `json:"plan"`
	CreatedAt       string `json:"createdAt"`
	ActiveUrlsCount int64  `json:"activeUrlsCount"`
	ActiveUrlsLimit *int   `json:"activeUrlsLimit"`
	TotalClicks     int64  `json:"totalClicks"`
}

// toAccountSummaryResponse traduce el agregado de dominio a su DTO de salida.
func toAccountSummaryResponse(s *AccountSummary) AccountSummaryResponse {
	return AccountSummaryResponse{
		Email:           s.Email,
		Plan:            string(s.Plan),
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
		ActiveUrlsCount: s.ActiveURLsCount,
		ActiveUrlsLimit: s.ActiveURLsLimit,
		TotalClicks:     s.TotalClicks,
	}
}

type AuthResponse struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken,omitempty"`
	User         UserDTO `json:"user"`
}

type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

// toUserDTO traduce el modelo de dominio a su representación de salida.
func toUserDTO(u *User) UserDTO {
	return UserDTO{ID: u.ID, Email: u.Email, Plan: string(u.Plan)}
}
