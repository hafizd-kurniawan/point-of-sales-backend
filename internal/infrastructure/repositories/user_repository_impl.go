package repositories

import (
	"database/sql"
	"fmt"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
)

type userRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repositories.UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(user *entities.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, phone, role, is_active)
		VALUES (:username, :email, :password_hash, :full_name, :phone, :role, :is_active)
		RETURNING id, created_at, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(user, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) GetByID(id int) (*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone, role, is_active, 
		       created_at, updated_at, deleted_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) GetByUsername(username string) (*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone, role, is_active, 
		       created_at, updated_at, deleted_at
		FROM users 
		WHERE username = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) GetByEmail(email string) (*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone, role, is_active, 
		       created_at, updated_at, deleted_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) GetAll(limit, offset int) ([]entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone, role, is_active, 
		       created_at, updated_at, deleted_at
		FROM users 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []entities.User
	err := r.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

func (r *userRepositoryImpl) Update(id int, user *entities.User) error {
	query := `
		UPDATE users 
		SET username = COALESCE(NULLIF(:username, ''), username),
		    email = COALESCE(NULLIF(:email, ''), email),
		    password_hash = COALESCE(NULLIF(:password_hash, ''), password_hash),
		    full_name = COALESCE(NULLIF(:full_name, ''), full_name),
		    phone = COALESCE(NULLIF(:phone, ''), phone),
		    role = COALESCE(NULLIF(:role, ''), role),
		    is_active = COALESCE(:is_active, is_active),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND deleted_at IS NULL
		RETURNING updated_at
	`

	user.ID = id
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(&user.UpdatedAt, user)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) Delete(id int) error {
	query := `
		UPDATE users 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepositoryImpl) Count() (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var count int
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}
