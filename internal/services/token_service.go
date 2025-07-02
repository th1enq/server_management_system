package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

type TokenService interface {
	GenerateAccessToken(user *models.User) (string, error)
	GenerateRefreshToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ParseTokenClaims(tokenString string) (*Claims, error)
	AddTokenToWhitelist(ctx context.Context, token string, expiration time.Duration) error
	IsTokenWhitelisted(ctx context.Context, token string) bool
	RemoveTokenFromWhitelist(ctx context.Context, token string)
	RemoveUserTokensFromWhitelist(ctx context.Context, userID uint)
}

type tokenService struct {
	jwtConfig configs.JWT
	logger    *zap.Logger
	cache     *redis.Client
}

func NewTokenService(jwtConfig configs.JWT, logger *zap.Logger, cache *redis.Client) TokenService {
	return &tokenService{
		jwtConfig: jwtConfig,
		logger:    logger,
		cache:     cache,
	}
}

// GenerateAccessToken generates a JWT access token for the user
func (t *tokenService) GenerateAccessToken(user *models.User) (string, error) {
	userScopes := models.GetDefaultScopes(user.Role)
	claims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Scopes:    userScopes,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.jwtConfig.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(t.jwtConfig.Secret))
}

// GenerateRefreshToken generates a JWT refresh token for the user
func (t *tokenService) GenerateRefreshToken(user *models.User) (string, error) {
	userScopes := models.GetDefaultScopes(user.Role)
	claims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Scopes:    userScopes,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.jwtConfig.Expiration * 7)), // 7x longer than access token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.jwtConfig.Secret))
}

// ValidateToken validates a JWT token and checks whitelist
func (t *tokenService) ValidateToken(tokenString string) (*Claims, error) {
	// Check if token is in whitelist
	if !t.IsTokenWhitelisted(context.Background(), tokenString) {
		return nil, fmt.Errorf("token is not valid or has been revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ParseTokenClaims parses token without validation for claims extraction
func (t *tokenService) ParseTokenClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// AddTokenToWhitelist adds a token to Redis whitelist with TTL
func (t *tokenService) AddTokenToWhitelist(ctx context.Context, token string, expiration time.Duration) error {
	if t.cache == nil {
		return nil // Skip if cache is not configured
	}

	key := fmt.Sprintf("token:whitelist:%s", token)
	return t.cache.Set(ctx, key, "valid", expiration).Err()
}

// IsTokenWhitelisted checks if a token exists in Redis whitelist
func (t *tokenService) IsTokenWhitelisted(ctx context.Context, token string) bool {
	if t.cache == nil {
		return true // Allow if cache is not configured
	}

	key := fmt.Sprintf("token:whitelist:%s", token)
	result := t.cache.Get(ctx, key)
	return result.Err() != redis.Nil
}

// RemoveTokenFromWhitelist removes a specific token from Redis whitelist
func (t *tokenService) RemoveTokenFromWhitelist(ctx context.Context, token string) {
	if t.cache == nil {
		return
	}

	key := fmt.Sprintf("token:whitelist:%s", token)
	t.cache.Del(ctx, key)
}

// RemoveUserTokensFromWhitelist removes all tokens for a specific user
func (t *tokenService) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) {
	if t.cache == nil {
		return
	}

	// Pattern to match all tokens for this user
	pattern := "token:whitelist:*"

	// Get all keys matching the pattern
	keys, err := t.cache.Keys(ctx, pattern).Result()
	if err != nil {
		t.logger.Error("Failed to get token keys for user", zap.Uint("user_id", userID), zap.Error(err))
		return
	}

	// For each token, validate and check if it belongs to the user
	for _, key := range keys {
		// Extract token from key
		token := key[len("token:whitelist:"):]

		// Parse token to get user ID
		claims, err := t.ParseTokenClaims(token)
		if err != nil {
			continue
		}

		// If token belongs to the user, remove it
		if claims.UserID == userID {
			t.cache.Del(ctx, key)
		}
	}
}
