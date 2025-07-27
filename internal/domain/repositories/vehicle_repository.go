package repositories

import (
	"vehicle-showroom-backend/internal/domain/entities"
)

type VehicleRepository interface {
	Create(vehicle *entities.Vehicle) error
	GetByID(id int) (*entities.Vehicle, error)
	GetByVehicleCode(vehicleCode string) (*entities.Vehicle, error)
	GetAll(limit, offset int) ([]entities.Vehicle, error)
	Search(req *entities.VehicleSearchRequest) ([]entities.Vehicle, int, error)
	Update(id int, vehicle *entities.Vehicle) error
	Delete(id int) error
	Count() (int, error)
	GenerateVehicleCode() (string, error)
}

type VehicleCategoryRepository interface {
	Create(category *entities.VehicleCategory) error
	GetByID(id int) (*entities.VehicleCategory, error)
	GetAll() ([]entities.VehicleCategory, error)
	Update(id int, category *entities.VehicleCategory) error
	Delete(id int) error
}

type VehiclePhotoRepository interface {
	Create(photo *entities.VehiclePhoto) error
	GetByVehicleID(vehicleID int) ([]entities.VehiclePhoto, error)
	GetByID(id int) (*entities.VehiclePhoto, error)
	Update(id int, photo *entities.VehiclePhoto) error
	Delete(id int) error
	SetPrimary(vehicleID int, photoID int) error
	UpdateSortOrder(photoID int, sortOrder int) error
}
