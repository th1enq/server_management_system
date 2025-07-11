package usecases

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/query"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/domain/scope"
	"github.com/th1enq/server_management_system/internal/domain/services"
	"github.com/th1enq/server_management_system/internal/dto"
	"go.uber.org/zap"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*entity.User, error)
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*entity.User, error)
	UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*entity.User, error)
	UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, pagination dto.UserPagination) ([]*entity.User, error)
}

type userUseCase struct {
	userRepo        repository.UserRepository
	passwordService services.PasswordService
	logger          *zap.Logger
}

func NewUserUseCase(userRepo repository.UserRepository, passwordService services.PasswordService, logger *zap.Logger) UserUseCase {
	return &userUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
		logger:          logger,
	}
}

func (u *userUseCase) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*entity.User, error) {
	hashedPassword, err := u.passwordService.Hash(req.Password)
	if err != nil {
		u.logger.Warn("Failed to hash password",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  hashedPassword,
	}
	if req.Role == "" {
		user.Role = scope.UserRoleUser
	} else {
		user.Role = req.Role
	}
	if len(req.Scopes) > 0 {
		user.HashedScopes = scope.ToBitmask(req.Scopes)
	} else {
		user.HashedScopes = scope.GetDefaultScopesHash(user.Role)
	}

	exist, err := u.userRepo.ExistsByUserNameOrEmail(ctx, user.Username, user.Email, 0)
	if err != nil {
		u.logger.Error("Failed to check if user exists",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	} else if exist {
		u.logger.Warn("User already exists",
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
	if updates.NewPassword != updates.RepeatPassword {
		u.logger.Warn("New password and repeat password do not match")
		return fmt.Errorf("new password and repeat password do not match")
	}

	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return fmt.Errorf("user not found: %w", err)
	}

	same, err := u.passwordService.Verify(user.Password, updates.OldPassword)
	if err != nil {
		u.logger.Error("Failed to verify old password",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
		return fmt.Errorf("failed to verify old password: %w", err)
	} else if !same {
		u.logger.Warn("Old password does not match",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
		)
		return fmt.Errorf("old password does not match")
	}

	hashedPassword, err := u.passwordService.Hash(updates.NewPassword)
	if err != nil {
		u.logger.Error("Failed to hash new password",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	user.Password = hashedPassword

	if err := u.userRepo.Update(ctx, user); err != nil {
		u.logger.Error("Failed to update user password",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update user password: %w", err)
	}

	u.logger.Info("User password updated successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

func (u *userUseCase) UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if updates.FirstName != "" {
		user.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		user.LastName = updates.LastName
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		u.logger.Error("Failed to update user profile",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err),
		)
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
		u.logger.Error("Failed to get user by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return fmt.Errorf("user not found: %w", err)
	}
	if err := u.userRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete user",
			zap.Uint("id", user.ID),
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	u.logger.Info("User deleted successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	return nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	exist, err := u.userRepo.ExistsByUserNameOrEmail(ctx, "", updates.Email, id)
	if err != nil {
		u.logger.Error("Failed to check if user exists",
			zap.String("email", updates.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exist {
		u.logger.Error("User already exists",
			zap.String("email", updates.Email),
		)
		return nil, fmt.Errorf("user with username or email already exists")
	}

	if updates.Email != "" {
		user.Email = updates.Email
	}
	if updates.Role != "" {
		user.Role = updates.Role
	}
	if updates.FirstName != "" {
		user.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		user.LastName = updates.LastName
	}
	if updates.IsActive != nil {
		user.IsActive = *updates.IsActive
	} else {
		user.IsActive = true
	}
	if len(updates.Scopes) > 0 {
		user.HashedScopes = scope.ToBitmask(updates.Scopes)
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

func (u *userUseCase) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	u.logger.Info("User retrieved successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)
	return user, nil
}

func (u *userUseCase) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		u.logger.Error("Failed to get user by username",
			zap.String("username", username),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	u.logger.Info("User retrieved successfully by username",
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)
	return user, nil
}

func (u *userUseCase) ListUsers(ctx context.Context, pagnination dto.UserPagination) ([]*entity.User, error) {
	pagninationQuery := query.Pagination{
		PageSize: pagnination.PageSize,
		Page:     pagnination.Page,
		Sort:     pagnination.Sort,
		Order:    pagnination.Order,
	}

	if pagnination.Page == 0 {
		pagninationQuery.Page = 1
	}
	if pagnination.PageSize == 0 {
		pagninationQuery.PageSize = 10
	}
	if pagnination.Sort == "" {
		pagninationQuery.Sort = "created_time"
	}
	if pagnination.Order == "" {
		pagninationQuery.Order = "desc"
	}

	users, err := u.userRepo.List(ctx, pagninationQuery)
	if err != nil {
		u.logger.Error("Failed to list users",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
