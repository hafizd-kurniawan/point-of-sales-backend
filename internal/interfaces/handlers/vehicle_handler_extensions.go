package handlers

import (
	"net/http"
	"strconv"
	"time"
	"vehicle-showroom-backend/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

// UpdateVehicleStatus updates vehicle availability status
func (h *VehicleHandler) UpdateVehicleStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid vehicle ID",
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		AvailabilityStatus string `json:"availability_status" binding:"required"`
		Notes              string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Validate status
	validStatuses := []string{"available", "reserved", "sold", "maintenance"}
	isValid := false
	for _, status := range validStatuses {
		if req.AvailabilityStatus == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid status. Must be one of: available, reserved, sold, maintenance",
		})
		return
	}

	// Update vehicle with new status
	updateReq := &entities.UpdateVehicleRequest{
		AvailabilityStatus: entities.AvailabilityStatus(req.AvailabilityStatus),
		Description:        req.Notes,
	}

	updatedVehicle, err := h.vehicleUsecase.UpdateVehicle(id, updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to update vehicle status",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vehicle status updated successfully",
		"data":    updatedVehicle,
	})
}

// UpdateVehicleCondition updates vehicle condition status
func (h *VehicleHandler) UpdateVehicleCondition(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid vehicle ID",
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		ConditionStatus string `json:"condition_status" binding:"required"`
		Mileage         int    `json:"mileage"`
		Notes           string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Validate condition
	validConditions := []string{"excellent", "good", "fair", "poor"}
	isValid := false
	for _, condition := range validConditions {
		if req.ConditionStatus == condition {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid condition. Must be one of: excellent, good, fair, poor",
		})
		return
	}

	// Update vehicle with new condition
	updateReq := &entities.UpdateVehicleRequest{
		ConditionStatus: entities.ConditionStatus(req.ConditionStatus),
		Mileage:         &req.Mileage,
		Description:     req.Notes,
	}

	updatedVehicle, err := h.vehicleUsecase.UpdateVehicle(id, updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to update vehicle condition",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vehicle condition updated successfully",
		"data":    updatedVehicle,
	})
}

// GetMechanicDashboard gets dashboard for mechanics
func (h *VehicleHandler) GetMechanicDashboard(c *gin.Context) {
	mechanicID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get all vehicles (simple approach)
	vehicles, total, err := h.vehicleUsecase.GetVehicles(1, 1000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to get vehicles",
			"error":   err.Error(),
		})
		return
	}

	// Filter vehicles manually
	maintenanceCount := 0
	conditionStats := map[string]int{
		"excellent": 0,
		"good":      0,
		"fair":      0,
		"poor":      0,
	}

	var maintenanceVehicles []entities.Vehicle

	for _, vehicle := range vehicles {
		if vehicle.AvailabilityStatus == "maintenance" {
			maintenanceCount++
			maintenanceVehicles = append(maintenanceVehicles, vehicle)
		}

		if count, exists := conditionStats[string(vehicle.ConditionStatus)]; exists {
			conditionStats[string(vehicle.ConditionStatus)] = count + 1
		}
	}

	dashboard := map[string]interface{}{
		"mechanic_id":             mechanicID,
		"vehicles_in_maintenance": maintenanceCount,
		"pending_work":            maintenanceCount,
		"condition_breakdown":     conditionStats,
		"maintenance_vehicles":    maintenanceVehicles,
		"total_vehicles":          total,
		"last_updated":            time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mechanic dashboard retrieved successfully",
		"data":    dashboard,
	})
}

// GetMaintenanceVehicles gets vehicles currently in maintenance
func (h *VehicleHandler) GetMaintenanceVehicles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get all vehicles and filter manually
	vehicles, _, err := h.vehicleUsecase.GetVehicles(1, 1000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to get vehicles",
			"error":   err.Error(),
		})
		return
	}

	// Filter maintenance vehicles
	var maintenanceVehicles []entities.Vehicle
	for _, vehicle := range vehicles {
		if vehicle.AvailabilityStatus == "maintenance" {
			maintenanceVehicles = append(maintenanceVehicles, vehicle)
		}
	}

	// Simple pagination
	start := (page - 1) * limit
	end := start + limit
	if start >= len(maintenanceVehicles) {
		maintenanceVehicles = []entities.Vehicle{}
	} else {
		if end > len(maintenanceVehicles) {
			end = len(maintenanceVehicles)
		}
		maintenanceVehicles = maintenanceVehicles[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Maintenance vehicles retrieved successfully",
		"data":    maintenanceVehicles,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       len(maintenanceVehicles),
			"total_pages": (len(maintenanceVehicles) + limit - 1) / limit,
		},
	})
}

// GetMechanicWorkSummary gets work summary for mechanic
func (h *VehicleHandler) GetMechanicWorkSummary(c *gin.Context) {
	mechanicID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get all vehicles
	vehicles, _, err := h.vehicleUsecase.GetVehicles(1, 1000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to get vehicles",
			"error":   err.Error(),
		})
		return
	}

	// Calculate summary
	summary := map[string]interface{}{
		"mechanic_id": mechanicID,
		"date":        time.Now().Format("2006-01-02"),
	}

	conditionCounts := map[string]int{
		"excellent": 0,
		"good":      0,
		"fair":      0,
		"poor":      0,
	}

	maintenanceCount := 0
	totalVehicles := len(vehicles)

	for _, vehicle := range vehicles {
		if vehicle.AvailabilityStatus == "maintenance" {
			maintenanceCount++
		}
		contionStatus := entities.ConditionStatus(vehicle.ConditionStatus)

		if count, exists := conditionCounts[string(contionStatus)]; exists {
			conditionCounts[string(contionStatus)] = count + 1
		}
	}

	summary["maintenance_count"] = maintenanceCount
	summary["total_vehicles"] = totalVehicles
	summary["excellent_count"] = conditionCounts["excellent"]
	summary["good_count"] = conditionCounts["good"]
	summary["fair_count"] = conditionCounts["fair"]
	summary["poor_count"] = conditionCounts["poor"]

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mechanic work summary retrieved successfully",
		"data":    summary,
	})
}
