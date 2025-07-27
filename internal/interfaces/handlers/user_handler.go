package handlers

import (
	"strconv"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase usecases.UserUsecase
}

func NewUserHandler(userUsecase usecases.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var request entities.CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	user, err := h.userUsecase.CreateUser(&request)
	if err != nil {
		response.BadRequest(c, "Failed to create user", err)
		return
	}

	response.Created(c, "User created successfully", user)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	user, err := h.userUsecase.GetUserByID(id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, "User retrieved successfully", user)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
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

	users, total, err := h.userUsecase.GetUsers(page, limit)
	if err != nil {
		response.InternalServerError(c, "Failed to get users", err)
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

	response.Paginated(c, "Users retrieved successfully", users, meta)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	var request entities.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	user, err := h.userUsecase.UpdateUser(id, &request)
	if err != nil {
		response.BadRequest(c, "Failed to update user", err)
		return
	}

	response.Success(c, "User updated successfully", user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	// Get current user ID to prevent self-deletion
	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(int) == id {
		response.BadRequest(c, "Cannot delete your own account", nil)
		return
	}

	err = h.userUsecase.DeleteUser(id)
	if err != nil {
		response.BadRequest(c, "Failed to delete user", err)
		return
	}

	response.Success(c, "User deleted successfully", nil)
}
