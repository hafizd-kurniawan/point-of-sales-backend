package repositories

import (
	"vehicle-showroom-backend/internal/domain/entities"
)

type SessionRepository interface {
	Create(session *entities.UserSession) error
	GetByToken(token string) (*entities.UserSession, error)
	GetActiveByUserID(userID int) ([]entities.UserSession, error)
	UpdateLogout(token string) error
	Delete(token string) error
	DeleteAllByUserID(userID int) error
}
