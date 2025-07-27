package entities

import (
	"time"
)

type TransmissionType string
type FuelType string
type ConditionStatus string
type AvailabilityStatus string
type PhotoType string

const (
	TransmissionManual    TransmissionType = "manual"
	TransmissionAutomatic TransmissionType = "automatic"
	TransmissionCVT       TransmissionType = "cvt"

	FuelGasoline FuelType = "gasoline"
	FuelDiesel   FuelType = "diesel"
	FuelElectric FuelType = "electric"
	FuelHybrid   FuelType = "hybrid"

	ConditionExcellent ConditionStatus = "excellent"
	ConditionGood      ConditionStatus = "good"
	ConditionFair      ConditionStatus = "fair"
	ConditionPoor      ConditionStatus = "poor"

	AvailabilityAvailable   AvailabilityStatus = "available"
	AvailabilitySold        AvailabilityStatus = "sold"
	AvailabilityReserved    AvailabilityStatus = "reserved"
	AvailabilityMaintenance AvailabilityStatus = "maintenance"

	PhotoExterior PhotoType = "exterior"
	PhotoInterior PhotoType = "interior"
	PhotoEngine   PhotoType = "engine"
	PhotoDocument PhotoType = "document"
)

type VehicleCategory struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Vehicle struct {
	ID                 int                `json:"id" db:"id"`
	VehicleCode        string             `json:"vehicle_code" db:"vehicle_code"`
	CategoryID         *int               `json:"category_id" db:"category_id"`
	Category           *VehicleCategory   `json:"category,omitempty"`
	Brand              string             `json:"brand" db:"brand"`
	Model              string             `json:"model" db:"model"`
	Year               int                `json:"year" db:"year"`
	Color              string             `json:"color" db:"color"`
	EngineType         string             `json:"engine_type" db:"engine_type"`
	Transmission       TransmissionType   `json:"transmission" db:"transmission"`
	FuelType           FuelType           `json:"fuel_type" db:"fuel_type"`
	ChassisNumber      string             `json:"chassis_number" db:"chassis_number"`
	EngineNumber       string             `json:"engine_number" db:"engine_number"`
	LicensePlate       string             `json:"license_plate" db:"license_plate"`
	PurchasePrice      float64            `json:"purchase_price" db:"purchase_price"`
	SellingPrice       float64            `json:"selling_price" db:"selling_price"`
	ConditionStatus    ConditionStatus    `json:"condition_status" db:"condition_status"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status" db:"availability_status"`
	Location           string             `json:"location" db:"location"`
	Mileage            int                `json:"mileage" db:"mileage"`
	Description        string             `json:"description" db:"description"`
	Features           []string           `json:"features" db:"features"`
	Photos             []VehiclePhoto     `json:"photos,omitempty"`
	Status             string             `json:"status" db:"status"`
	IsActive           bool               `json:"is_active" db:"is_active"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
}

type VehiclePhoto struct {
	ID         int        `json:"id" db:"id"`
	VehicleID  int        `json:"vehicle_id" db:"vehicle_id"`
	PhotoURL   string     `json:"photo_url" db:"photo_url"`
	PhotoType  PhotoType  `json:"photo_type" db:"photo_type"`
	IsPrimary  bool       `json:"is_primary" db:"is_primary"`
	SortOrder  int        `json:"sort_order" db:"sort_order"`
	UploadedBy int        `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateVehicleRequest struct {
	CategoryID         *int               `json:"category_id" binding:"omitempty"`
	Brand              string             `json:"brand" binding:"required,min=1,max=50"`
	Model              string             `json:"model" binding:"required,min=1,max=100"`
	Year               int                `json:"year" binding:"required,min=1900,max=2030"`
	Color              string             `json:"color" binding:"required,min=1,max=30"`
	EngineType         string             `json:"engine_type" binding:"omitempty,max=50"`
	Transmission       TransmissionType   `json:"transmission" binding:"required,oneof=manual automatic cvt"`
	FuelType           FuelType           `json:"fuel_type" binding:"required,oneof=gasoline diesel electric hybrid"`
	ChassisNumber      string             `json:"chassis_number" binding:"omitempty,max=50"`
	EngineNumber       string             `json:"engine_number" binding:"omitempty,max=50"`
	LicensePlate       string             `json:"license_plate" binding:"omitempty,max=15"`
	PurchasePrice      float64            `json:"purchase_price" binding:"omitempty,min=0"`
	SellingPrice       float64            `json:"selling_price" binding:"omitempty,min=0"`
	ConditionStatus    ConditionStatus    `json:"condition_status" binding:"required,oneof=excellent good fair poor"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status" binding:"omitempty,oneof=available sold reserved maintenance"`
	Location           string             `json:"location" binding:"omitempty,max=100"`
	Mileage            int                `json:"mileage" binding:"omitempty,min=0"`
	Description        string             `json:"description" binding:"omitempty"`
	Features           []string           `json:"features" binding:"omitempty"`
}

type UpdateVehicleRequest struct {
	CategoryID         *int               `json:"category_id" binding:"omitempty"`
	Brand              string             `json:"brand" binding:"omitempty,min=1,max=50"`
	Model              string             `json:"model" binding:"omitempty,min=1,max=100"`
	Year               int                `json:"year" binding:"omitempty,min=1900,max=2030"`
	Color              string             `json:"color" binding:"omitempty,min=1,max=30"`
	EngineType         string             `json:"engine_type" binding:"omitempty,max=50"`
	Transmission       TransmissionType   `json:"transmission" binding:"omitempty,oneof=manual automatic cvt"`
	FuelType           FuelType           `json:"fuel_type" binding:"omitempty,oneof=gasoline diesel electric hybrid"`
	ChassisNumber      string             `json:"chassis_number" binding:"omitempty,max=50"`
	EngineNumber       string             `json:"engine_number" binding:"omitempty,max=50"`
	LicensePlate       string             `json:"license_plate" binding:"omitempty,max=15"`
	PurchasePrice      *float64           `json:"purchase_price" binding:"omitempty,min=0"`
	SellingPrice       *float64           `json:"selling_price" binding:"omitempty,min=0"`
	ConditionStatus    ConditionStatus    `json:"condition_status" binding:"omitempty,oneof=excellent good fair poor"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status" binding:"omitempty,oneof=available sold reserved maintenance"`
	Location           string             `json:"location" binding:"omitempty,max=100"`
	Mileage            *int               `json:"mileage" binding:"omitempty,min=0"`
	Description        string             `json:"description" binding:"omitempty"`
	Features           []string           `json:"features" binding:"omitempty"`
	IsActive           *bool              `json:"is_active" binding:"omitempty"`
}

type VehicleSearchRequest struct {
	Query              string             `json:"query" form:"query"`
	Brand              string             `json:"brand" form:"brand"`
	Model              string             `json:"model" form:"model"`
	CategoryID         *int               `json:"category_id" form:"category_id"`
	YearFrom           int                `json:"year_from" form:"year_from"`
	YearTo             int                `json:"year_to" form:"year_to"`
	PriceFrom          float64            `json:"price_from" form:"price_from"`
	PriceTo            float64            `json:"price_to" form:"price_to"`
	Transmission       TransmissionType   `json:"transmission" form:"transmission"`
	FuelType           FuelType           `json:"fuel_type" form:"fuel_type"`
	ConditionStatus    ConditionStatus    `json:"condition_status" form:"condition_status"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status" form:"availability_status"`
	Color              string             `json:"color" form:"color"`
	Page               int                `json:"page" form:"page"`
	Limit              int                `json:"limit" form:"limit"`
	SortBy             string             `json:"sort_by" form:"sort_by"`
	SortOrder          string             `json:"sort_order" form:"sort_order"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"omitempty"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"omitempty"`
	IsActive    *bool  `json:"is_active" binding:"omitempty"`
}

type VehicleFilter struct {
	Brand              string  `json:"brand"`
	Model              string  `json:"model"`
	YearFrom           int     `json:"year_from"`
	YearTo             int     `json:"year_to"`
	PriceFrom          float64 `json:"price_from"`
	PriceTo            float64 `json:"price_to"`
	FuelType           string  `json:"fuel_type"`
	Transmission       string  `json:"transmission"`
	ConditionStatus    string  `json:"condition_status"`    // NEW
	AvailabilityStatus string  `json:"availability_status"` // NEW
	Location           string  `json:"location"`
	CategoryID         int     `json:"category_id"`
	IsActive           *bool   `json:"is_active"`
}
