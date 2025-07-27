package handlers

import (
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	statsUsecase usecases.StatisticsUsecase
}

func NewStatisticsHandler(statsUsecase usecases.StatisticsUsecase) *StatisticsHandler {
	return &StatisticsHandler{
		statsUsecase: statsUsecase,
	}
}

func (h *StatisticsHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.statsUsecase.GetDashboardStats()
	if err != nil {
		response.InternalServerError(c, "Failed to get dashboard statistics", err)
		return
	}

	response.Success(c, "Dashboard statistics retrieved successfully", stats)
}

func (h *StatisticsHandler) GetVehicleStats(c *gin.Context) {
	stats, err := h.statsUsecase.GetVehicleStats()
	if err != nil {
		response.InternalServerError(c, "Failed to get vehicle statistics", err)
		return
	}

	response.Success(c, "Vehicle statistics retrieved successfully", stats)
}

func (h *StatisticsHandler) GetCustomerStats(c *gin.Context) {
	stats, err := h.statsUsecase.GetCustomerStats()
	if err != nil {
		response.InternalServerError(c, "Failed to get customer statistics", err)
		return
	}

	response.Success(c, "Customer statistics retrieved successfully", stats)
}
