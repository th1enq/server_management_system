package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/configs"
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

func createTestAuthService() (*authService, *MockUserService) {
	mockUserService := &MockUserService{}
	logger := zap.NewNop()

	authSrv := &authService{
		userService: mockUserService,
		logger:      logger,
	}

	return authSrv, mockUserService
}

func TestAuthService_Login_Success(t *testing.T) {
	authSrv, mockUserService := createTestAuthService()
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
	authSrv, mockUserService := createTestAuthService()
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
	authSrv, mockUserService := createTestAuthService()
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
	authSrv, mockUserService := createTestAuthService()
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
	authSrv, mockUserService := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	createdUser := &models.User{
		ID:       1,
		Username: "newuser",
		Email:    "new@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}
	createdUser.SetPassword("password123")

	mockUserService.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)
	mockUserService.On("GetUserByUsername", ctx, "newuser").Return(createdUser, nil)

	// Test
	result, err := authSrv.Register(ctx, user)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, createdUser.ID, result.User.ID)
	assert.Equal(t, models.RoleUser, result.User.Role)
	assert.True(t, result.User.IsActive)

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Register_CreateUserFails(t *testing.T) {
	authSrv, mockUserService := createTestAuthService()
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
	authSrv, _ := createTestAuthService()

	// Create a test user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Generate a token
	token, err := authSrv.generateAccessToken(user)
	assert.NoError(t, err)

	// Test
	claims, err := authSrv.ValidateToken(token)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	authSrv, _ := createTestAuthService()

	// Test with invalid token
	claims, err := authSrv.ValidateToken("invalid.token.here")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	authSrv, mockUserService := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Generate a refresh token
	refreshToken, err := authSrv.generateRefreshToken(user)
	assert.NoError(t, err)

	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Test
	result, err := authSrv.RefreshToken(ctx, refreshToken)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, user.ID, result.User.ID)

	mockUserService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	authSrv, _ := createTestAuthService()
	ctx := context.Background()

	// Test with invalid refresh token
	result, err := authSrv.RefreshToken(ctx, "invalid.token.here")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	authSrv, mockUserService := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Generate a refresh token
	refreshToken, err := authSrv.generateRefreshToken(user)
	assert.NoError(t, err)

	mockUserService.On("GetUserByID", ctx, uint(1)).Return(nil, errors.New("user not found"))

	// Test
	result, err := authSrv.RefreshToken(ctx, refreshToken)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	mockUserService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InactiveUser(t *testing.T) {
	authSrv, mockUserService := createTestAuthService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: false, // User is inactive
	}

	// Generate a refresh token
	refreshToken, err := authSrv.generateRefreshToken(user)
	assert.NoError(t, err)

	mockUserService.On("GetUserByID", ctx, uint(1)).Return(user, nil)

	// Test
	result, err := authSrv.RefreshToken(ctx, refreshToken)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "account is disabled")

	mockUserService.AssertExpectations(t)
}

func TestAuthService_Logout(t *testing.T) {
	authSrv, _ := createTestAuthService()
	ctx := context.Background()

	// Test
	err := authSrv.Logout(ctx, 1)

	// Assertions
	assert.NoError(t, err)
}

func TestAuthService_GenerateAccessToken(t *testing.T) {
	authSrv, _ := createTestAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Test
	token, err := authSrv.generateAccessToken(user)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(authSrv.jwtConfig.Secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestAuthService_GenerateRefreshToken(t *testing.T) {
	authSrv, _ := createTestAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Test
	token, err := authSrv.generateRefreshToken(user)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(authSrv.jwtConfig.Secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestNewAuthService(t *testing.T) {
	mockUserService := &MockUserService{}
	config := configs.JWT{
		Secret:     "test-secret",
		Expiration: time.Hour,
	}
	logger := zap.NewNop()

	service := NewAuthService(mockUserService, config, logger)

	assert.NotNil(t, service)
	assert.IsType(t, &authService{}, service)
}
