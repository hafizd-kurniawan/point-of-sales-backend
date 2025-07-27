package repositories

import (
	"database/sql"
	"fmt"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
)

type vehicleCategoryRepositoryImpl struct {
	db *sqlx.DB
}

func NewVehicleCategoryRepository(db *sqlx.DB) repositories.VehicleCategoryRepository {
	return &vehicleCategoryRepositoryImpl{db: db}
}

func (r *vehicleCategoryRepositoryImpl) Create(category *entities.VehicleCategory) error {
	query := `
		INSERT INTO vehicle_categories (name, description, is_active)
		VALUES (:name, :description, :is_active)
		RETURNING id, created_at, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(category, category)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *vehicleCategoryRepositoryImpl) GetByID(id int) (*entities.VehicleCategory, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at, deleted_at
		FROM vehicle_categories 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var category entities.VehicleCategory
	err := r.db.Get(&category, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

func (r *vehicleCategoryRepositoryImpl) GetAll() ([]entities.VehicleCategory, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at, deleted_at
		FROM vehicle_categories 
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY name ASC
	`

	var categories []entities.VehicleCategory
	err := r.db.Select(&categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (r *vehicleCategoryRepositoryImpl) Update(id int, category *entities.VehicleCategory) error {
	query := `
		UPDATE vehicle_categories 
		SET name = COALESCE(NULLIF(:name, ''), name),
		    description = COALESCE(NULLIF(:description, ''), description),
		    is_active = COALESCE(:is_active, is_active),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND deleted_at IS NULL
		RETURNING updated_at
	`

	category.ID = id
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(&category.UpdatedAt, category)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("category not found")
		}
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (r *vehicleCategoryRepositoryImpl) Delete(id int) error {
	query := `
		UPDATE vehicle_categories 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
