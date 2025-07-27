package handlers

import (
	"strconv"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type VehicleCategoryHandler struct {
	categoryUsecase usecases.VehicleCategoryUsecase
}

func NewVehicleCategoryHandler(categoryUsecase usecases.VehicleCategoryUsecase) *VehicleCategoryHandler {
	return &VehicleCategoryHandler{
		categoryUsecase: categoryUsecase,
	}
}

func (h *VehicleCategoryHandler) CreateCategory(c *gin.Context) {
	var request entities.CreateCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	category, err := h.categoryUsecase.CreateCategory(&request)
	if err != nil {
		response.BadRequest(c, "Failed to create category", err)
		return
	}

	response.Created(c, "Category created successfully", category)
}

func (h *VehicleCategoryHandler) GetCategoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid category ID", err)
		return
	}

	category, err := h.categoryUsecase.GetCategoryByID(id)
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	response.Success(c, "Category retrieved successfully", category)
}

func (h *VehicleCategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryUsecase.GetCategories()
	if err != nil {
		response.InternalServerError(c, "Failed to get categories", err)
		return
	}

	response.Success(c, "Categories retrieved successfully", categories)
}

func (h *VehicleCategoryHandler) UpdateCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid category ID", err)
		return
	}

	var request entities.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	category, err := h.categoryUsecase.UpdateCategory(id, &request)
	if err != nil {
		response.BadRequest(c, "Failed to update category", err)
		return
	}

	response.Success(c, "Category updated successfully", category)
}

func (h *VehicleCategoryHandler) DeleteCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid category ID", err)
		return
	}

	err = h.categoryUsecase.DeleteCategory(id)
	if err != nil {
		response.BadRequest(c, "Failed to delete category", err)
		return
	}

	response.Success(c, "Category deleted successfully", nil)
}
