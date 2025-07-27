package repositories

import (
	"vehicle-showroom-backend/internal/domain/entities"
)

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id int) (*entities.User, error)
	GetByUsername(username string) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
	GetAll(limit, offset int) ([]entities.User, error)
	Update(id int, user *entities.User) error
	Delete(id int) error
	Count() (int, error)
}
