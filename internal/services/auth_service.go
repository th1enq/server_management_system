package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*AuthResponse, error)
	Register(ctx context.Context, user *models.User) (*AuthResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID uint) error
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
	UserID   uint              `json:"user_id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Role     models.UserRole   `json:"role"`
	Scopes   []models.APIScope `json:"scopes"`
	jwt.RegisteredClaims
}

type authService struct {
	userService UserService
	jwtConfig   configs.JWT
	logger      *zap.Logger
}

func NewAuthService(userService UserService, jwtConfig configs.JWT, logger *zap.Logger) AuthService {
	return &authService{
		userService: userService,
		jwtConfig:   jwtConfig,
		logger:      logger,
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

	// Update last login
	if err := a.userService.UpdateLastLogin(ctx, user.ID); err != nil {
		a.logger.Error("Failed to update last login", zap.Uint("user_id", user.ID), zap.Error(err))
	}

	// Generate tokens
	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.generateRefreshToken(user)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
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
		ExpiresIn:    int64(a.jwtConfig.Expiration.Seconds()),
		User:         user,
		Scopes:       scopeStrings,
	}, nil
}

// Register implements AuthService.
func (a *authService) Register(ctx context.Context, user *models.User) (*AuthResponse, error) {
	// Set default values
	if user.Role == "" {
		user.Role = models.RoleUser
	}
	user.IsActive = true

	// Create user
	if err := a.userService.CreateUser(ctx, user); err != nil {
		a.logger.Error("Failed to create user", zap.String("username", user.Username), zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get the created user (to get the ID)
	createdUser, err := a.userService.GetUserByUsername(ctx, user.Username)
	if err != nil {
		a.logger.Error("Failed to get created user", zap.String("username", user.Username), zap.Error(err))
		return nil, fmt.Errorf("failed to get created user")
	}

	// Generate tokens
	accessToken, err := a.generateAccessToken(createdUser)
	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Uint("user_id", createdUser.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	refreshToken, err := a.generateRefreshToken(createdUser)
	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Uint("user_id", createdUser.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	a.logger.Info("User registered successfully", zap.String("username", user.Username), zap.Uint("user_id", createdUser.ID))

	// Get user scopes
	userScopes := models.GetDefaultScopes(createdUser.Role)
	scopeStrings := make([]string, len(userScopes))
	for i, scope := range userScopes {
		scopeStrings[i] = string(scope)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(a.jwtConfig.Expiration.Seconds()),
		User:         createdUser,
		Scopes:       scopeStrings,
	}, nil
}

// ValidateToken implements AuthService.
func (a *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken implements AuthService.
func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := a.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
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
	newAccessToken, err := a.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := a.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

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
		ExpiresIn:    int64(a.jwtConfig.Expiration.Seconds()),
		User:         user,
		Scopes:       scopeStrings,
	}, nil
}

// Logout implements AuthService.
func (a *authService) Logout(ctx context.Context, userID uint) error {
	// In a real implementation, you might want to blacklist the token
	// For now, we'll just log the logout
	a.logger.Info("User logged out", zap.Uint("user_id", userID))
	return nil
}

// generateAccessToken generates a JWT access token for the user
func (a *authService) generateAccessToken(user *models.User) (string, error) {
	userScopes := models.GetDefaultScopes(user.Role)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Scopes:   userScopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.jwtConfig.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtConfig.Secret))
}

// generateRefreshToken generates a JWT refresh token for the user
func (a *authService) generateRefreshToken(user *models.User) (string, error) {
	userScopes := models.GetDefaultScopes(user.Role)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Scopes:   userScopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.jwtConfig.Expiration * 7)), // 7x longer than access token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtConfig.Secret))
}
