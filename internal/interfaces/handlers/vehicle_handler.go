package handlers

import (
	"strconv"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	vehicleUsecase usecases.VehicleUsecase
}

func NewVehicleHandler(vehicleUsecase usecases.VehicleUsecase) *VehicleHandler {
	return &VehicleHandler{
		vehicleUsecase: vehicleUsecase,
	}
}

func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	var request entities.CreateVehicleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	vehicle, err := h.vehicleUsecase.CreateVehicle(&request)
	if err != nil {
		response.BadRequest(c, "Failed to create vehicle", err)
		return
	}

	response.Created(c, "Vehicle created successfully", vehicle)
}

func (h *VehicleHandler) GetVehicleByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle ID", err)
		return
	}

	vehicle, err := h.vehicleUsecase.GetVehicleByID(id)
	if err != nil {
		response.NotFound(c, "Vehicle not found")
		return
	}

	response.Success(c, "Vehicle retrieved successfully", vehicle)
}

func (h *VehicleHandler) GetVehicles(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	vehicles, total, err := h.vehicleUsecase.GetVehicles(page, limit)
	if err != nil {
		response.InternalServerError(c, "Failed to get vehicles", err)
		return
	}

	// Calculate pagination meta
	totalPages := (total + limit - 1) / limit
	meta := response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Vehicles retrieved successfully", vehicles, meta)
}

func (h *VehicleHandler) SearchVehicles(c *gin.Context) {
	var request entities.VehicleSearchRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.BadRequest(c, "Invalid search parameters", err)
		return
	}

	// Set defaults
	if request.Page < 1 {
		request.Page = 1
	}
	if request.Limit < 1 || request.Limit > 100 {
		request.Limit = 10
	}

	vehicles, total, err := h.vehicleUsecase.SearchVehicles(&request)
	if err != nil {
		response.InternalServerError(c, "Failed to search vehicles", err)
		return
	}

	// Calculate pagination meta
	totalPages := (total + request.Limit - 1) / request.Limit
	meta := response.PaginationMeta{
		Page:       request.Page,
		Limit:      request.Limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Vehicle search completed successfully", vehicles, meta)
}

func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle ID", err)
		return
	}

	var request entities.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	vehicle, err := h.vehicleUsecase.UpdateVehicle(id, &request)
	if err != nil {
		response.BadRequest(c, "Failed to update vehicle", err)
		return
	}

	response.Success(c, "Vehicle updated successfully", vehicle)
}

func (h *VehicleHandler) DeleteVehicle(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle ID", err)
		return
	}

	err = h.vehicleUsecase.DeleteVehicle(id)
	if err != nil {
		response.BadRequest(c, "Failed to delete vehicle", err)
		return
	}

	response.Success(c, "Vehicle deleted successfully", nil)
}
