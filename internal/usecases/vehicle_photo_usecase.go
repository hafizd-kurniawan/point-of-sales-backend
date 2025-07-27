package usecases

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/google/uuid"
)

type VehiclePhotoUsecase interface {
	UploadPhoto(vehicleID int, file *multipart.FileHeader, photoType entities.PhotoType, uploadedBy int) (*entities.VehiclePhoto, error)
	GetVehiclePhotos(vehicleID int) ([]entities.VehiclePhoto, error)
	DeletePhoto(photoID int) error
	SetPrimaryPhoto(vehicleID int, photoID int) error
	UpdatePhotoSortOrder(photoID int, sortOrder int) error
}

type vehiclePhotoUsecase struct {
	photoRepo   repositories.VehiclePhotoRepository
	vehicleRepo repositories.VehicleRepository
	config      *config.Config
}

func NewVehiclePhotoUsecase(
	photoRepo repositories.VehiclePhotoRepository,
	vehicleRepo repositories.VehicleRepository,
	config *config.Config,
) VehiclePhotoUsecase {
	return &vehiclePhotoUsecase{
		photoRepo:   photoRepo,
		vehicleRepo: vehicleRepo,
		config:      config,
	}
}

func (u *vehiclePhotoUsecase) UploadPhoto(vehicleID int, file *multipart.FileHeader, photoType entities.PhotoType, uploadedBy int) (*entities.VehiclePhoto, error) {
	// Verify vehicle exists
	_, err := u.vehicleRepo.GetByID(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	// Validate file type
	if !u.isValidImageType(file.Filename) {
		return nil, fmt.Errorf("invalid file type. Only jpg, jpeg, png, webp are allowed")
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return nil, fmt.Errorf("file size too large. Maximum 5MB allowed")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%s_%d_%s%s",
		uuid.New().String(),
		time.Now().Unix(),
		string(photoType),
		ext)

	// Create upload directory if not exists
	uploadDir := filepath.Join("uploads", "vehicles", strconv.Itoa(vehicleID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Save file
	filePath := filepath.Join(uploadDir, newFilename)
	if err := u.saveFile(file, filePath); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Get current photo count for sort order
	existingPhotos, err := u.photoRepo.GetByVehicleID(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing photos: %w", err)
	}

	// Determine if this should be primary (first photo)
	isPrimary := len(existingPhotos) == 0

	// Create photo record
	photo := &entities.VehiclePhoto{
		VehicleID:  vehicleID,
		PhotoURL:   "/" + strings.ReplaceAll(filePath, "\\", "/"), // Ensure forward slashes
		PhotoType:  photoType,
		IsPrimary:  isPrimary,
		SortOrder:  len(existingPhotos),
		UploadedBy: uploadedBy,
	}

	if err := u.photoRepo.Create(photo); err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save photo record: %w", err)
	}

	return photo, nil
}

func (u *vehiclePhotoUsecase) GetVehiclePhotos(vehicleID int) ([]entities.VehiclePhoto, error) {
	// Verify vehicle exists
	_, err := u.vehicleRepo.GetByID(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	photos, err := u.photoRepo.GetByVehicleID(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle photos: %w", err)
	}

	return photos, nil
}

func (u *vehiclePhotoUsecase) DeletePhoto(photoID int) error {
	// Get photo to get file path
	photo, err := u.photoRepo.GetByID(photoID)
	if err != nil {
		return fmt.Errorf("photo not found: %w", err)
	}

	// Delete from database first
	if err := u.photoRepo.Delete(photoID); err != nil {
		return fmt.Errorf("failed to delete photo record: %w", err)
	}

	// Delete file (ignore errors as file might not exist)
	if photo.PhotoURL != "" {
		filePath := strings.TrimPrefix(photo.PhotoURL, "/")
		os.Remove(filePath)
	}

	return nil
}

func (u *vehiclePhotoUsecase) SetPrimaryPhoto(vehicleID int, photoID int) error {
	// Verify photo belongs to vehicle
	photo, err := u.photoRepo.GetByID(photoID)
	if err != nil {
		return fmt.Errorf("photo not found: %w", err)
	}

	if photo.VehicleID != vehicleID {
		return fmt.Errorf("photo does not belong to this vehicle")
	}

	if err := u.photoRepo.SetPrimary(vehicleID, photoID); err != nil {
		return fmt.Errorf("failed to set primary photo: %w", err)
	}

	return nil
}

func (u *vehiclePhotoUsecase) UpdatePhotoSortOrder(photoID int, sortOrder int) error {
	if err := u.photoRepo.UpdateSortOrder(photoID, sortOrder); err != nil {
		return fmt.Errorf("failed to update sort order: %w", err)
	}

	return nil
}

func (u *vehiclePhotoUsecase) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validTypes := []string{".jpg", ".jpeg", ".png", ".webp"}

	for _, validType := range validTypes {
		if ext == validType {
			return true
		}
	}

	return false
}

func (u *vehiclePhotoUsecase) saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
