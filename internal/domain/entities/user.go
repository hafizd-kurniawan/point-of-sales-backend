package entities

import (
	"time"
)

type UserRole string

const (
	RoleAdmin        UserRole = "admin"
	RoleMechanic     UserRole = "mechanic"
	RoleCashier      UserRole = "cashier"
	UserRoleAdmin    UserRole = "admin"
	UserRoleMechanic UserRole = "mechanic"
	UserRoleCashier  UserRole = "cashier"
)

type User struct {
	ID           int        `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	Phone        string     `json:"phone" db:"phone"`
	Role         UserRole   `json:"role" db:"role"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	FullName string   `json:"full_name" binding:"required,min=2,max=100"`
	Phone    string   `json:"phone" binding:"required"`
	Role     UserRole `json:"role" binding:"required,oneof=admin mechanic cashier"`
}

type UpdateUserRequest struct {
	Username string   `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string   `json:"email" binding:"omitempty,email"`
	Password string   `json:"password" binding:"omitempty,min=6"`
	FullName string   `json:"full_name" binding:"omitempty,min=2,max=100"`
	Phone    string   `json:"phone" binding:"omitempty"`
	Role     UserRole `json:"role" binding:"omitempty,oneof=admin mechanic cashier"`
	IsActive *bool    `json:"is_active" binding:"omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      User   `json:"user"`
}
