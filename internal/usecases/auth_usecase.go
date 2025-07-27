package usecases

import (
	"fmt"
	"time"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"
	"vehicle-showroom-backend/pkg/jwt"
	"vehicle-showroom-backend/pkg/password"
)

type AuthUsecase interface {
	Login(request *entities.LoginRequest, ipAddress string) (*entities.LoginResponse, error)
	Logout(token string) error
	RefreshToken(token string) (*entities.LoginResponse, error)
	GetCurrentUser(token string) (*entities.User, error)
}

type authUsecase struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	jwtManager  *jwt.JWTManager
	config      *config.Config
}

func NewAuthUsecase(userRepo repositories.UserRepository, sessionRepo repositories.SessionRepository, config *config.Config) AuthUsecase {
	jwtManager := jwt.NewJWTManager(config.JWTSecret, config.JWTExpireHours)
	return &authUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
		config:      config,
	}
}

func (u *authUsecase) Login(request *entities.LoginRequest, ipAddress string) (*entities.LoginResponse, error) {
	// Get user by username
	user, err := u.userRepo.GetByUsername(request.Username)

	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Verify password
	if err = password.CheckPassword(request.Password, user.PasswordHash); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, expiresAt, err := u.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session record
	session := &entities.UserSession{
		UserID:       user.ID,
		SessionToken: token,
		LoginAt:      time.Now(),
		IPAddress:    ipAddress,
		IsActive:     true,
	}

	if err := u.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &entities.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

func (u *authUsecase) Logout(token string) error {
	// Update session logout time
	if err := u.sessionRepo.UpdateLogout(token); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

func (u *authUsecase) RefreshToken(token string) (*entities.LoginResponse, error) {
	// Validate current token and get claims
	claims, err := u.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user by ID from claims
	user, err := u.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Generate new token
	newToken, expiresAt, err := u.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	// Update old session logout time
	u.sessionRepo.UpdateLogout(token)

	// Create new session record
	session := &entities.UserSession{
		UserID:       user.ID,
		SessionToken: newToken,
		LoginAt:      time.Now(),
		IPAddress:    "", // We don't have IP from token, could be improved
		IsActive:     true,
	}

	if err := u.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &entities.LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

func (u *authUsecase) GetCurrentUser(token string) (*entities.User, error) {
	// Validate token and get claims
	claims, err := u.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user by ID from claims
	user, err := u.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Verify session is still active
	session, err := u.sessionRepo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	if !session.IsActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}
