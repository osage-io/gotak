package auth

import (
	"time"
	"github.com/google/uuid"
)

// SimpleUser represents a user in the existing database schema
type SimpleUser struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Role         string     `json:"role" db:"role"`
	Callsign     *string    `json:"callsign,omitempty" db:"callsign"`
	Active       bool       `json:"active" db:"active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
}

// SimpleRegisterRequest represents a simplified registration request
type SimpleRegisterRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	Callsign *string `json:"callsign,omitempty"`
}

// SimpleLoginRequest represents a simplified login request
type SimpleLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}