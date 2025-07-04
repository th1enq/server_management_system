package services

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/repositories"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*models.User, error)
	UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
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
func (u *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Scopes:    models.ToBitmask(req.Scopes),
	}
	user.SetPassword(req.Password)

	if _, err := u.userRepo.GetByUsername(ctx, user.Username); err == nil {
		u.logger.Error("Username already exists",
			zap.String("username", user.Username),
		)
		return nil, fmt.Errorf("username already exists")
	}

	if _, err := u.userRepo.GetByEmail(ctx, user.Email); err == nil {
		u.logger.Error("Email already exists",
			zap.String("email", user.Email),
		)
		return nil, fmt.Errorf("email already exists")
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		u.logger.Error("Failed to create user",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	u.logger.Info("User created successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return user, nil
}

func (u *userService) UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.CheckPassword(updates.OldPassword) != nil {
		return fmt.Errorf("old password is incorrect")
	}

	if updates.NewPassword != updates.RepeatPassword {
		return fmt.Errorf("new password and repeat password do not match")
	}

	user.SetPassword(updates.NewPassword)

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	u.logger.Info("User password updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

func (u *userService) UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*models.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.FirstName = updates.FirstName
	user.LastName = updates.LastName

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	u.logger.Info("User profile updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return user, nil
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
func (u *userService) UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*models.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Username = updates.Username
	user.Email = updates.Email
	if updates.Password != "" {
		user.SetPassword(updates.Password)
	}
	user.Role = models.UserRole(updates.Role)
	user.FirstName = updates.FirstName
	user.LastName = updates.LastName
	user.IsActive = updates.IsActive
	user.Scopes = models.ToBitmask(updates.Scopes)

	if existUser, err := u.userRepo.GetByUsername(ctx, user.Username); err == nil && existUser.ID != user.ID {
		u.logger.Error("Username already exists",
			zap.String("username", user.Username),
		)
		return nil, fmt.Errorf("username already exists")
	}

	if existUser, err := u.userRepo.GetByEmail(ctx, user.Email); err == nil && existUser.ID != user.ID {
		u.logger.Error("Email already exists",
			zap.String("email", user.Email),
		)
		return nil, fmt.Errorf("email already exists")
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
