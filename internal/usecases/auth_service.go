package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
	"go.uber.org/zap"
)

type AuthUsecase interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, userID uint) error
	LogoutWithToken(ctx context.Context, token string) error
}

type authUsecase struct {
	userUsecase  UserUsecase
	tokenUsecase TokenUsecase
	logger       *zap.Logger
}

func NewAuthUsecase(userUsecase UserUsecase, tokenUsecase TokenUsecase, logger *zap.Logger) AuthUsecase {
	return &authUsecase{
		userUsecase:  userUsecase,
		tokenUsecase: tokenUsecase,
		logger:       logger,
	}
}

func (a *authUsecase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	username := req.Username
	password := req.Password

	// Get user by username
	user, err := a.userUsecase.GetUserByUsername(ctx, username)
	if err != nil {
		a.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		a.logger.Warn("Inactive user attempted login", zap.String("username", username))
		return nil, fmt.Errorf("account is disabled")
	}

	// Verify password
	if err := user.CheckPassword(password); err != nil {
		a.logger.Warn("Invalid password attempt", zap.String("username", username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, err := a.tokenUsecase.GenerateAccessToken(user)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.tokenUsecase.GenerateRefreshToken(user)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	// Add tokens to Redis whitelist
	if err := a.tokenUsecase.AddTokenToWhitelist(ctx, accessToken, user.ID, time.Hour*24); err != nil {
		a.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenUsecase.AddTokenToWhitelist(ctx, refreshToken, user.ID, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	a.logger.Info("User logged in successfully", zap.String("username", username), zap.Uint("user_id", user.ID))

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * 60 * 60), // 24 hours in seconds
		User:         user,
		Scopes:       domain.ToArray(user.Scopes),
	}, nil
}

// Register implements authUsecase.
func (a *authUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {

	createdUser, err := a.userUsecase.CreateUser(ctx, req)
	if err != nil {
		a.logger.Error("Failed to create user", zap.String("username", req.Username), zap.Error(err))
		if err.Error() == "user already exists" {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to create user")
	}

	// Generate tokens
	accessToken, err := a.tokenUsecase.GenerateAccessToken(createdUser)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", createdUser.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.tokenUsecase.GenerateRefreshToken(createdUser)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", createdUser.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	// Add tokens to Redis whitelist
	if err := a.tokenUsecase.AddTokenToWhitelist(ctx, accessToken, createdUser.ID, time.Hour*24); err != nil {
		a.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenUsecase.AddTokenToWhitelist(ctx, refreshToken, createdUser.ID, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	a.logger.Info("User registered successfully", zap.String("username", createdUser.Username), zap.Uint("user_id", createdUser.ID))

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * 60 * 60), // 24 hours in seconds
		User:         createdUser,
		Scopes:       domain.ToArray(createdUser.Scopes),
	}, nil
}

// ValidateToken implements authUsecase.
func (a *authUsecase) ValidateToken(tokenString string) (*dto.Claims, error) {
	return a.tokenUsecase.ValidateToken(tokenString)
}

// RefreshToken implements authUsecase.
func (a *authUsecase) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	refreshToken := req.RefreshToken
	// Validate refresh token
	claims, err := a.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := a.userUsecase.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Generate new tokens
	newAccessToken, err := a.tokenUsecase.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := a.tokenUsecase.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// Add new tokens to whitelist and remove old refresh token
	if err := a.tokenUsecase.AddTokenToWhitelist(ctx, newAccessToken, user.ID, time.Hour*24); err != nil {
		a.logger.Error("Failed to add new access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenUsecase.AddTokenToWhitelist(ctx, newRefreshToken, user.ID, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add new refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	// Remove old refresh token from whitelist
	a.tokenUsecase.RemoveTokenFromWhitelist(ctx, refreshToken)

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * 60 * 60), // 24 hours in seconds
		User:         user,
		Scopes:       domain.ToArray(user.Scopes),
	}, nil
}

// Logout implements authUsecase.
func (a *authUsecase) Logout(ctx context.Context, userID uint) error {
	// Remove all user tokens from whitelist
	a.tokenUsecase.RemoveUserTokensFromWhitelist(ctx, userID)
	a.logger.Info("User logged out", zap.Uint("user_id", userID))
	return nil
}

// LogoutWithToken implements authUsecase - logout using specific token
func (a *authUsecase) LogoutWithToken(ctx context.Context, token string) error {
	// Remove the specific token from whitelist
	a.tokenUsecase.RemoveTokenFromWhitelist(ctx, token)
	a.logger.Info("Token revoked during logout")
	return nil
}
