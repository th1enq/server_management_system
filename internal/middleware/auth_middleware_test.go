package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func createTestAuthMiddleware() (*AuthMiddleware, *MockAuthService) {
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	middleware := NewAuthMiddleware(mockService, logger)
	return middleware, mockService
}

func setupGinTestContext(method, url string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, url, nil)
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	c.Request = req
	return c, w
}

// Test RequireAuth Success
func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	middleware, mockService := createTestAuthMiddleware()

	claims := &services.Claims{
		UserID:    1,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      models.RoleUser,
		TokenType: "access",
		Scopes:    []models.APIScope{models.ScopeServerRead},
	}

	mockService.On("ValidateToken", "valid_token").Return(claims, nil)

	// Set up a flag to check if next was called
	nextCalled := false
	originalHandler := func(c *gin.Context) {
		nextCalled = true
	}

	// Create a handler chain
	r := gin.New()
	r.Use(middleware.RequireAuth())
	r.GET("/test", originalHandler)

	// Perform the request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	mockService.AssertExpectations(t)
}

// Test RequireAuth Missing Authorization Header
func TestAuthMiddleware_RequireAuth_MissingHeader(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/test", nil)

	handler := middleware.RequireAuth()
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
	assert.Equal(t, "Authentication required", response.Message)
}

// Test RequireAuth Invalid Token Format
func TestAuthMiddleware_RequireAuth_InvalidTokenFormat(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/test", map[string]string{
		"Authorization": "InvalidFormat token",
	})

	handler := middleware.RequireAuth()
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
}

// Test RequireAuth Invalid Token
func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	middleware, mockService := createTestAuthMiddleware()

	mockService.On("ValidateToken", "invalid_token").Return(nil, errors.New("token expired"))

	c, w := setupGinTestContext("GET", "/test", map[string]string{
		"Authorization": "Bearer invalid_token",
	})

	handler := middleware.RequireAuth()
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	mockService.AssertExpectations(t)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
	assert.Equal(t, "Invalid token", response.Message)
}

// Test RequireAuth Wrong Token Type
func TestAuthMiddleware_RequireAuth_WrongTokenType(t *testing.T) {
	middleware, mockService := createTestAuthMiddleware()

	claims := &services.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh", // Wrong type
		Scopes:    []models.APIScope{},
	}

	mockService.On("ValidateToken", "refresh_token").Return(claims, nil)

	c, w := setupGinTestContext("GET", "/test", map[string]string{
		"Authorization": "Bearer refresh_token",
	})

	handler := middleware.RequireAuth()
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	mockService.AssertExpectations(t)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
	assert.Equal(t, "Invalid token type", response.Message)
}

// Test RequireRole Success
func TestAuthMiddleware_RequireRole_Success(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	// Create a handler chain
	nextCalled := false
	originalHandler := func(c *gin.Context) {
		nextCalled = true
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	r.Use(middleware.RequireRole("admin"))
	r.GET("/admin", originalHandler)

	// Perform the request
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test RequireRole No Authentication
func TestAuthMiddleware_RequireRole_NoAuth(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/admin", nil)
	// Don't set role to simulate missing authentication

	handler := middleware.RequireRole("admin")
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
	assert.Equal(t, "Authentication required", response.Message)
}

// Test RequireRole Insufficient Permission
func TestAuthMiddleware_RequireRole_InsufficientPermission(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/admin", nil)
	c.Set("role", "user") // User trying to access admin endpoint

	handler := middleware.RequireRole("admin")
	handler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "FORBIDDEN", response.Code)
	assert.Equal(t, "Insufficient permissions", response.Message)
}

// Test RequireRole Invalid Role Data
func TestAuthMiddleware_RequireRole_InvalidRoleData(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/admin", nil)
	c.Set("role", 123) // Invalid type - should be string

	handler := middleware.RequireRole("admin")
	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Invalid role data", response.Message)
}

// Test RequireAdmin Success
func TestAuthMiddleware_RequireAdmin_Success(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	// Create a handler chain
	nextCalled := false
	originalHandler := func(c *gin.Context) {
		nextCalled = true
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	r.Use(middleware.RequireAdmin())
	r.GET("/admin", originalHandler)

	// Perform the request
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test RequireScope Success
func TestAuthMiddleware_RequireScope_Success(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	// Create a handler chain
	nextCalled := false
	originalHandler := func(c *gin.Context) {
		nextCalled = true
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("scopes", []models.APIScope{models.ScopeServerRead, models.ScopeServerWrite})
		c.Next()
	})
	r.Use(middleware.RequireScope("server:read"))
	r.GET("/servers", originalHandler)

	// Perform the request
	req := httptest.NewRequest("GET", "/servers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test RequireScope No Authentication
func TestAuthMiddleware_RequireScope_NoAuth(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/servers", nil)
	// Don't set scopes to simulate missing authentication

	handler := middleware.RequireScope("server:read")
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
	assert.Equal(t, "Authentication required", response.Message)
}

// Test RequireScope Insufficient Permission
func TestAuthMiddleware_RequireScope_InsufficientPermission(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/servers", nil)
	c.Set("scopes", []models.APIScope{models.ScopeReportRead}) // Different scope

	handler := middleware.RequireScope("server:read")
	handler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "FORBIDDEN", response.Code)
	assert.Equal(t, "Insufficient scope permissions", response.Message)
}

// Test RequireScope Invalid Scope Data
func TestAuthMiddleware_RequireScope_InvalidScopeData(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/servers", nil)
	c.Set("scopes", "invalid") // Invalid type - should be []models.APIScope

	handler := middleware.RequireScope("server:read")
	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Invalid scope data", response.Message)
}

// Test RequireAnyScope Success
func TestAuthMiddleware_RequireAnyScope_Success(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	// Create a handler chain
	nextCalled := false
	originalHandler := func(c *gin.Context) {
		nextCalled = true
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("scopes", []models.APIScope{models.ScopeServerRead})
		c.Next()
	})
	r.Use(middleware.RequireAnyScope("server:read", "server:write"))
	r.GET("/servers", originalHandler)

	// Perform the request
	req := httptest.NewRequest("GET", "/servers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test RequireAnyScope No Authentication
func TestAuthMiddleware_RequireAnyScope_NoAuth(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/servers", nil)
	// Don't set scopes to simulate missing authentication

	handler := middleware.RequireAnyScope("server:read", "server:write")
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Code)
	assert.Equal(t, "Authentication required", response.Message)
}

// Test RequireAnyScope Insufficient Permission
func TestAuthMiddleware_RequireAnyScope_InsufficientPermission(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/servers", nil)
	c.Set("scopes", []models.APIScope{models.ScopeReportRead}) // Different scope

	handler := middleware.RequireAnyScope("server:read", "server:write")
	handler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "FORBIDDEN", response.Code)
	assert.Equal(t, "Insufficient scope permissions", response.Message)
}

// Test RequireAnyScope Invalid Scope Data
func TestAuthMiddleware_RequireAnyScope_InvalidScopeData(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	c, w := setupGinTestContext("GET", "/servers", nil)
	c.Set("scopes", "invalid") // Invalid type - should be []models.APIScope

	handler := middleware.RequireAnyScope("server:read", "server:write")
	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, c.IsAborted())

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Invalid scope data", response.Message)
}

// Test extractTokenFromHeader
func TestAuthMiddleware_ExtractTokenFromHeader(t *testing.T) {
	middleware, _ := createTestAuthMiddleware()

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Valid Bearer token",
			header:   "Bearer valid_token_123",
			expected: "valid_token_123",
		},
		{
			name:     "Empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "Invalid format - no Bearer",
			header:   "token123",
			expected: "",
		},
		{
			name:     "Invalid format - wrong prefix",
			header:   "Basic token123",
			expected: "",
		},
		{
			name:     "Bearer with no token",
			header:   "Bearer ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupGinTestContext("GET", "/test", map[string]string{
				"Authorization": tt.header,
			})

			result := middleware.extractTokenFromHeader(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test GetUserID
func TestGetUserID(t *testing.T) {
	tests := []struct {
		name       string
		userID     interface{}
		setUserID  bool
		expectedID uint
		expectedOK bool
	}{
		{
			name:       "Valid user ID",
			userID:     uint(123),
			setUserID:  true,
			expectedID: 123,
			expectedOK: true,
		},
		{
			name:       "No user ID set",
			setUserID:  false,
			expectedID: 0,
			expectedOK: false,
		},
		{
			name:       "Invalid user ID type",
			userID:     "123",
			setUserID:  true,
			expectedID: 0,
			expectedOK: false,
		},
		{
			name:       "Zero user ID",
			userID:     uint(0),
			setUserID:  true,
			expectedID: 0,
			expectedOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupGinTestContext("GET", "/test", nil)

			if tt.setUserID {
				c.Set("user_id", tt.userID)
			}

			id, ok := GetUserID(c)
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}
