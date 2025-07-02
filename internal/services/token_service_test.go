package services

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

func setupTokenService() (TokenService, *miniredis.Miniredis) {
	// Create a mock Redis server
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	// Create a Redis client using the mock server
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create JWT config
	jwtConfig := configs.JWT{
		Secret:     "test-secret-key",
		Expiration: time.Hour,
	}

	// Create logger
	logger := zap.NewNop()

	// Create token service
	tokenService := NewTokenService(jwtConfig, logger, redisClient)

	return tokenService, mr
}

func TestTokenService_GenerateAccessToken(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Test
	token, err := tokenService.GenerateAccessToken(user)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := tokenService.ParseTokenClaims(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "access", claims.TokenType)
}

func TestTokenService_GenerateRefreshToken(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Test
	token, err := tokenService.GenerateRefreshToken(user)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := tokenService.ParseTokenClaims(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestTokenService_ValidateToken_Success(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Generate token
	token, err := tokenService.GenerateAccessToken(user)
	require.NoError(t, err)

	// Add token to whitelist
	err = tokenService.AddTokenToWhitelist(context.Background(), token, time.Hour)
	require.NoError(t, err)

	// Test
	claims, err := tokenService.ValidateToken(token)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestTokenService_ValidateToken_NotWhitelisted(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Generate token without adding to whitelist
	token, err := tokenService.GenerateAccessToken(user)
	require.NoError(t, err)

	// Test
	claims, err := tokenService.ValidateToken(token)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is not valid or has been revoked")
}

func TestTokenService_ValidateToken_InvalidToken(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	// Test with invalid token
	claims, err := tokenService.ValidateToken("invalid.token.format")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ParseTokenClaims_Success(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Generate token
	token, err := tokenService.GenerateAccessToken(user)
	require.NoError(t, err)

	// Test
	claims, err := tokenService.ParseTokenClaims(token)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestTokenService_ParseTokenClaims_InvalidToken(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	// Test with invalid token
	claims, err := tokenService.ParseTokenClaims("invalid.token.format")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_AddTokenToWhitelist(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	ctx := context.Background()
	token := "test.token.value"

	// Test
	err := tokenService.AddTokenToWhitelist(ctx, token, time.Hour)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, tokenService.IsTokenWhitelisted(ctx, token))
}

func TestTokenService_IsTokenWhitelisted(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	ctx := context.Background()
	token := "test.token.value"
	nonWhitelistedToken := "non.whitelisted.token"

	// Add token to whitelist
	err := tokenService.AddTokenToWhitelist(ctx, token, time.Hour)
	require.NoError(t, err)

	// Test
	isWhitelisted := tokenService.IsTokenWhitelisted(ctx, token)
	isNonWhitelisted := tokenService.IsTokenWhitelisted(ctx, nonWhitelistedToken)

	// Assertions
	assert.True(t, isWhitelisted)
	assert.False(t, isNonWhitelisted)
}

func TestTokenService_RemoveTokenFromWhitelist(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()

	ctx := context.Background()
	token := "test.token.value"

	// Add token to whitelist
	err := tokenService.AddTokenToWhitelist(ctx, token, time.Hour)
	require.NoError(t, err)

	// Verify token is whitelisted
	assert.True(t, tokenService.IsTokenWhitelisted(ctx, token))

	// Test
	tokenService.RemoveTokenFromWhitelist(ctx, token)

	// Assertions
	assert.False(t, tokenService.IsTokenWhitelisted(ctx, token))
}

func TestTokenService_RemoveUserTokensFromWhitelist(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()
	ctx := context.Background()

	user1 := &models.User{ID: 1, Username: "user1", Email: "user1@example.com", Role: models.RoleUser}
	user2 := &models.User{ID: 2, Username: "user2", Email: "user2@example.com", Role: models.RoleUser}

	// Generate and whitelist tokens for users
	token1, err := tokenService.GenerateAccessToken(user1)
	require.NoError(t, err)
	token2, err := tokenService.GenerateAccessToken(user1) // Another token for user1
	require.NoError(t, err)
	token3, err := tokenService.GenerateAccessToken(user2) // Token for user2
	require.NoError(t, err)

	// Add tokens to whitelist
	err = tokenService.AddTokenToWhitelist(ctx, token1, time.Hour)
	require.NoError(t, err)
	err = tokenService.AddTokenToWhitelist(ctx, token2, time.Hour)
	require.NoError(t, err)
	err = tokenService.AddTokenToWhitelist(ctx, token3, time.Hour)
	require.NoError(t, err)

	// Test
	tokenService.RemoveUserTokensFromWhitelist(ctx, user1.ID)

	// Assertions
	// User1's tokens should be removed
	assert.False(t, tokenService.IsTokenWhitelisted(ctx, token1))
	assert.False(t, tokenService.IsTokenWhitelisted(ctx, token2))
	// User2's token should still be whitelisted
	assert.True(t, tokenService.IsTokenWhitelisted(ctx, token3))
}

func TestTokenService_TokenExpiry(t *testing.T) {
	// Setup
	tokenService, mr := setupTokenService()
	defer mr.Close()
	ctx := context.Background()

	token := "test.token.value"

	// Add token with short expiry
	shortExpiry := time.Millisecond * 100
	err := tokenService.AddTokenToWhitelist(ctx, token, shortExpiry)
	require.NoError(t, err)

	// Token should be whitelisted initially
	assert.True(t, tokenService.IsTokenWhitelisted(ctx, token))

	// Force key expiry in miniredis
	mr.FastForward(shortExpiry * 2)

	// Token should no longer be whitelisted
	assert.False(t, tokenService.IsTokenWhitelisted(ctx, token))
}
