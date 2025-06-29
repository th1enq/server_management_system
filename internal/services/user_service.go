package services

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	UpdateLastLogin(ctx context.Context, userID uint) error
}

type userService struct {
	userRepo repositories.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repositories.UserRepository, logger *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
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

	u.logger.Info("User deleted successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
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
				switch role {
				case string(models.RoleAdmin):
					user.Role = models.RoleAdmin
				case string(models.RoleUser):
					user.Role = models.RoleUser
				default:
					return nil, fmt.Errorf("invalid role value: %s", role)
				}
			} else {
				return nil, fmt.Errorf("invalid type: %v", value)
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

	u.logger.Info("User updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)

	return user, nil
}

// GetUserByID implements UserService.
func (u *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetUserByUsername implements UserService.
func (u *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// GetUserByEmail implements UserService.
func (u *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// ListUsers implements UserService.
func (u *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users, err := u.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// UpdateLastLogin implements UserService.
func (u *userService) UpdateLastLogin(ctx context.Context, userID uint) error {
	err := u.userRepo.UpdateLastLogin(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}
