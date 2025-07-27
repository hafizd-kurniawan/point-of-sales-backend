package usecases

import (
	"fmt"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"
)

type VehicleUsecase interface {
	CreateVehicle(request *entities.CreateVehicleRequest) (*entities.Vehicle, error)
	GetVehicleByID(id int) (*entities.Vehicle, error)
	GetVehicles(page, limit int) ([]entities.Vehicle, int, error)
	SearchVehicles(request *entities.VehicleSearchRequest) ([]entities.Vehicle, int, error)
	UpdateVehicle(id int, request *entities.UpdateVehicleRequest) (*entities.Vehicle, error)
	DeleteVehicle(id int) error
}

type VehicleCategoryUsecase interface {
	CreateCategory(request *entities.CreateCategoryRequest) (*entities.VehicleCategory, error)
	GetCategoryByID(id int) (*entities.VehicleCategory, error)
	GetCategories() ([]entities.VehicleCategory, error)
	UpdateCategory(id int, request *entities.UpdateCategoryRequest) (*entities.VehicleCategory, error)
	DeleteCategory(id int) error
}

type vehicleUsecase struct {
	vehicleRepo repositories.VehicleRepository
	config      *config.Config
}

type vehicleCategoryUsecase struct {
	categoryRepo repositories.VehicleCategoryRepository
	config       *config.Config
}

func NewVehicleUsecase(vehicleRepo repositories.VehicleRepository, config *config.Config) VehicleUsecase {
	return &vehicleUsecase{
		vehicleRepo: vehicleRepo,
		config:      config,
	}
}

func NewVehicleCategoryUsecase(categoryRepo repositories.VehicleCategoryRepository, config *config.Config) VehicleCategoryUsecase {
	return &vehicleCategoryUsecase{
		categoryRepo: categoryRepo,
		config:       config,
	}
}

// Vehicle UseCase Implementation
func (u *vehicleUsecase) CreateVehicle(request *entities.CreateVehicleRequest) (*entities.Vehicle, error) {
	// Create vehicle entity
	vehicle := &entities.Vehicle{
		CategoryID:         request.CategoryID,
		Brand:              request.Brand,
		Model:              request.Model,
		Year:               request.Year,
		Color:              request.Color,
		EngineType:         request.EngineType,
		Transmission:       request.Transmission,
		FuelType:           request.FuelType,
		ChassisNumber:      request.ChassisNumber,
		EngineNumber:       request.EngineNumber,
		LicensePlate:       request.LicensePlate,
		PurchasePrice:      request.PurchasePrice,
		SellingPrice:       request.SellingPrice,
		ConditionStatus:    request.ConditionStatus,
		AvailabilityStatus: request.AvailabilityStatus,
		Location:           request.Location,
		Mileage:            request.Mileage,
		Description:        request.Description,
		Features:           request.Features,
		IsActive:           true,
	}

	// Set default availability status if not provided
	if vehicle.AvailabilityStatus == "" {
		vehicle.AvailabilityStatus = entities.AvailabilityAvailable
	}

	// Save to database
	if err := u.vehicleRepo.Create(vehicle); err != nil {
		return nil, fmt.Errorf("failed to create vehicle: %w", err)
	}

	return vehicle, nil
}

func (u *vehicleUsecase) GetVehicleByID(id int) (*entities.Vehicle, error) {
	vehicle, err := u.vehicleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	return vehicle, nil
}

func (u *vehicleUsecase) GetVehicles(page, limit int) ([]entities.Vehicle, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	vehicles, err := u.vehicleRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get vehicles: %w", err)
	}

	// Get total count
	total, err := u.vehicleRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count vehicles: %w", err)
	}

	return vehicles, total, nil
}

func (u *vehicleUsecase) SearchVehicles(request *entities.VehicleSearchRequest) ([]entities.Vehicle, int, error) {
	vehicles, total, err := u.vehicleRepo.Search(request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search vehicles: %w", err)
	}

	return vehicles, total, nil
}

func (u *vehicleUsecase) UpdateVehicle(id int, request *entities.UpdateVehicleRequest) (*entities.Vehicle, error) {
	// Get existing vehicle
	existingVehicle, err := u.vehicleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	// Prepare update entity
	updateVehicle := &entities.Vehicle{
		CategoryID:         request.CategoryID,
		Brand:              request.Brand,
		Model:              request.Model,
		Color:              request.Color,
		EngineType:         request.EngineType,
		Transmission:       request.Transmission,
		FuelType:           request.FuelType,
		ChassisNumber:      request.ChassisNumber,
		EngineNumber:       request.EngineNumber,
		LicensePlate:       request.LicensePlate,
		ConditionStatus:    request.ConditionStatus,
		AvailabilityStatus: request.AvailabilityStatus,
		Location:           request.Location,
		Description:        request.Description,
		Features:           request.Features,
	}

	// Handle optional fields
	if request.Year > 0 {
		updateVehicle.Year = request.Year
	} else {
		updateVehicle.Year = existingVehicle.Year
	}

	if request.PurchasePrice != nil {
		updateVehicle.PurchasePrice = *request.PurchasePrice
	} else {
		updateVehicle.PurchasePrice = existingVehicle.PurchasePrice
	}

	if request.SellingPrice != nil {
		updateVehicle.SellingPrice = *request.SellingPrice
	} else {
		updateVehicle.SellingPrice = existingVehicle.SellingPrice
	}

	if request.Mileage != nil {
		updateVehicle.Mileage = *request.Mileage
	} else {
		updateVehicle.Mileage = existingVehicle.Mileage
	}

	if request.IsActive != nil {
		updateVehicle.IsActive = *request.IsActive
	} else {
		updateVehicle.IsActive = existingVehicle.IsActive
	}

	// Update vehicle
	if err := u.vehicleRepo.Update(id, updateVehicle); err != nil {
		return nil, fmt.Errorf("failed to update vehicle: %w", err)
	}

	// Get updated vehicle
	updatedVehicle, err := u.vehicleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated vehicle: %w", err)
	}

	return updatedVehicle, nil
}

func (u *vehicleUsecase) DeleteVehicle(id int) error {
	// Check if vehicle exists
	_, err := u.vehicleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("vehicle not found: %w", err)
	}

	// Soft delete vehicle
	if err := u.vehicleRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	return nil
}

// Category UseCase Implementation
func (u *vehicleCategoryUsecase) CreateCategory(request *entities.CreateCategoryRequest) (*entities.VehicleCategory, error) {
	category := &entities.VehicleCategory{
		Name:        request.Name,
		Description: request.Description,
		IsActive:    true,
	}

	if err := u.categoryRepo.Create(category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (u *vehicleCategoryUsecase) GetCategoryByID(id int) (*entities.VehicleCategory, error) {
	category, err := u.categoryRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	return category, nil
}

func (u *vehicleCategoryUsecase) GetCategories() ([]entities.VehicleCategory, error) {
	categories, err := u.categoryRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (u *vehicleCategoryUsecase) UpdateCategory(id int, request *entities.UpdateCategoryRequest) (*entities.VehicleCategory, error) {
	// Get existing category
	existingCategory, err := u.categoryRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Prepare update entity
	updateCategory := &entities.VehicleCategory{
		Name:        request.Name,
		Description: request.Description,
	}

	if request.IsActive != nil {
		updateCategory.IsActive = *request.IsActive
	} else {
		updateCategory.IsActive = existingCategory.IsActive
	}

	// Update category
	if err := u.categoryRepo.Update(id, updateCategory); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	// Get updated category
	updatedCategory, err := u.categoryRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated category: %w", err)
	}

	return updatedCategory, nil
}

func (u *vehicleCategoryUsecase) DeleteCategory(id int) error {
	// Check if category exists
	_, err := u.categoryRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Soft delete category
	if err := u.categoryRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
