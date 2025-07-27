package repositories

import (
	"database/sql"
	"fmt"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
)

type sessionRepositoryImpl struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) repositories.SessionRepository {
	return &sessionRepositoryImpl{db: db}
}

func (r *sessionRepositoryImpl) Create(session *entities.UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, session_token, login_at, ip_address, is_active)
		VALUES (:user_id, :session_token, :login_at, :ip_address, :is_active)
		RETURNING id, created_at, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(session, session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *sessionRepositoryImpl) GetByToken(token string) (*entities.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, login_at, logout_at, ip_address, is_active,
		       created_at, updated_at, deleted_at
		FROM user_sessions 
		WHERE session_token = $1 AND deleted_at IS NULL AND is_active = true
	`

	var session entities.UserSession
	err := r.db.Get(&session, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepositoryImpl) GetActiveByUserID(userID int) ([]entities.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, login_at, logout_at, ip_address, is_active,
		       created_at, updated_at, deleted_at
		FROM user_sessions 
		WHERE user_id = $1 AND deleted_at IS NULL AND is_active = true
		ORDER BY login_at DESC
	`

	var sessions []entities.UserSession
	err := r.db.Select(&sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

func (r *sessionRepositoryImpl) UpdateLogout(token string) error {
	query := `
		UPDATE user_sessions 
		SET logout_at = CURRENT_TIMESTAMP, is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE session_token = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *sessionRepositoryImpl) Delete(token string) error {
	query := `
		UPDATE user_sessions 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE session_token = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *sessionRepositoryImpl) DeleteAllByUserID(userID int) error {
	query := `
		UPDATE user_sessions 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}
