package usecases

import (
	"fmt"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/repositories"
)

type StatisticsUsecase interface {
	GetDashboardStats() (*DashboardStats, error)
	GetVehicleStats() (*VehicleStats, error)
	GetCustomerStats() (*CustomerStats, error)
}

type DashboardStats struct {
	TotalVehicles      int              `json:"total_vehicles"`
	AvailableVehicles  int              `json:"available_vehicles"`
	SoldVehicles       int              `json:"sold_vehicles"`
	TotalCustomers     int              `json:"total_customers"`
	ActiveCustomers    int              `json:"active_customers"`
	VehiclesByCategory map[string]int   `json:"vehicles_by_category"`
	CustomersByType    map[string]int   `json:"customers_by_type"`
	RecentActivity     []RecentActivity `json:"recent_activity"`
}

type VehicleStats struct {
	TotalVehicles  int            `json:"total_vehicles"`
	ByStatus       map[string]int `json:"by_status"`
	ByBrand        map[string]int `json:"by_brand"`
	ByFuelType     map[string]int `json:"by_fuel_type"`
	ByTransmission map[string]int `json:"by_transmission"`
	ByCondition    map[string]int `json:"by_condition"`
	PriceRanges    map[string]int `json:"price_ranges"`
}

type CustomerStats struct {
	TotalCustomers  int            `json:"total_customers"`
	ByType          map[string]int `json:"by_type"`
	ByCity          map[string]int `json:"by_city"`
	ActiveCustomers int            `json:"active_customers"`
}

type RecentActivity struct {
	ID          int    `json:"id"`
	Type        string `json:"type"` // vehicle_added, customer_added, etc
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
	UserName    string `json:"user_name"`
}

type statisticsUsecase struct {
	vehicleRepo  repositories.VehicleRepository
	customerRepo repositories.CustomerRepository
	userRepo     repositories.UserRepository
	config       *config.Config
}

func NewStatisticsUsecase(
	vehicleRepo repositories.VehicleRepository,
	customerRepo repositories.CustomerRepository,
	userRepo repositories.UserRepository,
	config *config.Config,
) StatisticsUsecase {
	return &statisticsUsecase{
		vehicleRepo:  vehicleRepo,
		customerRepo: customerRepo,
		userRepo:     userRepo,
		config:       config,
	}
}

func (u *statisticsUsecase) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{
		VehiclesByCategory: make(map[string]int),
		CustomersByType:    make(map[string]int),
		RecentActivity:     []RecentActivity{},
	}

	// Get vehicle counts
	totalVehicles, err := u.vehicleRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count vehicles: %w", err)
	}
	stats.TotalVehicles = totalVehicles

	// Get customer counts
	totalCustomers, err := u.customerRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}
	stats.TotalCustomers = totalCustomers

	// TODO: Add more sophisticated queries for:
	// - Available/sold vehicles count
	// - Active customers count
	// - Vehicles by category
	// - Customers by type
	// - Recent activity

	return stats, nil
}

func (u *statisticsUsecase) GetVehicleStats() (*VehicleStats, error) {
	stats := &VehicleStats{
		ByStatus:       make(map[string]int),
		ByBrand:        make(map[string]int),
		ByFuelType:     make(map[string]int),
		ByTransmission: make(map[string]int),
		ByCondition:    make(map[string]int),
		PriceRanges:    make(map[string]int),
	}

	totalVehicles, err := u.vehicleRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count vehicles: %w", err)
	}
	stats.TotalVehicles = totalVehicles

	// TODO: Add detailed vehicle statistics queries

	return stats, nil
}

func (u *statisticsUsecase) GetCustomerStats() (*CustomerStats, error) {
	stats := &CustomerStats{
		ByType: make(map[string]int),
		ByCity: make(map[string]int),
	}

	totalCustomers, err := u.customerRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}
	stats.TotalCustomers = totalCustomers

	// TODO: Add detailed customer statistics queries

	return stats, nil
}
