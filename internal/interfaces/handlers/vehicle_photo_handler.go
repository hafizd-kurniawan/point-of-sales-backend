package handlers

import (
	"strconv"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type VehiclePhotoHandler struct {
	photoUsecase usecases.VehiclePhotoUsecase
}

func NewVehiclePhotoHandler(photoUsecase usecases.VehiclePhotoUsecase) *VehiclePhotoHandler {
	return &VehiclePhotoHandler{
		photoUsecase: photoUsecase,
	}
}

func (h *VehiclePhotoHandler) UploadPhoto(c *gin.Context) {
	vehicleIDParam := c.Param("vehicleId")
	vehicleID, err := strconv.Atoi(vehicleIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle ID", err)
		return
	}

	// Get uploaded file
	file, err := c.FormFile("photo")
	if err != nil {
		response.BadRequest(c, "No photo file uploaded", err)
		return
	}

	// Get photo type (default to exterior)
	photoTypeStr := c.DefaultPostForm("photo_type", "exterior")
	photoType := entities.PhotoType(photoTypeStr)

	// Validate photo type
	validTypes := []entities.PhotoType{
		entities.PhotoExterior,
		entities.PhotoInterior,
		entities.PhotoEngine,
		entities.PhotoDocument,
	}

	isValidType := false
	for _, validType := range validTypes {
		if photoType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		response.BadRequest(c, "Invalid photo type", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	uploadedBy, ok := userID.(int)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format", nil)
		return
	}

	// Upload photo
	photo, err := h.photoUsecase.UploadPhoto(vehicleID, file, photoType, uploadedBy)
	if err != nil {
		response.BadRequest(c, "Failed to upload photo", err)
		return
	}

	response.Created(c, "Photo uploaded successfully", photo)
}

func (h *VehiclePhotoHandler) GetVehiclePhotos(c *gin.Context) {
	vehicleIDParam := c.Param("vehicleId")
	vehicleID, err := strconv.Atoi(vehicleIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle ID", err)
		return
	}

	photos, err := h.photoUsecase.GetVehiclePhotos(vehicleID)
	if err != nil {
		response.BadRequest(c, "Failed to get vehicle photos", err)
		return
	}

	response.Success(c, "Photos retrieved successfully", photos)
}

func (h *VehiclePhotoHandler) DeletePhoto(c *gin.Context) {
	photoIDParam := c.Param("photoId")
	photoID, err := strconv.Atoi(photoIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID", err)
		return
	}

	err = h.photoUsecase.DeletePhoto(photoID)
	if err != nil {
		response.BadRequest(c, "Failed to delete photo", err)
		return
	}

	response.Success(c, "Photo deleted successfully", nil)
}

func (h *VehiclePhotoHandler) SetPrimaryPhoto(c *gin.Context) {
	vehicleIDParam := c.Param("vehicleId")
	vehicleID, err := strconv.Atoi(vehicleIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle ID", err)
		return
	}

	photoIDParam := c.Param("photoId")
	photoID, err := strconv.Atoi(photoIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID", err)
		return
	}

	err = h.photoUsecase.SetPrimaryPhoto(vehicleID, photoID)
	if err != nil {
		response.BadRequest(c, "Failed to set primary photo", err)
		return
	}

	response.Success(c, "Primary photo set successfully", nil)
}

func (h *VehiclePhotoHandler) UpdateSortOrder(c *gin.Context) {
	photoIDParam := c.Param("photoId")
	photoID, err := strconv.Atoi(photoIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID", err)
		return
	}

	var request struct {
		SortOrder int `json:"sort_order" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	err = h.photoUsecase.UpdatePhotoSortOrder(photoID, request.SortOrder)
	if err != nil {
		response.BadRequest(c, "Failed to update sort order", err)
		return
	}

	response.Success(c, "Sort order updated successfully", nil)
}
