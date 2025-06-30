package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*AuthResponse, error)
	Register(ctx context.Context, user *models.User) (*AuthResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID uint) error
	LogoutWithToken(ctx context.Context, token string) error
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *models.User `json:"user"`
	Scopes       []string     `json:"scopes"`
}

type Claims struct {
	UserID    uint              `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Role      models.UserRole   `json:"role"`
	Scopes    []models.APIScope `json:"scopes"`
	TokenType string            `json:"token_type"`
	jwt.RegisteredClaims
}

type authService struct {
	userService  UserService
	tokenService TokenService
	logger       *zap.Logger
}

func NewAuthService(userService UserService, tokenService TokenService, logger *zap.Logger) AuthService {
	return &authService{
		userService:  userService,
		tokenService: tokenService,
		logger:       logger,
	}
}

// Login implements AuthService.
func (a *authService) Login(ctx context.Context, username, password string) (*AuthResponse, error) {
	// Get user by username
	user, err := a.userService.GetUserByUsername(ctx, username)
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
	accessToken, err := a.tokenService.GenerateAccessToken(user)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.tokenService.GenerateRefreshToken(user)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	// Add tokens to Redis whitelist
	if err := a.tokenService.AddTokenToWhitelist(ctx, accessToken, time.Hour*24); err != nil {
		a.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenService.AddTokenToWhitelist(ctx, refreshToken, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	a.logger.Info("User logged in successfully", zap.String("username", username), zap.Uint("user_id", user.ID))

	// Get user scopes
	userScopes := models.GetDefaultScopes(user.Role)
	scopeStrings := make([]string, len(userScopes))
	for i, scope := range userScopes {
		scopeStrings[i] = string(scope)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * 60 * 60), // 24 hours in seconds
		User:         user,
		Scopes:       scopeStrings,
	}, nil
}

// Register implements AuthService.
func (a *authService) Register(ctx context.Context, user *models.User) (*AuthResponse, error) {
	// Create user
	if err := a.userService.CreateUser(ctx, user); err != nil {
		a.logger.Error("Failed to create user", zap.String("username", user.Username), zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := a.tokenService.GenerateAccessToken(user)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.tokenService.GenerateRefreshToken(user)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	// Add tokens to Redis whitelist
	if err := a.tokenService.AddTokenToWhitelist(ctx, accessToken, time.Hour*24); err != nil {
		a.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenService.AddTokenToWhitelist(ctx, refreshToken, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	a.logger.Info("User registered successfully", zap.String("username", user.Username), zap.Uint("user_id", user.ID))

	// Get user scopes
	userScopes := models.GetDefaultScopes(user.Role)
	scopeStrings := make([]string, len(userScopes))
	for i, scope := range userScopes {
		scopeStrings[i] = string(scope)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * 60 * 60), // 24 hours in seconds
		User:         user,
		Scopes:       scopeStrings,
	}, nil
}

// ValidateToken implements AuthService.
func (a *authService) ValidateToken(tokenString string) (*Claims, error) {
	return a.tokenService.ValidateToken(tokenString)
}

// RefreshToken implements AuthService.
func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := a.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := a.userService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Generate new tokens
	newAccessToken, err := a.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := a.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// Add new tokens to whitelist and remove old refresh token
	if err := a.tokenService.AddTokenToWhitelist(ctx, newAccessToken, time.Hour*24); err != nil {
		a.logger.Error("Failed to add new access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := a.tokenService.AddTokenToWhitelist(ctx, newRefreshToken, time.Hour*24*7); err != nil {
		a.logger.Error("Failed to add new refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	// Remove old refresh token from whitelist
	a.tokenService.RemoveTokenFromWhitelist(ctx, refreshToken)

	// Get user scopes
	userScopes := models.GetDefaultScopes(user.Role)
	scopeStrings := make([]string, len(userScopes))
	for i, scope := range userScopes {
		scopeStrings[i] = string(scope)
	}

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * 60 * 60), // 24 hours in seconds
		User:         user,
		Scopes:       scopeStrings,
	}, nil
}

// Logout implements AuthService.
func (a *authService) Logout(ctx context.Context, userID uint) error {
	// Remove all user tokens from whitelist
	a.tokenService.RemoveUserTokensFromWhitelist(ctx, userID)
	a.logger.Info("User logged out", zap.Uint("user_id", userID))
	return nil
}

// LogoutWithToken implements AuthService - logout using specific token
func (a *authService) LogoutWithToken(ctx context.Context, token string) error {
	// Remove the specific token from whitelist
	a.tokenService.RemoveTokenFromWhitelist(ctx, token)
	a.logger.Info("Token revoked during logout", zap.String("token_prefix", token[:min(20, len(token))]))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
