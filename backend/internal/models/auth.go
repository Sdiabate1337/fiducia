package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID                  uuid.UUID `json:"id"`
	CabinetID           uuid.UUID `json:"cabinet_id"`
	Email               string    `json:"email"`
	PasswordHash        string    `json:"-"` // Never JSON encode
	FullName            string    `json:"full_name"`
	Role                string    `json:"role"`
	OnboardingCompleted bool      `json:"onboarding_completed"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UserRole constants
const (
	RoleAdmin        = "admin"
	RoleCollaborator = "collaborator"
)

// LoginRequest payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest payload
type RegisterRequest struct {
	CabinetName string `json:"cabinet_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
}

// LoginResponse payload
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// AuthClaims for JWT
type AuthClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	CabinetID uuid.UUID `json:"cabinet_id"`
	Role      string    `json:"role"`
	jwt.RegisteredClaims
}
