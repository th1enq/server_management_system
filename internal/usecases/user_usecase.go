package usecases

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/infrastructure/repository"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, req dto.RegisterRequest) (*domain.User, error)
	GetUserByID(ctx context.Context, id uint) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*domain.User, error)
	UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*domain.User, error)
	UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewUserUseCase(userRepo repository.UserRepository, logger *zap.Logger) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *userUseCase) CreateUser(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	user := &domain.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Scopes:    domain.GetDefaultScopesMask(domain.RoleUser),
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

func (u *userUseCase) UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if utils.CheckPassword(updates.OldPassword, user.Password) {
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

func (u *userUseCase) UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*domain.User, error) {
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

func (u *userUseCase) DeleteUser(ctx context.Context, id uint) error {
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

func (u *userUseCase) UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Username = updates.Username
	user.Email = updates.Email
	if updates.Password != "" {
		user.SetPassword(updates.Password)
	}
	user.Role = domain.UserRole(updates.Role)
	user.FirstName = updates.FirstName
	user.LastName = updates.LastName
	user.IsActive = updates.IsActive
	user.Scopes = domain.ToBitmask(updates.Scopes)

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

	u.logger.Info("User updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)

	return user, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (u *userUseCase) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

func (u *userUseCase) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	users, err := u.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
