package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

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

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func createTestUserHandler() (*UserHandler, *MockUserService) {
	mockService := new(MockUserService)
	logger := zap.NewNop()
	userHandler := NewUserHandler(mockService, logger)
	return userHandler, mockService
}

func setUserIDInContext(c *gin.Context, userID uint) {
	c.Set("user_id", userID)
}

// Tests for GetProfile
func TestUserHandler_GetProfile_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	user := &models.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)

	c, w := setupGinTestContext("GET", "/api/v1/users/profile", nil)
	setUserIDInContext(c, 1)

	userHandler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_NoUserID(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContext("GET", "/api/v1/users/profile", nil)
	// Don't set user_id in context

	userHandler.GetProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_UserNotFound(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(nil, errors.New("user not found"))

	c, w := setupGinTestContext("GET", "/api/v1/users/profile", nil)
	setUserIDInContext(c, 1)

	userHandler.GetProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for UpdateProfile
func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	updates := map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "Name",
		"email":      "updated@example.com",
	}

	updatedUser := &models.User{
		ID:        1,
		Username:  "testuser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "Name",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	expectedUpdates := map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "Name",
		"email":      "updated@example.com",
	}

	mockService.On("UpdateUser", mock.Anything, uint(1), expectedUpdates).Return(updatedUser, nil)

	c, w := setupGinTestContext("PUT", "/api/v1/users/profile", updates)
	setUserIDInContext(c, 1)

	userHandler.UpdateProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_NoUserID(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	updates := map[string]interface{}{
		"first_name": "Updated",
	}

	c, w := setupGinTestContext("PUT", "/api/v1/users/profile", updates)
	// Don't set user_id in context

	userHandler.UpdateProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContext("PUT", "/api/v1/users/profile", "invalid json")
	setUserIDInContext(c, 1)

	userHandler.UpdateProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_ServiceError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	updates := map[string]interface{}{
		"email": "invalid-email",
	}

	mockService.On("UpdateUser", mock.Anything, uint(1), updates).Return(nil, errors.New("invalid email format"))

	c, w := setupGinTestContext("PUT", "/api/v1/users/profile", updates)
	setUserIDInContext(c, 1)

	userHandler.UpdateProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_FiltersSensitiveFields(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	updates := map[string]interface{}{
		"first_name": "Updated",
		"id":         999,      // Should be filtered out
		"username":   "hacker", // Should be filtered out
		"role":       "admin",  // Should be filtered out
		"is_active":  false,    // Should be filtered out
	}

	updatedUser := &models.User{
		ID:        1,
		Username:  "testuser",
		FirstName: "Updated",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	// Should only contain first_name after filtering
	expectedUpdates := map[string]interface{}{
		"first_name": "Updated",
	}

	mockService.On("UpdateUser", mock.Anything, uint(1), expectedUpdates).Return(updatedUser, nil)

	c, w := setupGinTestContext("PUT", "/api/v1/users/profile", updates)
	setUserIDInContext(c, 1)

	userHandler.UpdateProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for ChangePassword
func TestUserHandler_ChangePassword_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	// Create a user with a proper bcrypt hash
	user := &models.User{
		ID:       1,
		Username: "testuser",
	}
	// Set a real bcrypt hash for "oldpassword"
	user.SetPassword("oldpassword")

	passwordReq := map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword123",
	}

	// Mock getting the user
	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)

	// Mock updating the password
	expectedUpdates := map[string]interface{}{
		"password": "newpassword123",
	}
	mockService.On("UpdateUser", mock.Anything, uint(1), expectedUpdates).Return(user, nil)

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", passwordReq)
	setUserIDInContext(c, 1)

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_WrongCurrentPassword(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	// Create a user with a proper bcrypt hash
	user := &models.User{
		ID:       1,
		Username: "testuser",
	}
	// Set a real bcrypt hash for "oldpassword"
	user.SetPassword("oldpassword")

	passwordReq := map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword123",
	}

	// Mock getting the user
	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", passwordReq)
	setUserIDInContext(c, 1)

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_UpdateServiceError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	// Create a user with a proper bcrypt hash
	user := &models.User{
		ID:       1,
		Username: "testuser",
	}
	// Set a real bcrypt hash for "oldpassword"
	user.SetPassword("oldpassword")

	passwordReq := map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword123",
	}

	// Mock getting the user
	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)

	// Mock updating the password with error
	expectedUpdates := map[string]interface{}{
		"password": "newpassword123",
	}
	mockService.On("UpdateUser", mock.Anything, uint(1), expectedUpdates).Return(nil, errors.New("database error"))

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", passwordReq)
	setUserIDInContext(c, 1)

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_NoUserID(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	passwordReq := map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword123",
	}

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", passwordReq)
	// Don't set user_id in context

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_InvalidJSON(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", "invalid json")
	setUserIDInContext(c, 1)

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_MissingFields(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	passwordReq := map[string]string{
		"current_password": "oldpassword",
		// Missing new_password
	}

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", passwordReq)
	setUserIDInContext(c, 1)

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_UserNotFound(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	passwordReq := map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword123",
	}

	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(nil, errors.New("user not found"))

	c, w := setupGinTestContext("POST", "/api/v1/users/change-password", passwordReq)
	setUserIDInContext(c, 1)

	userHandler.ChangePassword(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for ListUsers
func TestUserHandler_ListUsers_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	users := []*models.User{
		{ID: 1, Username: "user1", Email: "user1@example.com"},
		{ID: 2, Username: "user2", Email: "user2@example.com"},
	}

	mockService.On("ListUsers", mock.Anything, 10, 0).Return(users, nil)

	c, w := setupGinTestContext("GET", "/api/v1/users", nil)
	userHandler.ListUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ListUsers_WithPagination(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	users := []*models.User{
		{ID: 3, Username: "user3", Email: "user3@example.com"},
	}

	mockService.On("ListUsers", mock.Anything, 5, 10).Return(users, nil)

	c, w := setupGinTestContext("GET", "/api/v1/users?limit=5&offset=10", nil)
	userHandler.ListUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ListUsers_ServiceError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	mockService.On("ListUsers", mock.Anything, 10, 0).Return(nil, errors.New("database error"))

	c, w := setupGinTestContext("GET", "/api/v1/users", nil)
	userHandler.ListUsers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for CreateUser
func TestUserHandler_CreateUser_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	registerReq := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	createdUser := &models.User{
		ID:        3,
		Username:  "newuser",
		Email:     "newuser@example.com",
		FirstName: "New",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Username == "newuser" && user.Email == "newuser@example.com"
	})).Return(nil)

	mockService.On("GetUserByUsername", mock.Anything, "newuser").Return(createdUser, nil)

	c, w := setupGinTestContext("POST", "/api/v1/users", registerReq)
	userHandler.CreateUser(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContext("POST", "/api/v1/users", "invalid json")
	userHandler.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_ServiceError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	registerReq := RegisterRequest{
		Username:  "existinguser",
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Existing",
		LastName:  "User",
	}

	mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Username == "existinguser"
	})).Return(errors.New("username already exists"))

	c, w := setupGinTestContext("POST", "/api/v1/users", registerReq)
	userHandler.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_GetUserError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	registerReq := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Username == "newuser"
	})).Return(nil)

	mockService.On("GetUserByUsername", mock.Anything, "newuser").Return(nil, errors.New("failed to retrieve user"))

	c, w := setupGinTestContext("POST", "/api/v1/users", registerReq)
	userHandler.CreateUser(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for UpdateUser
func TestUserHandler_UpdateUser_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	updates := map[string]interface{}{
		"first_name": "Updated",
		"role":       "admin",
	}

	updatedUser := &models.User{
		ID:        1,
		Username:  "testuser",
		FirstName: "Updated",
		Role:      models.RoleAdmin,
	}

	mockService.On("UpdateUser", mock.Anything, uint(1), updates).Return(updatedUser, nil)

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/users/1", updates, map[string]string{"id": "1"})
	userHandler.UpdateUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_InvalidID(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/users/invalid", nil, map[string]string{"id": "invalid"})
	userHandler.UpdateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_InvalidJSON(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/users/1", "invalid json", map[string]string{"id": "1"})
	userHandler.UpdateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_ServiceError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	updates := map[string]interface{}{
		"username": "duplicateuser",
	}

	mockService.On("UpdateUser", mock.Anything, uint(1), updates).Return(nil, errors.New("username already exists"))

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/users/1", updates, map[string]string{"id": "1"})
	userHandler.UpdateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for DeleteUser
func TestUserHandler_DeleteUser_Success(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	mockService.On("DeleteUser", mock.Anything, uint(1)).Return(nil)

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/users/1", nil, map[string]string{"id": "1"})
	userHandler.DeleteUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_InvalidID(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/users/invalid", nil, map[string]string{"id": "invalid"})
	userHandler.DeleteUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_ServiceError(t *testing.T) {
	userHandler, mockService := createTestUserHandler()

	mockService.On("DeleteUser", mock.Anything, uint(1)).Return(errors.New("cannot delete user"))

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/users/1", nil, map[string]string{"id": "1"})
	userHandler.DeleteUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
