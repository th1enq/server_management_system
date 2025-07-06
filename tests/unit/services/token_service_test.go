package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type TokenServiceTestSuite struct {
	suite.Suite
	tokenService services.ITokenService
	mockCache    MockCacheClient
}

func (suite *TokenServiceTestSuite) SetupTest() {
	suite.mockCache = MockCacheClient{}
	cfg := configs.JWT{
		Secret:     "test-secret-key-for-jwt-token-generation",
		Expiration: time.Hour, // Set a proper expiration time
	}
	suite.tokenService = services.NewTokenService(cfg, zap.NewNop(), &suite.mockCache)
}

func TestTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TokenServiceTestSuite))
}

func (suite *TokenServiceTestSuite) TestGenerateAccessToken() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}
	token, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)
	suite.NotEmpty(token)
	claims, err := suite.tokenService.ParseTokenClaims(token)
	suite.NoError(err)
	suite.Equal(user.ID, claims.UserID)
	suite.Equal(user.Username, claims.Username)
	suite.Equal(user.Email, claims.Email)
	suite.Equal(user.Role, claims.Role)
	suite.ElementsMatch(user.Scopes, claims.Scopes)
}

func (suite *TokenServiceTestSuite) TestGenerateRefreshToken() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}
	token, err := suite.tokenService.GenerateRefreshToken(user)
	suite.NoError(err)
	suite.NotEmpty(token)
	claims, err := suite.tokenService.ParseTokenClaims(token)
	suite.NoError(err)
	suite.Equal(user.ID, claims.UserID)
	suite.Equal(user.Username, claims.Username)
	suite.Equal(user.Email, claims.Email)
	suite.Equal(user.Role, claims.Role)
	suite.ElementsMatch(user.Scopes, claims.Scopes)
}

func (suite *TokenServiceTestSuite) TestValidateToken() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}
	token, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)

	// Mock the whitelist check to return "valid"
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*string)
		*arg = "valid"
	}).Return(nil)

	claims, err := suite.tokenService.ValidateToken(token)
	suite.NoError(err)
	suite.NotNil(claims)
	suite.Equal(user.ID, claims.UserID)
	suite.Equal(user.Username, claims.Username)
	suite.Equal(user.Email, claims.Email)
	suite.Equal(user.Role, claims.Role)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestValidateTokenInvalid() {
	invalidToken := "invalid.token.string"

	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

	claims, err := suite.tokenService.ValidateToken(invalidToken)
	suite.Error(err)
	suite.Nil(claims)
	suite.Contains(err.Error(), "token is not valid or has been revoked")
}

func (suite *TokenServiceTestSuite) TestValidateTokenNotWhitelisted() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}
	token, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)

	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(db.ErrCacheMiss)

	claims, err := suite.tokenService.ValidateToken(token)
	suite.Error(err)
	suite.Nil(claims)
	suite.Contains(err.Error(), "token is not valid or has been revoked")
}

func (suite *TokenServiceTestSuite) TestValidateTokenWhitelisted() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}
	token, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)

	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*string)
		*arg = "valid"
	}).Return(nil)

	claims, err := suite.tokenService.ValidateToken(token)
	suite.NoError(err)
	suite.NotNil(claims)
	suite.Equal(user.ID, claims.UserID)
	suite.Equal(user.Username, claims.Username)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestParseTokenClaimsInvalid() {
	invalidToken := "invalid.token.string"

	claims, err := suite.tokenService.ParseTokenClaims(invalidToken)
	suite.Error(err)
	suite.Nil(claims)
}

func (suite *TokenServiceTestSuite) TestAddTokenToWhitelist() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}
	token, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)

	suite.mockCache.On("SADD", mock.Anything, "user_tokens:1", mock.AnythingOfType("[]string")).Return(nil)
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), "valid", mock.AnythingOfType("time.Duration")).Return(nil)

	err = suite.tokenService.AddTokenToWhitelist(context.Background(), token, user.ID, time.Hour)
	suite.NoError(err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestAddTokenToWhitelistSADDFail() {
	token := "test.token.string"
	userID := uint(1)

	suite.mockCache.On("SADD", mock.Anything, "user_tokens:1", mock.AnythingOfType("[]string")).Return(errors.New("redis error"))

	err := suite.tokenService.AddTokenToWhitelist(context.Background(), token, userID, time.Hour)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to add token to whitelist")

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestAddTokenToWhitelistSetFail() {
	token := "test.token.string"
	userID := uint(1)

	suite.mockCache.On("SADD", mock.Anything, "user_tokens:1", mock.AnythingOfType("[]string")).Return(nil)
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), "valid", mock.AnythingOfType("time.Duration")).Return(errors.New("redis error"))

	err := suite.tokenService.AddTokenToWhitelist(context.Background(), token, userID, time.Hour)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to set token in whitelist")

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestIsTokenWhitelisted() {
	token := "test.token.string"

	suite.mockCache.On("Get", mock.Anything, "token:whitelist:test.token.string", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*string)
		*arg = "valid"
	}).Return(nil)

	result := suite.tokenService.IsTokenWhitelisted(context.Background(), token)
	suite.True(result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestIsTokenWhitelistedNotFound() {
	token := "test.token.string"

	suite.mockCache.On("Get", mock.Anything, "token:whitelist:test.token.string", mock.Anything).Return(db.ErrCacheMiss)

	result := suite.tokenService.IsTokenWhitelisted(context.Background(), token)
	suite.False(result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestIsTokenWhitelistedError() {
	token := "test.token.string"

	suite.mockCache.On("Get", mock.Anything, "token:whitelist:test.token.string", mock.Anything).Return(errors.New("redis error"))

	result := suite.tokenService.IsTokenWhitelisted(context.Background(), token)
	suite.False(result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestIsTokenWhitelistedInvalidValue() {
	token := "test.token.string"

	suite.mockCache.On("Get", mock.Anything, "token:whitelist:test.token.string", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*string)
		*arg = "invalid"
	}).Return(nil)

	result := suite.tokenService.IsTokenWhitelisted(context.Background(), token)
	suite.False(result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestRemoveTokenFromWhitelist() {
	token := "test.token.string"

	suite.mockCache.On("Del", mock.Anything, "token:whitelist:test.token.string").Return(nil)

	suite.tokenService.RemoveTokenFromWhitelist(context.Background(), token)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestRemoveUserTokensFromWhitelist() {
	userID := uint(1)
	tokens := []string{"token1", "token2", "token3"}

	suite.mockCache.On("SMEMBERS", mock.Anything, "user_tokens:1").Return(tokens, nil)
	suite.mockCache.On("Del", mock.Anything, "token:whitelist:token1").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "token:whitelist:token2").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "token:whitelist:token3").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "user_tokens:1").Return(nil)

	suite.tokenService.RemoveUserTokensFromWhitelist(context.Background(), userID)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestRemoveUserTokensFromWhitelistSMEMBERSError() {
	userID := uint(1)

	suite.mockCache.On("SMEMBERS", mock.Anything, "user_tokens:1").Return(nil, errors.New("redis error"))

	suite.tokenService.RemoveUserTokensFromWhitelist(context.Background(), userID)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestRemoveUserTokensFromWhitelistDelError() {
	userID := uint(1)
	tokens := []string{"token1", "token2"}

	suite.mockCache.On("SMEMBERS", mock.Anything, "user_tokens:1").Return(tokens, nil)
	suite.mockCache.On("Del", mock.Anything, "token:whitelist:token1").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "token:whitelist:token2").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "user_tokens:1").Return(errors.New("redis error"))

	suite.tokenService.RemoveUserTokensFromWhitelist(context.Background(), userID)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenServiceTestSuite) TestGenerateTokenWithDifferentScopes() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "admin",
		Scopes:   models.ToBitmask([]models.APIScope{"server:read", "server:write", "user:read"}),
	}

	accessToken, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)
	suite.NotEmpty(accessToken)

	refreshToken, err := suite.tokenService.GenerateRefreshToken(user)
	suite.NoError(err)
	suite.NotEmpty(refreshToken)

	accessClaims, err := suite.tokenService.ParseTokenClaims(accessToken)
	suite.NoError(err)
	suite.Equal("access", accessClaims.TokenType)
	suite.Equal(user.Role, accessClaims.Role)
	suite.Len(accessClaims.Scopes, 3)

	refreshClaims, err := suite.tokenService.ParseTokenClaims(refreshToken)
	suite.NoError(err)
	suite.Equal("refresh", refreshClaims.TokenType)
	suite.Equal(user.Role, refreshClaims.Role)
	suite.Len(refreshClaims.Scopes, 3)
}

func (suite *TokenServiceTestSuite) TestTokenExpirationDifference() {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   0,
	}

	accessToken, err := suite.tokenService.GenerateAccessToken(user)
	suite.NoError(err)

	refreshToken, err := suite.tokenService.GenerateRefreshToken(user)
	suite.NoError(err)

	accessClaims, err := suite.tokenService.ParseTokenClaims(accessToken)
	suite.NoError(err)

	refreshClaims, err := suite.tokenService.ParseTokenClaims(refreshToken)
	suite.NoError(err)

	// Refresh token should expire later than access token (7x longer)
	accessExp := accessClaims.ExpiresAt.Time
	refreshExp := refreshClaims.ExpiresAt.Time

	suite.True(refreshExp.After(accessExp), "Refresh token should expire later than access token")

	// The difference should be approximately 6 times the access token duration (7x - 1x = 6x)
	expectedDiff := time.Hour * 6 // 6 hours difference since refresh is 7x longer
	actualDiff := refreshExp.Sub(accessExp)

	// Allow some tolerance for execution time
	suite.True(actualDiff >= expectedDiff-time.Second && actualDiff <= expectedDiff+time.Second,
		"Expected difference of ~6 hours, got %v", actualDiff)
}
