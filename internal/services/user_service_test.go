package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func createTestUserService() (*userService, *MockUserRepository) {
	mockRepo := &MockUserRepository{}
	logger := zap.NewNop()
	userSrv := &userService{
		userRepo: mockRepo,
		logger:   logger,
	}

	return userSrv, mockRepo
}

func TestNewUserService(t *testing.T) {
	mockRepo := &MockUserRepository{}
	logger := zap.NewNop()

	service := NewUserService(mockRepo, logger)

	assert.NotNil(t, service)
	assert.IsType(t, &userService{}, service)
}

func TestUserService_CreateUser_Success(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleUser,
	}

	mockRepo.On("GetByUsername", ctx, "testuser").Return(nil, errors.New("user not found"))
	mockRepo.On("Create", ctx, user).Return(nil)

	// Test
	err := userSrv.CreateUser(ctx, user)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_MissingUsername(t *testing.T) {
	userSrv, _ := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Test
	err := userSrv.CreateUser(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username and password are required")
}

func TestUserService_CreateUser_MissingPassword(t *testing.T) {
	userSrv, _ := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test
	err := userSrv.CreateUser(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username and password are required")
}

func TestUserService_CreateUser_MissingEmail(t *testing.T) {
	userSrv, _ := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		Username: "testuser",
		Password: "password123",
	}

	// Test
	err := userSrv.CreateUser(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestUserService_CreateUser_UserAlreadyExists(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	existingUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByUsername", ctx, "testuser").Return(existingUser, nil)

	// Test
	err := userSrv.CreateUser(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_RepositoryError(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", ctx, "testuser").Return(nil, errors.New("user not found"))
	mockRepo.On("Create", ctx, user).Return(errors.New("database error"))

	// Test
	err := userSrv.CreateUser(ctx, user)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	expectedUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(expectedUser, nil)

	// Test
	user, err := userSrv.GetUserByID(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("user not found"))

	// Test
	user, err := userSrv.GetUserByID(ctx, 1)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user by ID")
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByUsername_Success(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	expectedUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByUsername", ctx, "testuser").Return(expectedUser, nil)

	// Test
	user, err := userSrv.GetUserByUsername(ctx, "testuser")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByUsername_NotFound(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	mockRepo.On("GetByUsername", ctx, "testuser").Return(nil, errors.New("user not found"))

	// Test
	user, err := userSrv.GetUserByUsername(ctx, "testuser")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user by username")
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByEmail_Success(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	expectedUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(expectedUser, nil)

	// Test
	user, err := userSrv.GetUserByEmail(ctx, "test@example.com")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByEmail_NotFound(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, errors.New("user not found"))

	// Test
	user, err := userSrv.GetUserByEmail(ctx, "test@example.com")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user by email")
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success_RoleAdminUpdate(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{
		"role": "admin",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, models.RoleAdmin, updatedUser.Role)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success_RoleUserUpdate(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{
		"role": "user",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, models.RoleUser, updatedUser.Role)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success_PasswordUpdate(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{
		"password": "newpassword123",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.NotEmpty(t, updatedUser.Password)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_InvalidPassword(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	updates := map[string]interface{}{
		"password": []int{123}, // Invalid type
	}
	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "invalid password value: [123]")
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_InvalidRole(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{
		"role": "invalid_role",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "invalid role value: invalid_role")
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_InvalidType(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{
		"role": 123,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "invalid type: 123")
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_NoUpdates(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, user, updatedUser)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_UserNotFound(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	updates := map[string]interface{}{
		"role": "admin",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("user not found"))

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "user not found")
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_RepositoryError(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	updates := map[string]interface{}{
		"role": "admin",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("database error"))

	// Test
	updatedUser, err := userSrv.UpdateUser(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "failed to update user")
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	// Test
	err := userSrv.DeleteUser(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_UserNotFound(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("user not found"))

	// Test
	err := userSrv.DeleteUser(ctx, 1)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_RepositoryError(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(errors.New("database error"))

	// Test
	err := userSrv.DeleteUser(ctx, 1)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")
	mockRepo.AssertExpectations(t)
}

func TestUserService_ListUsers_Success(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	expectedUsers := []*models.User{
		{ID: 1, Username: "user1", Email: "user1@example.com"},
		{ID: 2, Username: "user2", Email: "user2@example.com"},
	}

	mockRepo.On("List", ctx, 10, 0).Return(expectedUsers, nil)

	// Test
	users, err := userSrv.ListUsers(ctx, 10, 0)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ListUsers_RepositoryError(t *testing.T) {
	userSrv, mockRepo := createTestUserService()
	ctx := context.Background()

	mockRepo.On("List", ctx, 10, 0).Return(nil, errors.New("database error"))

	// Test
	users, err := userSrv.ListUsers(ctx, 10, 0)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Contains(t, err.Error(), "failed to list users")
	mockRepo.AssertExpectations(t)
}
