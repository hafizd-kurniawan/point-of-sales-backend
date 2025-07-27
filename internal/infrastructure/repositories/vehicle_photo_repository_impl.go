package repositories

import (
	"database/sql"
	"fmt"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
)

type vehiclePhotoRepositoryImpl struct {
	db *sqlx.DB
}

func NewVehiclePhotoRepository(db *sqlx.DB) repositories.VehiclePhotoRepository {
	return &vehiclePhotoRepositoryImpl{db: db}
}

func (r *vehiclePhotoRepositoryImpl) Create(photo *entities.VehiclePhoto) error {
	query := `
		INSERT INTO vehicle_photos (vehicle_id, photo_url, photo_type, is_primary, sort_order, uploaded_by)
		VALUES (:vehicle_id, :photo_url, :photo_type, :is_primary, :sort_order, :uploaded_by)
		RETURNING id, created_at, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(photo, photo)
	if err != nil {
		return fmt.Errorf("failed to create photo: %w", err)
	}

	return nil
}

func (r *vehiclePhotoRepositoryImpl) GetByVehicleID(vehicleID int) ([]entities.VehiclePhoto, error) {
	query := `
		SELECT id, vehicle_id, photo_url, photo_type, is_primary, sort_order, uploaded_by,
		       created_at, updated_at, deleted_at
		FROM vehicle_photos 
		WHERE vehicle_id = $1 AND deleted_at IS NULL
		ORDER BY is_primary DESC, sort_order ASC, created_at ASC
	`

	var photos []entities.VehiclePhoto
	err := r.db.Select(&photos, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}

	return photos, nil
}

func (r *vehiclePhotoRepositoryImpl) GetByID(id int) (*entities.VehiclePhoto, error) {
	query := `
		SELECT id, vehicle_id, photo_url, photo_type, is_primary, sort_order, uploaded_by,
		       created_at, updated_at, deleted_at
		FROM vehicle_photos 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var photo entities.VehiclePhoto
	err := r.db.Get(&photo, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("photo not found")
		}
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	return &photo, nil
}

func (r *vehiclePhotoRepositoryImpl) Update(id int, photo *entities.VehiclePhoto) error {
	query := `
		UPDATE vehicle_photos 
		SET photo_url = COALESCE(NULLIF(:photo_url, ''), photo_url),
		    photo_type = COALESCE(NULLIF(:photo_type, ''), photo_type),
		    is_primary = COALESCE(:is_primary, is_primary),
		    sort_order = COALESCE(:sort_order, sort_order),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND deleted_at IS NULL
		RETURNING updated_at
	`

	photo.ID = id
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(&photo.UpdatedAt, photo)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("photo not found")
		}
		return fmt.Errorf("failed to update photo: %w", err)
	}

	return nil
}

func (r *vehiclePhotoRepositoryImpl) Delete(id int) error {
	query := `
		UPDATE vehicle_photos 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("photo not found")
	}

	return nil
}

func (r *vehiclePhotoRepositoryImpl) SetPrimary(vehicleID int, photoID int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove primary flag from all photos of this vehicle
	_, err = tx.Exec(`
		UPDATE vehicle_photos 
		SET is_primary = false 
		WHERE vehicle_id = $1 AND deleted_at IS NULL
	`, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to remove primary flags: %w", err)
	}

	// Set new primary photo
	result, err := tx.Exec(`
		UPDATE vehicle_photos 
		SET is_primary = true 
		WHERE id = $1 AND vehicle_id = $2 AND deleted_at IS NULL
	`, photoID, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to set primary photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("photo not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *vehiclePhotoRepositoryImpl) UpdateSortOrder(photoID int, sortOrder int) error {
	query := `
		UPDATE vehicle_photos 
		SET sort_order = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, sortOrder, photoID)
	if err != nil {
		return fmt.Errorf("failed to update sort order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("photo not found")
	}

	return nil
}
