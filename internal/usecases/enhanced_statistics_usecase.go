package usecases

import (
	"time"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"
)

type EnhancedStatisticsUsecase interface {
	GetEnhancedDashboardStats() (*EnhancedDashboardStats, error)
}

type EnhancedDashboardStats struct {
	TotalVehicles     int                      `json:"total_vehicles"`
	AvailableVehicles int                      `json:"available_vehicles"`
	SoldVehicles      int                      `json:"sold_vehicles"`
	TotalCustomers    int                      `json:"total_customers"`
	TotalUsers        int                      `json:"total_users"`
	RecentActivities  []EnhancedRecentActivity `json:"recent_activities"`
	VehicleStats      EnhancedVehicleStats     `json:"vehicle_stats"`
	CustomerStats     EnhancedCustomerStats    `json:"customer_stats"`
	MonthlyTrends     map[string]any           `json:"monthly_trends"`
}

type EnhancedVehicleStats struct {
	StatusBreakdown       map[string]int `json:"status_breakdown"`
	FuelTypeBreakdown     map[string]int `json:"fuel_type_breakdown"`
	TransmissionBreakdown map[string]int `json:"transmission_breakdown"`
	BrandBreakdown        map[string]int `json:"brand_breakdown"`
	AveragePurchasePrice  float64        `json:"average_purchase_price"`
	AverageSellingPrice   float64        `json:"average_selling_price"`
	TotalInventoryValue   float64        `json:"total_inventory_value"`
}

type EnhancedCustomerStats struct {
	TypeBreakdown       map[string]int `json:"type_breakdown"`
	RecentRegistrations int            `json:"recent_registrations"`
	ActiveCustomers     int            `json:"active_customers"`
	MonthlyGrowth       float64        `json:"monthly_growth"`
}

type EnhancedRecentActivity struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type enhancedStatisticsUsecase struct {
	vehicleRepo  repositories.VehicleRepository
	customerRepo repositories.CustomerRepository
	userRepo     repositories.UserRepository
	config       *config.Config
}

func NewEnhancedStatisticsUsecase(
	vehicleRepo repositories.VehicleRepository,
	customerRepo repositories.CustomerRepository,
	userRepo repositories.UserRepository,
	config *config.Config,
) EnhancedStatisticsUsecase {
	return &enhancedStatisticsUsecase{
		vehicleRepo:  vehicleRepo,
		customerRepo: customerRepo,
		userRepo:     userRepo,
		config:       config,
	}
}

func (u *enhancedStatisticsUsecase) GetEnhancedDashboardStats() (*EnhancedDashboardStats, error) {
	// Get all vehicles
	vehicles, err := u.vehicleRepo.GetAll(0, 0)
	if err != nil {
		return nil, err
	}

	// Get all customers
	customers, err := u.customerRepo.GetAll(0, 0)
	if err != nil {
		return nil, err
	}

	// Get all users
	users, err := u.userRepo.GetAll(0, 0)
	if err != nil {
		return nil, err
	}

	// Calculate vehicle statistics
	vehicleStats := u.calculateEnhancedVehicleStats(vehicles)

	// Calculate customer statistics
	customerStats := u.calculateEnhancedCustomerStats(customers)

	// Get recent activities
	recentActivities := u.getEnhancedRecentActivities()

	stats := &EnhancedDashboardStats{
		TotalVehicles:     len(vehicles),
		AvailableVehicles: u.countVehiclesByStatus(vehicles, "available"),
		SoldVehicles:      u.countVehiclesByStatus(vehicles, "sold"),
		TotalCustomers:    len(customers),
		TotalUsers:        len(users),
		RecentActivities:  recentActivities,
		VehicleStats:      vehicleStats,
		CustomerStats:     customerStats,
		MonthlyTrends:     u.calculateEnhancedMonthlyTrends(),
	}

	return stats, nil
}

func (u *enhancedStatisticsUsecase) calculateEnhancedVehicleStats(vehicles []entities.Vehicle) EnhancedVehicleStats {
	// Initialize breakdown maps
	statusBreakdown := make(map[string]int)
	fuelTypeBreakdown := make(map[string]int)
	transmissionBreakdown := make(map[string]int)
	brandBreakdown := make(map[string]int)

	var totalPurchasePrice, totalSellingPrice, totalInventoryValue float64

	for _, vehicle := range vehicles {
		// Fix: Convert enum types to string by casting
		if vehicle.AvailabilityStatus != "" {
			statusBreakdown[string(vehicle.AvailabilityStatus)]++
		}

		if vehicle.FuelType != "" {
			fuelTypeBreakdown[string(vehicle.FuelType)]++
		}

		if vehicle.Transmission != "" {
			transmissionBreakdown[string(vehicle.Transmission)]++
		}

		// Skip ConditionStatus if it's not defined in Vehicle struct
		// conditionBreakdown[string(vehicle.ConditionStatus)]++

		if vehicle.Brand != "" {
			brandBreakdown[vehicle.Brand]++
		}

		totalPurchasePrice += vehicle.PurchasePrice
		totalSellingPrice += vehicle.SellingPrice

		// Calculate inventory value for available vehicles
		if string(vehicle.AvailabilityStatus) == "available" || string(vehicle.AvailabilityStatus) == "reserved" {
			totalInventoryValue += vehicle.SellingPrice
		}
	}

	vehicleCount := len(vehicles)
	var avgPurchasePrice, avgSellingPrice float64
	if vehicleCount > 0 {
		avgPurchasePrice = totalPurchasePrice / float64(vehicleCount)
		avgSellingPrice = totalSellingPrice / float64(vehicleCount)
	}

	return EnhancedVehicleStats{
		StatusBreakdown:       statusBreakdown,
		FuelTypeBreakdown:     fuelTypeBreakdown,
		TransmissionBreakdown: transmissionBreakdown,
		BrandBreakdown:        brandBreakdown,
		AveragePurchasePrice:  avgPurchasePrice,
		AverageSellingPrice:   avgSellingPrice,
		TotalInventoryValue:   totalInventoryValue,
	}
}

func (u *enhancedStatisticsUsecase) calculateEnhancedCustomerStats(customers []entities.Customer) EnhancedCustomerStats {
	typeBreakdown := make(map[string]int)
	var recentRegistrations int

	// Calculate recent registrations (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, customer := range customers {
		// Convert enum to string
		if customer.CustomerType != "" {
			typeBreakdown[string(customer.CustomerType)]++
		}

		if customer.CreatedAt.After(thirtyDaysAgo) {
			recentRegistrations++
		}
	}

	return EnhancedCustomerStats{
		TypeBreakdown:       typeBreakdown,
		RecentRegistrations: recentRegistrations,
		ActiveCustomers:     len(customers),
		MonthlyGrowth:       5.2, // Mock value
	}
}

func (u *enhancedStatisticsUsecase) countVehiclesByStatus(vehicles []entities.Vehicle, status string) int {
	count := 0
	for _, vehicle := range vehicles {
		if string(vehicle.AvailabilityStatus) == status {
			count++
		}
	}
	return count
}

func (u *enhancedStatisticsUsecase) getEnhancedRecentActivities() []EnhancedRecentActivity {
	return []EnhancedRecentActivity{
		{
			ID:          1,
			Type:        "vehicle_added",
			Description: "New Toyota Camry 2023 added to inventory",
			UserName:    "Admin User",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          2,
			Type:        "customer_registered",
			Description: "New customer John Doe registered",
			UserName:    "Cashier User",
			CreatedAt:   time.Now().Add(-4 * time.Hour),
		},
	}
}

func (u *enhancedStatisticsUsecase) calculateEnhancedMonthlyTrends() map[string]any {
	return map[string]any{
		"sales_trend": []map[string]any{
			{"month": "Jan", "sales": 12, "revenue": 2500000000},
			{"month": "Feb", "sales": 15, "revenue": 3200000000},
		},
	}
}
