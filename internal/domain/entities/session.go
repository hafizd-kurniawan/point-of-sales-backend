package entities

import (
	"time"
)

type UserSession struct {
	ID           int        `json:"id" db:"id"`
	UserID       int        `json:"user_id" db:"user_id"`
	SessionToken string     `json:"session_token" db:"session_token"`
	LoginAt      time.Time  `json:"login_at" db:"login_at"`
	LogoutAt     *time.Time `json:"logout_at,omitempty" db:"logout_at"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
