package services

import (
	"context"
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/repository"
	"go.uber.org/zap"
)

type IUserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*models.User, error)
	UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
	clearUserCache(ctx context.Context, user *models.User) error
}

type userService struct {
	userRepo repository.IUserRepository
	cache    db.IRedisClient
	logger   *zap.Logger
}

func NewUserService(userRepo repository.IUserRepository, cache db.IRedisClient, logger *zap.Logger) IUserService {
	return &userService{
		userRepo: userRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (u *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Scopes:    models.ToBitmask(req.Scopes),
	}
	user.SetPassword(req.Password)

	exist, err := u.userRepo.ExistsByUserNameOrEmail(ctx, user.Username, user.Email)
	if err != nil {
		u.logger.Error("Failed to check if user exists",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exist {
		u.logger.Error("User already exists",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
		)
		return nil, fmt.Errorf("user with username or email already exists")
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

	if err := u.clearUserCache(ctx, user); err != nil {
		u.logger.Error("Failed to clear user cache",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
	}

	u.logger.Info("User profile updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return user, nil
}

func (u *userService) DeleteUser(ctx context.Context, id uint) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := u.clearUserCache(ctx, user); err != nil {
		u.logger.Error("Failed to clear user cache",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
	}

	u.logger.Info("User deleted successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return nil
}

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

	exist, err := u.userRepo.ExistsByUserNameOrEmail(ctx, user.Username, user.Email)
	if err != nil {
		u.logger.Error("Failed to check if user exists",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exist {
		u.logger.Error("User already exists",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
		)
		return nil, fmt.Errorf("user with username or email already exists")
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if err := u.clearUserCache(ctx, user); err != nil {
		u.logger.Error("Failed to clear user cache",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
	}

	u.logger.Info("User updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)

	return user, nil
}

func (u *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)
	var cachedUser *models.User
	if err := u.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
		u.logger.Info("User retrieved from cache",
			zap.Uint("id", cachedUser.ID),
			zap.String("username", cachedUser.Username),
		)
		return cachedUser, nil
	} else if err != db.ErrCacheMiss {
		u.logger.Warn("Cache miss for user",
			zap.Uint("id", id),
			zap.Error(err),
		)
	}
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	if err := u.cache.Set(ctx, cacheKey, user, 30*time.Minute); err != nil {
		u.logger.Error("Failed to cache user data",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
	}
	return user, nil
}

func (u *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:username:%s", username)
	var cachedUser *models.User
	if err := u.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
		u.logger.Info("User retrieved from cache",
			zap.String("username", cachedUser.Username),
			zap.Uint("id", cachedUser.ID),
		)
		return cachedUser, nil
	} else if err != db.ErrCacheMiss {
		u.logger.Warn("Cache miss for user by username",
			zap.String("username", username),
			zap.Error(err),
		)
	}
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	if err := u.cache.Set(ctx, cacheKey, user, 30*time.Minute); err != nil {
		u.logger.Error("Failed to cache user data",
			zap.String("username", user.Username),
			zap.Uint("id", user.ID),
			zap.Error(err),
		)
	}
	return user, nil
}

func (u *userService) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	users, err := u.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (u *userService) clearUserCache(ctx context.Context, user *models.User) error {
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	if err := u.cache.Del(ctx, cacheKey); err != nil {
		u.logger.Warn("Failed to delete user from cache",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
	}

	cacheKey = fmt.Sprintf("user:username:%s", user.Username)
	if err := u.cache.Del(ctx, cacheKey); err != nil {
		u.logger.Warn("Failed to delete user by username from cache",
			zap.String("username", user.Username),
			zap.Uint("id", user.ID),
			zap.Error(err),
		)
	}

	return nil
}
