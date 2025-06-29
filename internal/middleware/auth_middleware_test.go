package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*services.AuthResponse, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, user *models.User) (*services.AuthResponse, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*services.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Claims), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*services.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Create test claims
	testClaims := &services.Claims{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", "valid-token").Return(testClaims, nil)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		username, _ := GetUsername(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
		})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["user_id"])
	assert.Equal(t, "testuser", response["username"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_NoToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request without token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization token required", response["error"])

	// Verify no service calls were made
	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Mock expectations
	mockAuthService.On("ValidateToken", "invalid-token").Return((*services.Claims)(nil), fmt.Errorf("invalid token"))

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid token", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAdmin_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Create test claims for admin user
	testClaims := &services.Claims{
		UserID:   1,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdmin,
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", "admin-token").Return(testClaims, nil)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.Use(authMiddleware.RequireAdmin())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "admin access granted", response["message"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAdmin_InsufficientPermissions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Create test claims for regular user
	testClaims := &services.Claims{
		UserID:   1,
		Username: "user",
		Email:    "user@example.com",
		Role:     models.RoleUser,
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", "user-token").Return(testClaims, nil)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.Use(authMiddleware.RequireAdmin())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}

func TestGetHelperFunctions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		// Set test data in context
		c.Set("user_id", uint(123))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("claims", &services.Claims{
			UserID:   123,
			Username: "testuser",
		})

		// Test helper functions
		userID, userIDExists := GetUserID(c)
		username, usernameExists := GetUsername(c)
		role, roleExists := GetUserRole(c)
		claims, claimsExist := GetClaims(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":         userID,
			"user_id_exists":  userIDExists,
			"username":        username,
			"username_exists": usernameExists,
			"role":            role,
			"role_exists":     roleExists,
			"claims_exist":    claimsExist,
			"claims_user_id":  claims.UserID,
		})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(123), response["user_id"])
	assert.Equal(t, true, response["user_id_exists"])
	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, true, response["username_exists"])
	assert.Equal(t, "admin", response["role"])
	assert.Equal(t, true, response["role_exists"])
	assert.Equal(t, true, response["claims_exist"])
	assert.Equal(t, float64(123), response["claims_user_id"])
}

func TestAuthMiddleware_RequireScope_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Create test claims with specific scopes
	testClaims := &services.Claims{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		Scopes:   []models.APIScope{models.ScopeServerRead, models.ScopeProfileRead},
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", "valid-token").Return(testClaims, nil)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.Use(authMiddleware.RequireScope("server:read"))
	router.GET("/servers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/servers", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "access granted", response["message"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireScope_InsufficientScope(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Create test claims without required scope
	testClaims := &services.Claims{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		Scopes:   []models.APIScope{models.ScopeProfileRead}, // Missing server:write scope
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", "valid-token").Return(testClaims, nil)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.Use(authMiddleware.RequireScope("server:write"))
	router.POST("/servers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/servers", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Insufficient scope permissions", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAnyScope_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	logger := zap.NewNop()
	authMiddleware := NewAuthMiddleware(mockAuthService, logger)

	// Create test claims with one of the required scopes
	testClaims := &services.Claims{
		UserID:   1,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdmin,
		Scopes:   []models.APIScope{models.ScopeAdminAll, models.ScopeUserRead},
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", "admin-token").Return(testClaims, nil)

	// Setup router and handler
	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.Use(authMiddleware.RequireAnyScope("admin:all", "user:write"))
	router.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "access granted", response["message"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
}
