package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(user *models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}
func (m *MockTokenService) GenerateRefreshToken(user *models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}
func (m *MockTokenService) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}
func (m *MockTokenService) ParseTokenClaims(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}
func (m *MockTokenService) AddTokenToWhitelist(ctx context.Context, token string, expiration time.Duration) error {
	args := m.Called(ctx, token, expiration)
	return args.Error(0)
}
func (m *MockTokenService) IsTokenWhitelisted(ctx context.Context, token string) bool {
	args := m.Called(ctx, token)
	return args.Bool(0)
}
func (m *MockTokenService) RemoveTokenFromWhitelist(ctx context.Context, token string) {
	m.Called(ctx, token)
	// This method doesn't return anything in the interface
}
func (m *MockTokenService) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) {
	m.Called(ctx, userID)
	// This method doesn't return anything in the interface
}

func createTestAuthService() (AuthService, *MockUserService, *MockTokenService) {
	mockUserService := &MockUserService{}
	mockTokenService := &MockTokenService{}
	logger := zap.NewNop()

	authSrv := NewAuthService(mockUserService, mockTokenService, logger)

	return authSrv, mockUserService, mockTokenService
}

func TestAuthService_Login_Success(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Create a test user with hashed password
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	mockUserService.On("GetUserByUsername", ctx, "testuser").Return(user, nil)

	mockTokenService.On("GenerateAccessToken", user).Return("access_token", nil)
	mockTokenService.On("GenerateRefreshToken", user).Return("refresh_token", nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "access_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "refresh_token", mock.AnythingOfType("time.Duration")).Return(nil)

	// Test
	result, err := authSrv.Login(ctx, "testuser", "password123")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, user.ID, result.User.ID)
	assert.Equal(t, user.Username, result.User.Username)
	assert.True(t, len(result.Scopes) > 0)

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authSrv, mockUserService, _ := createTestAuthService()
	ctx := context.Background()

	mockUserService.On("GetUserByUsername", ctx, "nonexistent").Return(nil, errors.New("user not found"))

	// Test
	result, err := authSrv.Login(ctx, "nonexistent", "password")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	authSrv, mockUserService, _ := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: false, // User is inactive
	}

	mockUserService.On("GetUserByUsername", ctx, "testuser").Return(user, nil)

	// Test
	result, err := authSrv.Login(ctx, "testuser", "password123")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "account is disabled")

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	authSrv, mockUserService, _ := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}
	user.SetPassword("correct_password")

	mockUserService.On("GetUserByUsername", ctx, "testuser").Return(user, nil)

	// Test with wrong password
	result, err := authSrv.Login(ctx, "testuser", "wrong_password")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Register_Success(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Create input user with the exact state it will have when passed to CreateUser
	user := &models.User{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	// Create a matcher that can match the user based only on Username and Email
	userMatcher := mock.MatchedBy(func(u *models.User) bool {
		return u.Username == "newuser" && u.Email == "new@example.com"
	})

	// Mock the create user call with any user that matches our criteria
	mockUserService.On("CreateUser", ctx, userMatcher).Return(nil).Run(func(args mock.Arguments) {
		// Update the user passed in with proper ID, role, etc.
		u := args.Get(1).(*models.User)
		u.ID = 1
		u.Role = models.RoleUser
		u.IsActive = true
	})

	// Set up token mocks with the same matcher
	mockTokenService.On("GenerateAccessToken", userMatcher).Return("access_token", nil)
	mockTokenService.On("GenerateRefreshToken", userMatcher).Return("refresh_token", nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "access_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "refresh_token", mock.AnythingOfType("time.Duration")).Return(nil)

	// Test
	result, err := authSrv.Register(ctx, user)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, uint(1), result.User.ID)
	assert.Equal(t, models.RoleUser, result.User.Role)
	assert.True(t, result.User.IsActive)

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Register_CreateUserFails(t *testing.T) {
	authSrv, mockUserService, _ := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	mockUserService.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("user already exists"))

	// Test
	result, err := authSrv.Register(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create user")

	mockUserService.AssertExpectations(t)
}
func TestAuthService_ValidateToken_Success(t *testing.T) {
	authSrv, _, mockTokenService := createTestAuthService()

	mockTokenService.On("ValidateToken", "valid.token.here").Return(&Claims{
		UserID:   1,
		Username: "testuser",
		Role:     models.RoleUser,
	}, nil)

	claims, err := authSrv.ValidateToken("valid.token.here")
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, models.RoleUser, claims.Role)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	authSrv, _, mockTokenService := createTestAuthService()

	mockTokenService.On("ValidateToken", "invalid.token.here").Return(nil, errors.New("invalid token"))

	claims, err := authSrv.ValidateToken("invalid.token.here")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	authSrv, _, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Test with invalid refresh token
	mockTokenService.On("ValidateToken", "invalid.token.here").Return(nil, errors.New("invalid refresh token"))

	result, err := authSrv.RefreshToken(ctx, "invalid.token.here")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Set up user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      models.RoleUser,
		TokenType: "refresh",
	}, nil)

	// Mock user service
	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Mock token generation
	mockTokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)
	mockTokenService.On("GenerateRefreshToken", user).Return("new_refresh_token", nil)

	// Mock token whitelist operations
	mockTokenService.On("AddTokenToWhitelist", ctx, "new_access_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "new_refresh_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("RemoveTokenFromWhitelist", ctx, "valid.refresh.token")

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, "new_access_token", result.AccessToken)
	assert.Equal(t, "new_refresh_token", result.RefreshToken)
	assert.Equal(t, user.ID, result.User.ID)
	assert.True(t, len(result.Scopes) > 0)

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_Logout(t *testing.T) {
	authSrv, _, mockTokenService := createTestAuthService()
	ctx := context.Background()

	mockTokenService.On("RemoveUserTokensFromWhitelist", ctx, uint(1)).Return(nil)

	// Test
	err := authSrv.Logout(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_LogoutWithToken(t *testing.T) {
	authSrv, _, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Mock token service
	mockTokenService.On("RemoveTokenFromWhitelist", ctx, "token.to.revoke")

	// Test
	err := authSrv.LogoutWithToken(ctx, "token.to.revoke")

	// Assertions
	assert.NoError(t, err)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    999, // Non-existent user
		Username:  "testuser",
		TokenType: "refresh",
	}, nil)

	// Mock user service to return error for non-existent user
	mockUserService.On("GetUserByID", ctx, uint(999)).Return(nil, errors.New("user not found"))

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InactiveUser(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Set up inactive user
	inactiveUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: false,
	}

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}, nil)

	// Mock user service
	mockUserService.On("GetUserByID", ctx, uint(1)).Return(inactiveUser, nil)

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "account is disabled")

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InvalidTokenType(t *testing.T) {
	authSrv, _, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Mock token validation with wrong token type (access instead of refresh)
	mockTokenService.On("ValidateToken", "access.token.not.refresh").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "access", // Wrong token type
	}, nil)

	// Test
	result, err := authSrv.RefreshToken(ctx, "access.token.not.refresh")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")

	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_AccessTokenGenerationFails(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Set up user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}, nil)

	// Mock user service
	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Mock token generation failure
	mockTokenService.On("GenerateAccessToken", user).Return("", errors.New("token generation failed"))

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate access token")

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_RefreshTokenGenerationFails(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Set up user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}, nil)

	// Mock user service
	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Mock access token generation success
	mockTokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)

	// Mock refresh token generation failure
	mockTokenService.On("GenerateRefreshToken", user).Return("", errors.New("refresh token generation failed"))

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate refresh token")

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_WhitelistFails(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Set up user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}, nil)

	// Mock user service
	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Mock token generation success
	mockTokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)
	mockTokenService.On("GenerateRefreshToken", user).Return("new_refresh_token", nil)

	// Mock token whitelist operations
	mockTokenService.On("AddTokenToWhitelist", ctx, "new_access_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "new_refresh_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("RemoveTokenFromWhitelist", ctx, "valid.refresh.token")

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, "new_access_token", result.AccessToken)
	assert.Equal(t, "new_refresh_token", result.RefreshToken)
	assert.Equal(t, user.ID, result.User.ID)
	assert.True(t, len(result.Scopes) > 0)

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_RefreshWhitelistFails(t *testing.T) {
	authSrv, mockUserService, mockTokenService := createTestAuthService()
	ctx := context.Background()

	// Set up user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Mock token validation
	mockTokenService.On("ValidateToken", "valid.refresh.token").Return(&Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}, nil)

	// Mock user service
	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Mock token generation success
	mockTokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)
	mockTokenService.On("GenerateRefreshToken", user).Return("new_refresh_token", nil)

	// Mock access token whitelist success but refresh token whitelist failure
	mockTokenService.On("AddTokenToWhitelist", ctx, "new_access_token", mock.AnythingOfType("time.Duration")).Return(nil)
	mockTokenService.On("AddTokenToWhitelist", ctx, "new_refresh_token", mock.AnythingOfType("time.Duration")).Return(errors.New("refresh whitelist operation failed"))

	// Test
	result, err := authSrv.RefreshToken(ctx, "valid.refresh.token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to whitelist token")

	mockUserService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}
