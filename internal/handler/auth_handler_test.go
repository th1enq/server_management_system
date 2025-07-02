package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*services.AuthResponse, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, user *models.User) (*services.AuthResponse, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*services.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*services.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Claims), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) LogoutWithToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func createTestAuthHandler() (*AuthHandler, *MockAuthService) {
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	authHandler := NewAuthHandler(mockService, logger)

	return authHandler, mockService
}

// Test Login Success
func TestAuthHandler_Login_Success(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	loginRequest := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}

	expectedResponse := &services.AuthResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: &models.User{
			ID:        1,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Role:      models.RoleUser,
			IsActive:  true,
		},
		Scopes: []string{"server:read", "report:write"},
	}

	mockAuthService.On("Login", mock.Anything, "testuser", "testpass").Return(expectedResponse, nil)

	c, w := setupGinTestContext("POST", "/api/v1/auth/login", loginRequest)
	authHandler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)

	var response services.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.AccessToken, response.AccessToken)
	assert.Equal(t, expectedResponse.User.Username, response.User.Username)
}

// Test Login Invalid JSON
func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	authHandler, _ := createTestAuthHandler()

	c, w := setupGinTestContext("POST", "/api/v1/auth/login", nil)
	c.Request.Body = http.NoBody

	authHandler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "EOF")
}

// Test Login Missing Fields
func TestAuthHandler_Login_MissingFields(t *testing.T) {
	authHandler, _ := createTestAuthHandler()

	loginRequest := LoginRequest{
		Username: "testuser",
		// Password is missing
	}

	c, w := setupGinTestContext("POST", "/api/v1/auth/login", loginRequest)
	authHandler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "required")
}

// Test Login Service Error
func TestAuthHandler_Login_ServiceError(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	loginRequest := LoginRequest{
		Username: "testuser",
		Password: "wrongpass",
	}

	mockAuthService.On("Login", mock.Anything, "testuser", "wrongpass").Return(nil, errors.New("invalid credentials"))

	c, w := setupGinTestContext("POST", "/api/v1/auth/login", loginRequest)
	authHandler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid credentials", response["error"])
}

// Test Register Success
func TestAuthHandler_Register_Success(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	registerRequest := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	expectedUser := &models.User{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	expectedResponse := &services.AuthResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: &models.User{
			ID:        2,
			Username:  "newuser",
			Email:     "newuser@example.com",
			FirstName: "New",
			LastName:  "User",
			Role:      models.RoleUser,
			IsActive:  true,
		},
	}

	mockAuthService.On("Register", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Username == expectedUser.Username &&
			user.Email == expectedUser.Email &&
			user.Password == expectedUser.Password &&
			user.FirstName == expectedUser.FirstName &&
			user.LastName == expectedUser.LastName &&
			user.Role == expectedUser.Role &&
			user.IsActive == expectedUser.IsActive
	})).Return(expectedResponse, nil)

	c, w := setupGinTestContext("POST", "/api/v1/auth/register", registerRequest)
	authHandler.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockAuthService.AssertExpectations(t)

	var response services.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.AccessToken, response.AccessToken)
	assert.Equal(t, expectedResponse.User.Username, response.User.Username)
}

// Test Register Invalid Email
func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	authHandler, _ := createTestAuthHandler()

	registerRequest := RegisterRequest{
		Username:  "newuser",
		Email:     "invalid-email", // Invalid email format
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	c, w := setupGinTestContext("POST", "/api/v1/auth/register", registerRequest)
	authHandler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "email")
}

// Test Register Short Password
func TestAuthHandler_Register_ShortPassword(t *testing.T) {
	authHandler, _ := createTestAuthHandler()

	registerRequest := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "123", // Too short
		FirstName: "New",
		LastName:  "User",
	}

	c, w := setupGinTestContext("POST", "/api/v1/auth/register", registerRequest)
	authHandler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "min")
}

// Test Register Service Error (User Already Exists)
func TestAuthHandler_Register_UserExists(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	registerRequest := RegisterRequest{
		Username:  "existinguser",
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Existing",
		LastName:  "User",
	}

	mockAuthService.On("Register", mock.Anything, mock.Anything).Return(nil, errors.New("user already exists"))

	c, w := setupGinTestContext("POST", "/api/v1/auth/register", registerRequest)
	authHandler.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockAuthService.AssertExpectations(t)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user already exists", response["error"])
}

// Test RefreshToken Success
func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	refreshRequest := RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedResponse := &services.AuthResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	mockAuthService.On("RefreshToken", mock.Anything, "valid_refresh_token").Return(expectedResponse, nil)

	c, w := setupGinTestContext("POST", "/api/v1/auth/refresh", refreshRequest)
	authHandler.RefreshToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)

	var response services.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.AccessToken, response.AccessToken)
}

// Test RefreshToken Invalid Token
func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	refreshRequest := RefreshTokenRequest{
		RefreshToken: "invalid_refresh_token",
	}

	mockAuthService.On("RefreshToken", mock.Anything, "invalid_refresh_token").Return(nil, errors.New("invalid refresh token"))

	c, w := setupGinTestContext("POST", "/api/v1/auth/refresh", refreshRequest)
	authHandler.RefreshToken(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid refresh token", response["error"])
}

// Test RefreshToken Missing Token
func TestAuthHandler_RefreshToken_MissingToken(t *testing.T) {
	authHandler, _ := createTestAuthHandler()

	refreshRequest := RefreshTokenRequest{
		// RefreshToken is missing
	}

	c, w := setupGinTestContext("POST", "/api/v1/auth/refresh", refreshRequest)
	authHandler.RefreshToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "required")
}

// Test Logout Success
func TestAuthHandler_Logout_Success(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	mockAuthService.On("Logout", mock.Anything, uint(1)).Return(nil)

	c, w := setupGinTestContext("POST", "/api/v1/auth/logout", nil)
	// Simulate middleware setting user ID
	c.Set("user_id", uint(1))

	authHandler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}

// Test Logout No User ID
func TestAuthHandler_Logout_NoUserID(t *testing.T) {
	authHandler, _ := createTestAuthHandler()

	c, w := setupGinTestContext("POST", "/api/v1/auth/logout", nil)
	// Don't set user_id to simulate missing authentication

	authHandler.Logout(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Authentication required", response["error"])
}

// Test Logout Service Error
func TestAuthHandler_Logout_ServiceError(t *testing.T) {
	authHandler, mockAuthService := createTestAuthHandler()

	mockAuthService.On("Logout", mock.Anything, uint(1)).Return(errors.New("failed to invalidate session"))

	c, w := setupGinTestContext("POST", "/api/v1/auth/logout", nil)
	// Simulate middleware setting user ID
	c.Set("user_id", uint(1))

	authHandler.Logout(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuthService.AssertExpectations(t)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Failed to logout", response["error"])
}
