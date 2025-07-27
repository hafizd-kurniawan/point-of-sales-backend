package usecases

import (
	"fmt"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"
	"vehicle-showroom-backend/pkg/password"
)

type UserUsecase interface {
	CreateUser(request *entities.CreateUserRequest) (*entities.User, error)
	GetUserByID(id int) (*entities.User, error)
	GetUsers(page, limit int) ([]entities.User, int, error)
	UpdateUser(id int, request *entities.UpdateUserRequest) (*entities.User, error)
	DeleteUser(id int) error
}

type userUsecase struct {
	userRepo repositories.UserRepository
	config   *config.Config
}

func NewUserUsecase(userRepo repositories.UserRepository, config *config.Config) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		config:   config,
	}
}

func (u *userUsecase) CreateUser(request *entities.CreateUserRequest) (*entities.User, error) {
	// Check if username already exists
	if _, err := u.userRepo.GetByUsername(request.Username); err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if _, err := u.userRepo.GetByEmail(request.Email); err == nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := password.HashPassword(request.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := &entities.User{
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: hashedPassword,
		FullName:     request.FullName,
		Phone:        request.Phone,
		Role:         request.Role,
		IsActive:     true,
	}

	// Save to database
	if err := u.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}

func (u *userUsecase) GetUserByID(id int) (*entities.User, error) {
	user, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}

func (u *userUsecase) GetUsers(page, limit int) ([]entities.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	users, err := u.userRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// Remove password hashes from response
	for i := range users {
		users[i].PasswordHash = ""
	}

	// Get total count
	total, err := u.userRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

func (u *userUsecase) UpdateUser(id int, request *entities.UpdateUserRequest) (*entities.User, error) {
	// Get existing user
	existingUser, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check username uniqueness if changed
	if request.Username != "" && request.Username != existingUser.Username {
		if _, err := u.userRepo.GetByUsername(request.Username); err == nil {
			return nil, fmt.Errorf("username already exists")
		}
	}

	// Check email uniqueness if changed
	if request.Email != "" && request.Email != existingUser.Email {
		if _, err := u.userRepo.GetByEmail(request.Email); err == nil {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Prepare update entity
	updateUser := &entities.User{
		Username: request.Username,
		Email:    request.Email,
		FullName: request.FullName,
		Phone:    request.Phone,
		Role:     request.Role,
	}

	// Hash new password if provided
	if request.Password != "" {
		hashedPassword, err := password.HashPassword(request.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		updateUser.PasswordHash = hashedPassword
	}

	// Set is_active if provided
	if request.IsActive != nil {
		updateUser.IsActive = *request.IsActive
	}

	// Update user
	if err := u.userRepo.Update(id, updateUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user
	updatedUser, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	// Remove password hash from response
	updatedUser.PasswordHash = ""

	return updatedUser, nil
}

func (u *userUsecase) DeleteUser(id int) error {
	// Check if user exists
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Soft delete user
	if err := u.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
