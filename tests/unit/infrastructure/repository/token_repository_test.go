package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	repoInterFace "github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/repository"
)

// MockCacheClient is a mock implementation of CacheClient
type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	args := m.Called(ctx, key, data, ttl)
	return args.Error(0)
}

func (m *MockCacheClient) Get(ctx context.Context, key string, dest any) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheClient) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	args := m.Called(ctx, pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheClient) SADD(ctx context.Context, key string, members ...string) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCacheClient) SMEMBERS(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

type TokenRepositoryTestSuite struct {
	suite.Suite
	repo      repoInterFace.TokenRepository
	mockCache *MockCacheClient
}

func (suite *TokenRepositoryTestSuite) SetupTest() {
	suite.mockCache = new(MockCacheClient)
	suite.repo = repository.NewTokenRepository(suite.mockCache)
}

func (suite *TokenRepositoryTestSuite) TestDownTest() {
	// Verify all expectations were met
	suite.mockCache.AssertExpectations(suite.T())
}

func TestTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TokenRepositoryTestSuite))
}

func (suite *TokenRepositoryTestSuite) TestAddTokenToWhitelist() {
	ctx := context.Background()
	token := "test-token"
	userID := uint(1)
	expiration := time.Hour

	// Mock expectations
	suite.mockCache.On("SADD", ctx, "user_tokens:1", []string{token}).Return(nil)
	suite.mockCache.On("Set", ctx, "token:whitelist:test-token", "valid", expiration).Return(nil)

	err := suite.repo.AddTokenToWhitelist(ctx, token, userID, expiration)
	assert.NoError(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestAddTokenToWhitelist_SADDError() {
	ctx := context.Background()
	token := "test-token"
	userID := uint(1)
	expiration := time.Hour

	// Mock expectations - SADD fails
	suite.mockCache.On("SADD", ctx, "user_tokens:1", []string{token}).Return(assert.AnError)

	err := suite.repo.AddTokenToWhitelist(ctx, token, userID, expiration)
	assert.Error(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestAddTokenToWhitelist_SetError() {
	ctx := context.Background()
	token := "test-token"
	userID := uint(1)
	expiration := time.Hour

	// Mock expectations - Set fails
	suite.mockCache.On("SADD", ctx, "user_tokens:1", []string{token}).Return(nil)
	suite.mockCache.On("Set", ctx, "token:whitelist:test-token", "valid", expiration).Return(assert.AnError)

	err := suite.repo.AddTokenToWhitelist(ctx, token, userID, expiration)
	assert.Error(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestIsTokenWhitelisted_ValidToken() {
	ctx := context.Background()
	token := "test-token"

	// Mock expectations - token is valid
	suite.mockCache.On("Get", ctx, "token:whitelist:test-token", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*string)
		*dest = "valid"
	})

	result := suite.repo.IsTokenWhitelisted(ctx, token)
	assert.True(suite.T(), result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestIsTokenWhitelisted_InvalidToken() {
	ctx := context.Background()
	token := "test-token"

	// Mock expectations - token not found
	suite.mockCache.On("Get", ctx, "token:whitelist:test-token", mock.AnythingOfType("*string")).Return(assert.AnError)

	result := suite.repo.IsTokenWhitelisted(ctx, token)
	assert.False(suite.T(), result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestIsTokenWhitelisted_InvalidValue() {
	ctx := context.Background()
	token := "test-token"

	// Mock expectations - token exists but has invalid value
	suite.mockCache.On("Get", ctx, "token:whitelist:test-token", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*string)
		*dest = "invalid"
	})

	result := suite.repo.IsTokenWhitelisted(ctx, token)
	assert.False(suite.T(), result)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestRemoveTokenFromWhitelist() {
	ctx := context.Background()
	token := "test-token"

	// Mock expectations
	suite.mockCache.On("Del", ctx, "token:whitelist:test-token").Return(nil)

	err := suite.repo.RemoveTokenFromWhitelist(ctx, token)
	assert.NoError(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestRemoveTokenFromWhitelist_Error() {
	ctx := context.Background()
	token := "test-token"

	// Mock expectations - Del fails
	suite.mockCache.On("Del", ctx, "token:whitelist:test-token").Return(assert.AnError)

	err := suite.repo.RemoveTokenFromWhitelist(ctx, token)
	assert.Error(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestRemoveUserTokensFromWhitelist() {
	ctx := context.Background()
	userID := uint(1)
	tokens := []string{"token1", "token2", "token3"}

	// Mock expectations
	suite.mockCache.On("SMEMBERS", ctx, "user_tokens:1").Return(tokens, nil)
	suite.mockCache.On("Del", ctx, "token:whitelist:token1").Return(nil)
	suite.mockCache.On("Del", ctx, "token:whitelist:token2").Return(nil)
	suite.mockCache.On("Del", ctx, "token:whitelist:token3").Return(nil)
	suite.mockCache.On("Del", ctx, "user_tokens:1").Return(nil)

	err := suite.repo.RemoveUserTokensFromWhitelist(ctx, userID)
	assert.NoError(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestRemoveUserTokensFromWhitelist_SMEMBERSError() {
	ctx := context.Background()
	userID := uint(1)

	// Mock expectations - SMEMBERS fails
	suite.mockCache.On("SMEMBERS", ctx, "user_tokens:1").Return([]string{}, assert.AnError)

	err := suite.repo.RemoveUserTokensFromWhitelist(ctx, userID)
	assert.Error(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestRemoveUserTokensFromWhitelist_DelUserTokensError() {
	ctx := context.Background()
	userID := uint(1)
	tokens := []string{"token1", "token2"}

	// Mock expectations - final Del fails
	suite.mockCache.On("SMEMBERS", ctx, "user_tokens:1").Return(tokens, nil)
	suite.mockCache.On("Del", ctx, "token:whitelist:token1").Return(nil)
	suite.mockCache.On("Del", ctx, "token:whitelist:token2").Return(nil)
	suite.mockCache.On("Del", ctx, "user_tokens:1").Return(assert.AnError)

	err := suite.repo.RemoveUserTokensFromWhitelist(ctx, userID)
	assert.Error(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *TokenRepositoryTestSuite) TestRemoveUserTokensFromWhitelist_EmptyTokens() {
	ctx := context.Background()
	userID := uint(1)
	tokens := []string{}

	// Mock expectations - no tokens to remove
	suite.mockCache.On("SMEMBERS", ctx, "user_tokens:1").Return(tokens, nil)
	suite.mockCache.On("Del", ctx, "user_tokens:1").Return(nil)

	err := suite.repo.RemoveUserTokensFromWhitelist(ctx, userID)
	assert.NoError(suite.T(), err)

	suite.mockCache.AssertExpectations(suite.T())
}
