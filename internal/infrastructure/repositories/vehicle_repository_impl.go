package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type vehicleRepositoryImpl struct {
	db *sqlx.DB
}

func NewVehicleRepository(db *sqlx.DB) repositories.VehicleRepository {
	return &vehicleRepositoryImpl{db: db}
}

func (r *vehicleRepositoryImpl) Create(vehicle *entities.Vehicle) error {
	// Generate vehicle code if not provided
	if vehicle.VehicleCode == "" {
		code, err := r.GenerateVehicleCode()
		if err != nil {
			return fmt.Errorf("failed to generate vehicle code: %w", err)
		}
		vehicle.VehicleCode = code
	}

	query := `
		INSERT INTO vehicles (vehicle_code, category_id, brand, model, year, color, engine_type,
		                     transmission, fuel_type, chassis_number, engine_number, license_plate,
		                     purchase_price, selling_price, condition_status, availability_status,
		                     location, mileage, description, features, is_active)
		VALUES (:vehicle_code, :category_id, :brand, :model, :year, :color, :engine_type,
		        :transmission, :fuel_type, :chassis_number, :engine_number, :license_plate,
		        :purchase_price, :selling_price, :condition_status, :availability_status,
		        :location, :mileage, :description, :features, :is_active)
		RETURNING id, created_at, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(vehicle, vehicle)
	if err != nil {
		return fmt.Errorf("failed to create vehicle: %w", err)
	}

	return nil
}

func (r *vehicleRepositoryImpl) GetByID(id int) (*entities.Vehicle, error) {
	query := `
		SELECT v.id, v.vehicle_code, v.category_id, v.brand, v.model, v.year, v.color,
		       v.engine_type, v.transmission, v.fuel_type, v.chassis_number, v.engine_number,
		       v.license_plate, v.purchase_price, v.selling_price, v.condition_status,
		       v.availability_status, v.location, v.mileage, v.description, v.features,
		       v.is_active, v.created_at, v.updated_at, v.deleted_at,
		       c.id as "category.id", c.name as "category.name", c.description as "category.description"
		FROM vehicles v
		LEFT JOIN vehicle_categories c ON v.category_id = c.id AND c.deleted_at IS NULL
		WHERE v.id = $1 AND v.deleted_at IS NULL
	`

	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("vehicle not found")
	}

	var vehicle entities.Vehicle
	var category entities.VehicleCategory
	var categoryIDPtr *int

	err = rows.Scan(
		&vehicle.ID, &vehicle.VehicleCode, &vehicle.CategoryID, &vehicle.Brand,
		&vehicle.Model, &vehicle.Year, &vehicle.Color, &vehicle.EngineType,
		&vehicle.Transmission, &vehicle.FuelType, &vehicle.ChassisNumber,
		&vehicle.EngineNumber, &vehicle.LicensePlate, &vehicle.PurchasePrice,
		&vehicle.SellingPrice, &vehicle.ConditionStatus, &vehicle.AvailabilityStatus,
		&vehicle.Location, &vehicle.Mileage, &vehicle.Description,
		pq.Array(&vehicle.Features), &vehicle.IsActive,
		&vehicle.CreatedAt, &vehicle.UpdatedAt, &vehicle.DeletedAt,
		&categoryIDPtr, &category.Name, &category.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan vehicle: %w", err)
	}

	if categoryIDPtr != nil {
		category.ID = *categoryIDPtr
		vehicle.Category = &category
	}

	return &vehicle, nil
}

func (r *vehicleRepositoryImpl) GetByVehicleCode(vehicleCode string) (*entities.Vehicle, error) {
	query := `
		SELECT v.id, v.vehicle_code, v.category_id, v.brand, v.model, v.year, v.color,
		       v.engine_type, v.transmission, v.fuel_type, v.chassis_number, v.engine_number,
		       v.license_plate, v.purchase_price, v.selling_price, v.condition_status,
		       v.availability_status, v.location, v.mileage, v.description, v.features,
		       v.is_active, v.created_at, v.updated_at, v.deleted_at,
		       c.id as "category.id", c.name as "category.name", c.description as "category.description"
		FROM vehicles v
		LEFT JOIN vehicle_categories c ON v.category_id = c.id AND c.deleted_at IS NULL
		WHERE v.vehicle_code = $1 AND v.deleted_at IS NULL
	`

	rows, err := r.db.Query(query, vehicleCode)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("vehicle not found")
	}

	var vehicle entities.Vehicle
	var category entities.VehicleCategory
	var categoryIDPtr *int

	err = rows.Scan(
		&vehicle.ID, &vehicle.VehicleCode, &vehicle.CategoryID, &vehicle.Brand,
		&vehicle.Model, &vehicle.Year, &vehicle.Color, &vehicle.EngineType,
		&vehicle.Transmission, &vehicle.FuelType, &vehicle.ChassisNumber,
		&vehicle.EngineNumber, &vehicle.LicensePlate, &vehicle.PurchasePrice,
		&vehicle.SellingPrice, &vehicle.ConditionStatus, &vehicle.AvailabilityStatus,
		&vehicle.Location, &vehicle.Mileage, &vehicle.Description,
		pq.Array(&vehicle.Features), &vehicle.IsActive,
		&vehicle.CreatedAt, &vehicle.UpdatedAt, &vehicle.DeletedAt,
		&categoryIDPtr, &category.Name, &category.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan vehicle: %w", err)
	}

	if categoryIDPtr != nil {
		category.ID = *categoryIDPtr
		vehicle.Category = &category
	}

	return &vehicle, nil
}

func (r *vehicleRepositoryImpl) GetAll(limit, offset int) ([]entities.Vehicle, error) {
	query := `
		SELECT v.id, v.vehicle_code, v.category_id, v.brand, v.model, v.year, v.color,
		       v.engine_type, v.transmission, v.fuel_type, v.chassis_number, v.engine_number,
		       v.license_plate, v.purchase_price, v.selling_price, v.condition_status,
		       v.availability_status, v.location, v.mileage, v.description, v.features,
		       v.is_active, v.created_at, v.updated_at, v.deleted_at,
		       c.id as "category.id", c.name as "category.name", c.description as "category.description"
		FROM vehicles v
		LEFT JOIN vehicle_categories c ON v.category_id = c.id AND c.deleted_at IS NULL
		WHERE v.deleted_at IS NULL
		ORDER BY v.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var vehicles []entities.Vehicle
	for rows.Next() {
		var vehicle entities.Vehicle
		var category entities.VehicleCategory
		var categoryIDPtr *int

		err = rows.Scan(
			&vehicle.ID, &vehicle.VehicleCode, &vehicle.CategoryID, &vehicle.Brand,
			&vehicle.Model, &vehicle.Year, &vehicle.Color, &vehicle.EngineType,
			&vehicle.Transmission, &vehicle.FuelType, &vehicle.ChassisNumber,
			&vehicle.EngineNumber, &vehicle.LicensePlate, &vehicle.PurchasePrice,
			&vehicle.SellingPrice, &vehicle.ConditionStatus, &vehicle.AvailabilityStatus,
			&vehicle.Location, &vehicle.Mileage, &vehicle.Description,
			pq.Array(&vehicle.Features), &vehicle.IsActive,
			&vehicle.CreatedAt, &vehicle.UpdatedAt, &vehicle.DeletedAt,
			&categoryIDPtr, &category.Name, &category.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle: %w", err)
		}

		if categoryIDPtr != nil {
			category.ID = *categoryIDPtr
			vehicle.Category = &category
		}

		vehicles = append(vehicles, vehicle)
	}

	return vehicles, nil
}

func (r *vehicleRepositoryImpl) Search(req *entities.VehicleSearchRequest) ([]entities.Vehicle, int, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Base where condition
	whereConditions = append(whereConditions, "v.deleted_at IS NULL")

	// Search query
	if req.Query != "" {
		whereConditions = append(whereConditions,
			fmt.Sprintf("(v.brand ILIKE $%d OR v.model ILIKE $%d OR v.vehicle_code ILIKE $%d OR v.license_plate ILIKE $%d)",
				argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+req.Query+"%")
		argIndex++
	}

	// Brand filter
	if req.Brand != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.brand ILIKE $%d", argIndex))
		args = append(args, "%"+req.Brand+"%")
		argIndex++
	}

	// Model filter
	if req.Model != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.model ILIKE $%d", argIndex))
		args = append(args, "%"+req.Model+"%")
		argIndex++
	}

	// Category filter
	if req.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("v.category_id = $%d", argIndex))
		args = append(args, *req.CategoryID)
		argIndex++
	}

	// Year range filter
	if req.YearFrom > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("v.year >= $%d", argIndex))
		args = append(args, req.YearFrom)
		argIndex++
	}
	if req.YearTo > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("v.year <= $%d", argIndex))
		args = append(args, req.YearTo)
		argIndex++
	}

	// Price range filter
	if req.PriceFrom > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("v.selling_price >= $%d", argIndex))
		args = append(args, req.PriceFrom)
		argIndex++
	}
	if req.PriceTo > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("v.selling_price <= $%d", argIndex))
		args = append(args, req.PriceTo)
		argIndex++
	}

	// Transmission filter
	if req.Transmission != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.transmission = $%d", argIndex))
		args = append(args, req.Transmission)
		argIndex++
	}

	// Fuel type filter
	if req.FuelType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.fuel_type = $%d", argIndex))
		args = append(args, req.FuelType)
		argIndex++
	}

	// Condition status filter
	if req.ConditionStatus != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.condition_status = $%d", argIndex))
		args = append(args, req.ConditionStatus)
		argIndex++
	}

	// Availability status filter
	if req.AvailabilityStatus != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.availability_status = $%d", argIndex))
		args = append(args, req.AvailabilityStatus)
		argIndex++
	}

	// Color filter
	if req.Color != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("v.color ILIKE $%d", argIndex))
		args = append(args, "%"+req.Color+"%")
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM vehicles v
		LEFT JOIN vehicle_categories c ON v.category_id = c.id AND c.deleted_at IS NULL
		WHERE %s`, whereClause)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count vehicles: %w", err)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "v.created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "DESC"
	}

	offset := (req.Page - 1) * req.Limit

	// Build order clause
	orderClause := fmt.Sprintf("ORDER BY %s %s", req.SortBy, req.SortOrder)

	// Search query
	searchQuery := fmt.Sprintf(`
		SELECT v.id, v.vehicle_code, v.category_id, v.brand, v.model, v.year, v.color,
		       v.engine_type, v.transmission, v.fuel_type, v.chassis_number, v.engine_number,
		       v.license_plate, v.purchase_price, v.selling_price, v.condition_status,
		       v.availability_status, v.location, v.mileage, v.description, v.features,
		       v.is_active, v.created_at, v.updated_at, v.deleted_at,
		       c.id as "category.id", c.name as "category.name", c.description as "category.description"
		FROM vehicles v
		LEFT JOIN vehicle_categories c ON v.category_id = c.id AND c.deleted_at IS NULL
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(searchQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []entities.Vehicle
	for rows.Next() {
		var vehicle entities.Vehicle
		var category entities.VehicleCategory
		var categoryIDPtr *int

		err = rows.Scan(
			&vehicle.ID, &vehicle.VehicleCode, &vehicle.CategoryID, &vehicle.Brand,
			&vehicle.Model, &vehicle.Year, &vehicle.Color, &vehicle.EngineType,
			&vehicle.Transmission, &vehicle.FuelType, &vehicle.ChassisNumber,
			&vehicle.EngineNumber, &vehicle.LicensePlate, &vehicle.PurchasePrice,
			&vehicle.SellingPrice, &vehicle.ConditionStatus, &vehicle.AvailabilityStatus,
			&vehicle.Location, &vehicle.Mileage, &vehicle.Description,
			pq.Array(&vehicle.Features), &vehicle.IsActive,
			&vehicle.CreatedAt, &vehicle.UpdatedAt, &vehicle.DeletedAt,
			&categoryIDPtr, &category.Name, &category.Description,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan vehicle: %w", err)
		}

		if categoryIDPtr != nil {
			category.ID = *categoryIDPtr
			vehicle.Category = &category
		}

		vehicles = append(vehicles, vehicle)
	}

	return vehicles, total, nil
}

func (r *vehicleRepositoryImpl) Update(id int, vehicle *entities.Vehicle) error {
	query := `
		UPDATE vehicles 
		SET category_id = COALESCE(:category_id, category_id),
		    brand = COALESCE(NULLIF(:brand, ''), brand),
		    model = COALESCE(NULLIF(:model, ''), model),
		    year = COALESCE(:year, year),
		    color = COALESCE(NULLIF(:color, ''), color),
		    engine_type = COALESCE(NULLIF(:engine_type, ''), engine_type),
		    transmission = COALESCE(NULLIF(:transmission, ''), transmission),
		    fuel_type = COALESCE(NULLIF(:fuel_type, ''), fuel_type),
		    chassis_number = COALESCE(NULLIF(:chassis_number, ''), chassis_number),
		    engine_number = COALESCE(NULLIF(:engine_number, ''), engine_number),
		    license_plate = COALESCE(NULLIF(:license_plate, ''), license_plate),
		    purchase_price = COALESCE(:purchase_price, purchase_price),
		    selling_price = COALESCE(:selling_price, selling_price),
		    condition_status = COALESCE(NULLIF(:condition_status, ''), condition_status),
		    availability_status = COALESCE(NULLIF(:availability_status, ''), availability_status),
		    location = COALESCE(NULLIF(:location, ''), location),
		    mileage = COALESCE(:mileage, mileage),
		    description = COALESCE(NULLIF(:description, ''), description),
		    features = COALESCE(:features, features),
		    is_active = COALESCE(:is_active, is_active),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND deleted_at IS NULL
		RETURNING updated_at
	`

	vehicle.ID = id
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(&vehicle.UpdatedAt, vehicle)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("vehicle not found")
		}
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	return nil
}

func (r *vehicleRepositoryImpl) Delete(id int) error {
	query := `
		UPDATE vehicles 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found")
	}

	return nil
}

func (r *vehicleRepositoryImpl) Count() (int, error) {
	query := `SELECT COUNT(*) FROM vehicles WHERE deleted_at IS NULL`

	var count int
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count vehicles: %w", err)
	}

	return count, nil
}

func (r *vehicleRepositoryImpl) GenerateVehicleCode() (string, error) {
	query := `SELECT generate_vehicle_code()`

	var code string
	err := r.db.Get(&code, query)
	if err != nil {
		return "", fmt.Errorf("failed to generate vehicle code: %w", err)
	}

	return code, nil
}
