package handlers

import (
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase usecases.AuthUsecase
}

func NewAuthHandler(authUsecase usecases.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request entities.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	loginResponse, err := h.authUsecase.Login(&request, clientIP)
	if err != nil {
		response.BadRequest(c, "Login failed", err)
		return
	}

	response.Success(c, "Login successful", loginResponse)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		response.BadRequest(c, "Token not found", nil)
		return
	}

	err := h.authUsecase.Logout(token.(string))
	if err != nil {
		response.InternalServerError(c, "Logout failed", err)
		return
	}

	response.Success(c, "Logout successful", nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		response.BadRequest(c, "Token not found", nil)
		return
	}

	refreshResponse, err := h.authUsecase.RefreshToken(token.(string))
	if err != nil {
		response.BadRequest(c, "Token refresh failed", err)
		return
	}

	response.Success(c, "Token refreshed successfully", refreshResponse)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		response.BadRequest(c, "Token not found", nil)
		return
	}

	user, err := h.authUsecase.GetCurrentUser(token.(string))
	if err != nil {
		response.BadRequest(c, "Failed to get current user", err)
		return
	}

	response.Success(c, "Current user retrieved successfully", user)
}
