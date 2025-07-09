package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"go.uber.org/zap"
)

type TokenUsecase interface {
	GenerateAccessToken(user *models.User) (string, error)
	GenerateRefreshToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*dto.Claims, error)
	ParseTokenClaims(tokenString string) (*dto.Claims, error)
	AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error
	IsTokenWhitelisted(ctx context.Context, token string) bool
	RemoveTokenFromWhitelist(ctx context.Context, token string)
	RemoveUserTokensFromWhitelist(ctx context.Context, userID uint)
}

type tokenUsecase struct {
	jwtConfig configs.JWT
	logger    *zap.Logger
	cache     db.IRedisClient
}

func NewTokenService(jwtConfig configs.JWT, logger *zap.Logger, cache db.IRedisClient) TokenUsecase {
	return &tokenUsecase{
		jwtConfig: jwtConfig,
		logger:    logger,
		cache:     cache,
	}
}

func (t *tokenUsecase) GenerateAccessToken(user *models.User) (string, error) {
	userScopes := models.ToArray(user.Scopes)
	claims := &dto.Claims{
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

func (t *tokenUsecase) GenerateRefreshToken(user *models.User) (string, error) {
	userScopes := models.ToArray(user.Scopes)
	claims := &dto.Claims{
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
func (t *tokenUsecase) ValidateToken(tokenString string) (*dto.Claims, error) {
	// Check if token is in whitelist
	if !t.IsTokenWhitelisted(context.Background(), tokenString) {
		return nil, fmt.Errorf("token is not valid or has been revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &dto.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*dto.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ParseTokenClaims parses token without validation for claims extraction
func (t *tokenUsecase) ParseTokenClaims(tokenString string) (*dto.Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &dto.Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*dto.Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// AddTokenToWhitelist adds a token to Redis whitelist with TTL
func (t *tokenUsecase) AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	cacheKey := fmt.Sprintf("user_tokens:%d", userID)
	err := t.cache.SADD(ctx, cacheKey, token)
	if err != nil {
		t.logger.Error("Failed to add token to whitelist", zap.String("token", token), zap.Error(err))
		return fmt.Errorf("failed to add token to whitelist: %w", err)
	}
	cacheKey = fmt.Sprintf("token:whitelist:%s", token)
	err = t.cache.Set(ctx, cacheKey, "valid", expiration)
	if err != nil {
		t.logger.Error("Failed to set token in whitelist", zap.String("token", token), zap.Error(err))
		return fmt.Errorf("failed to set token in whitelist: %w", err)
	}
	t.logger.Info("Token added to whitelist", zap.String("token", token), zap.Uint("user_id", userID))
	return nil
}

// IsTokenWhitelisted checks if a token exists in Redis whitelist
func (t *tokenUsecase) IsTokenWhitelisted(ctx context.Context, token string) bool {
	key := fmt.Sprintf("token:whitelist:%s", token)
	var valid string
	err := t.cache.Get(ctx, key, &valid)
	if err != nil {
		if err == db.ErrCacheMiss {
			t.logger.Warn("Token not found in whitelist", zap.String("token", token))
			return false
		}
		t.logger.Error("Failed to check token in whitelist", zap.String("token", token), zap.Error(err))
		return false
	}
	if valid == "valid" {
		t.logger.Info("Token is whitelisted", zap.String("token", token))
		return true
	}
	t.logger.Warn("Token is not valid", zap.String("token", token))
	return false
}

// RemoveTokenFromWhitelist removes a specific token from Redis whitelist
func (t *tokenUsecase) RemoveTokenFromWhitelist(ctx context.Context, token string) {
	key := fmt.Sprintf("token:whitelist:%s", token)
	t.cache.Del(ctx, key)
}

// RemoveUserTokensFromWhitelist removes all tokens for a specific user
func (t *tokenUsecase) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) {
	cacheKey := fmt.Sprintf("user_tokens:%d", userID)
	token, err := t.cache.SMEMBERS(ctx, cacheKey)
	if err != nil {
		t.logger.Error("Failed to get user tokens from whitelist", zap.Uint("user_id", userID), zap.Error(err))
		return
	}
	for _, to := range token {
		t.cache.Del(ctx, fmt.Sprintf("token:whitelist:%s", to))
	}
	if err := t.cache.Del(ctx, cacheKey); err != nil {
		t.logger.Error("Failed to delete user tokens from whitelist", zap.Uint("user_id", userID), zap.Error(err))
		return
	}
	t.logger.Info("All user tokens removed from whitelist", zap.Uint("user_id", userID))
}
