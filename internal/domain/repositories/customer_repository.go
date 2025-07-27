package repositories

import (
	"vehicle-showroom-backend/internal/domain/entities"
)

type CustomerRepository interface {
	Create(customer *entities.Customer) error
	GetByID(id int) (*entities.Customer, error)
	GetByCustomerCode(customerCode string) (*entities.Customer, error)
	GetAll(limit, offset int) ([]entities.Customer, error)
	Search(req *entities.CustomerSearchRequest) ([]entities.Customer, int, error)
	Update(id int, customer *entities.Customer) error
	Delete(id int) error
	Count() (int, error)
	GenerateCustomerCode() (string, error)
}
