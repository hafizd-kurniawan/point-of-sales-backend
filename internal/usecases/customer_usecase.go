package usecases

import (
	"fmt"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"
)

type CustomerUsecase interface {
	CreateCustomer(request *entities.CreateCustomerRequest) (*entities.Customer, error)
	GetCustomerByID(id int) (*entities.Customer, error)
	GetCustomers(page, limit int) ([]entities.Customer, int, error)
	SearchCustomers(request *entities.CustomerSearchRequest) ([]entities.Customer, int, error)
	UpdateCustomer(id int, request *entities.UpdateCustomerRequest) (*entities.Customer, error)
	DeleteCustomer(id int) error
}

type customerUsecase struct {
	customerRepo repositories.CustomerRepository
	config       *config.Config
}

func NewCustomerUsecase(customerRepo repositories.CustomerRepository, config *config.Config) CustomerUsecase {
	return &customerUsecase{
		customerRepo: customerRepo,
		config:       config,
	}
}

func (u *customerUsecase) CreateCustomer(request *entities.CreateCustomerRequest) (*entities.Customer, error) {
	// Validate required fields based on customer type
	if request.CustomerType == entities.CustomerTypeCompany && request.CompanyName == "" {
		return nil, fmt.Errorf("company name is required for company customers")
	}

	// Check if email already exists (if provided)
	if request.Email != "" {
		// Note: You might want to add a GetByEmail method to check uniqueness
	}

	// Create customer entity
	customer := &entities.Customer{
		FullName:     request.FullName,
		Email:        request.Email,
		Phone:        request.Phone,
		Address:      request.Address,
		City:         request.City,
		IDNumber:     request.IDNumber,
		CustomerType: request.CustomerType,
		CompanyName:  &request.CompanyName,
		TaxNumber:    request.TaxNumber,
		Notes:        request.Notes,
		IsActive:     true,
	}

	// Save to database
	if err := u.customerRepo.Create(customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

func (u *customerUsecase) GetCustomerByID(id int) (*entities.Customer, error) {
	customer, err := u.customerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	return customer, nil
}

func (u *customerUsecase) GetCustomers(page, limit int) ([]entities.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	customers, err := u.customerRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get customers: %w", err)
	}

	// Get total count
	total, err := u.customerRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return customers, total, nil
}

func (u *customerUsecase) SearchCustomers(request *entities.CustomerSearchRequest) ([]entities.Customer, int, error) {
	customers, total, err := u.customerRepo.Search(request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}

	return customers, total, nil
}

func (u *customerUsecase) UpdateCustomer(id int, request *entities.UpdateCustomerRequest) (*entities.Customer, error) {
	// Get existing customer
	existingCustomer, err := u.customerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Validate customer type specific fields
	if request.CustomerType != "" && request.CustomerType == entities.CustomerTypeCompany && request.CompanyName == "" {
		return nil, fmt.Errorf("company name is required for company customers")
	}

	// Prepare update entity
	updateCustomer := &entities.Customer{
		FullName:     request.FullName,
		Email:        request.Email,
		Phone:        request.Phone,
		Address:      request.Address,
		City:         request.City,
		IDNumber:     request.IDNumber,
		CustomerType: request.CustomerType,
		CompanyName:  &request.CompanyName,
		TaxNumber:    request.TaxNumber,
		Notes:        request.Notes,
	}

	// Set is_active if provided
	if request.IsActive != nil {
		updateCustomer.IsActive = *request.IsActive
	} else {
		updateCustomer.IsActive = existingCustomer.IsActive
	}

	// Update customer
	if err := u.customerRepo.Update(id, updateCustomer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Get updated customer
	updatedCustomer, err := u.customerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated customer: %w", err)
	}

	return updatedCustomer, nil
}

func (u *customerUsecase) DeleteCustomer(id int) error {
	// Check if customer exists
	_, err := u.customerRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Soft delete customer
	if err := u.customerRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}
