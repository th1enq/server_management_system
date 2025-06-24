package services

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser implements UserService.
func (u *userService) CreateUser(ctx context.Context, user *models.User) error {
	if user.Username == "" || user.Password == "" {
		return fmt.Errorf("username and password are required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	existing, _ := u.userRepo.GetByUsername(ctx, user.Username)
	if existing != nil {
		return fmt.Errorf("user already exists")
	}

	err := u.userRepo.Create(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// DeleteUser implements UserService.
func (u *userService) DeleteUser(ctx context.Context, id uint) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted successfully",
		zap.Uint("id", id),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.Time("deleted_at", user.DeletedAt.Time),
	)
	return nil
}

// UpdateUser implements UserService.
func (u *userService) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if len(updates) == 0 {
		return user, nil // No updates to apply
	}

	for key, value := range updates {
		switch key {
		case "role":
			if role, ok := value.(string); ok {
				// Validate role
				if role != string(models.UserRoleAdmin) && role != string(models.UserRoleUser) {
					return nil, fmt.Errorf("invalid role: %s", role)
				}
				user.Role = models.UserRole(role)
			} else {
				return nil, fmt.Errorf("invalid role value: %v", value)
			}
		case "password":
			if password, ok := value.(string); ok && password != "" {
				if err := user.SetPassword(password); err != nil {
					return nil, fmt.Errorf("failed to set password: %w", err)
				}
			} else {
				return nil, fmt.Errorf("invalid password value: %v", value)
			}
		}
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return user, nil
}
