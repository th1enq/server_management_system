package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/domain/services"
	"github.com/th1enq/server_management_system/internal/dto"
	"go.uber.org/zap"
)

type AuthUseCase interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (*dto.Claims, error)
	Logout(ctx context.Context, userID uint) error
}

type authUseCase struct {
	userUseCase     UserUseCase
	tokenServices   services.TokenServices
	tokenRepository repository.TokenRepository
	passwordService services.PasswordService
	logger          *zap.Logger
}

func NewAuthUseCase(userUseCase UserUseCase, tokenRepository repository.TokenRepository, tokenServices services.TokenServices, passwordServices services.PasswordService, logger *zap.Logger) AuthUseCase {
	return &authUseCase{
		userUseCase:     userUseCase,
		tokenRepository: tokenRepository,
		tokenServices:   tokenServices,
		passwordService: passwordServices,
		logger:          logger,
	}
}

func (a *authUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	username := req.Username
	password := req.Password

	user, err := a.userUseCase.GetUserByUsername(ctx, username)
	if err != nil {
		a.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		a.logger.Warn("Inactive user attempted login", zap.String("username", username))
		return nil, fmt.Errorf("account is disabled")
	}

	same, err := a.passwordService.Verify(user.Password, password)
	if err != nil {
		a.logger.Error("Failed to verify password", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("invalid credentials")
	} else if !same {
		a.logger.Warn("Invalid password attempt", zap.String("username", username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, err := a.tokenServices.GenerateAccessToken(user)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.tokenServices.GenerateRefreshToken(user)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	// Add tokens to Redis whitelist
	if err := a.tokenRepository.AddTokenToWhitelist(ctx, accessToken, user.ID, time.Hour*24); err != nil {
		a.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenRepository.AddTokenToWhitelist(ctx, refreshToken, user.ID, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	a.logger.Info("User logged in successfully", zap.String("username", username), zap.Uint("user_id", user.ID))

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (a *authUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	createUserRequeset := dto.CreateUserRequest{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	createdUser, err := a.userUseCase.CreateUser(ctx, createUserRequeset)
	if err != nil {
		a.logger.Error("Failed to create user", zap.String("username", req.Username), zap.Error(err))
		if err.Error() == "user already exists" {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to create user")
	}

	// Generate tokens
	accessToken, err := a.tokenServices.GenerateAccessToken(createdUser)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", createdUser.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.tokenServices.GenerateRefreshToken(createdUser)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", createdUser.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	// Add tokens to Redis whitelist
	if err := a.tokenRepository.AddTokenToWhitelist(ctx, accessToken, createdUser.ID, time.Hour*24); err != nil {
		a.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenRepository.AddTokenToWhitelist(ctx, refreshToken, createdUser.ID, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	a.logger.Info("User registered successfully", zap.String("username", createdUser.Username), zap.Uint("user_id", createdUser.ID))

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (a *authUseCase) ValidateToken(ctx context.Context, tokenString string) (*dto.Claims, error) {
	// Check if token is whitelisted
	if !a.tokenRepository.IsTokenWhitelisted(ctx, tokenString) {
		a.logger.Warn("Token is not whitelisted", zap.String("token", tokenString))
		return nil, fmt.Errorf("token is not whitelisted")
	}
	return a.tokenServices.ValidateToken(tokenString)
}

// RefreshToken implements authUseCase.
func (a *authUseCase) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	refreshToken := req.RefreshToken
	// Validate refresh token
	claims, err := a.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := a.userUseCase.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Generate new tokens
	newAccessToken, err := a.tokenServices.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := a.tokenServices.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// Add new tokens to whitelist and remove old refresh token
	if err := a.tokenRepository.AddTokenToWhitelist(ctx, newAccessToken, user.ID, time.Hour*24); err != nil {
		a.logger.Error("Failed to add new access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenRepository.AddTokenToWhitelist(ctx, newRefreshToken, user.ID, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add new refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	// Remove old refresh token from whitelist
	if err := a.tokenRepository.RemoveTokenFromWhitelist(ctx, refreshToken); err != nil {
		a.logger.Error("Failed to remove old refresh token from whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to remove old token from whitelist")
	}

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}, nil
}

// Logout implements authUseCase.
func (a *authUseCase) Logout(ctx context.Context, userID uint) error {
	// Remove all user tokens from whitelist
	if err := a.tokenRepository.RemoveUserTokensFromWhitelist(ctx, userID); err != nil {
		a.logger.Error("Failed to remove user tokens from whitelist", zap.Uint("user_id", userID), zap.Error(err))
		return fmt.Errorf("failed to logout user")
	}
	a.logger.Info("User logged out", zap.Uint("user_id", userID))
	return nil
}
