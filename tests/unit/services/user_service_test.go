package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type UserServiceTestSuite struct {
	suite.Suite
	userService services.IUserService
	mockRepo    MockUserRepository
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockRepo = MockUserRepository{}
	suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (suite *UserServiceTestSuite) TestCreateUser() {
	testCases := []struct {
		name           string
		request        dto.CreateUserRequest
		setupMocks     func()
		expectedError  string
		validateResult func(*models.User)
	}{
		{
			name: "Success",
			request: dto.CreateUserRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Scopes:    []models.APIScope{"server:read", "server:write"},
			},
			setupMocks: func() {
				suite.mockRepo.On("ExistsByUserNameOrEmail", mock.Anything, "testuser", "test@example.com").Return(false, nil)
				suite.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: "",
			validateResult: func(user *models.User) {
				suite.NotNil(user)
				suite.Equal("testuser", user.Username)
				suite.Equal("test@example.com", user.Email)
				suite.Equal("Test", user.FirstName)
				suite.Equal("User", user.LastName)
			},
		},
		{
			name: "UserExists",
			request: dto.CreateUserRequest{
				Username:  "existinguser",
				Email:     "existing@example.com",
				FirstName: "Existing",
				LastName:  "User",
				Password:  "password123",
				Scopes:    []models.APIScope{"server:read"},
			},
			setupMocks: func() {
				suite.mockRepo.On("ExistsByUserNameOrEmail", mock.Anything, "existinguser", "existing@example.com").Return(true, nil)
			},
			expectedError: "user with username or email already exists",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
		{
			name: "DatabaseError",
			request: dto.CreateUserRequest{
				Username:  "testuser2",
				Email:     "test2@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Scopes:    []models.APIScope{"server:read"},
			},
			setupMocks: func() {
				suite.mockRepo.On("ExistsByUserNameOrEmail", mock.Anything, "testuser2", "test2@example.com").Return(false, errors.New("database error"))
			},
			expectedError: "failed to check if user exists: database error",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
		{
			name: "CreateError",
			request: dto.CreateUserRequest{
				Username:  "testuser3",
				Email:     "test3@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Scopes:    []models.APIScope{"server:read"},
			},
			setupMocks: func() {
				suite.mockRepo.On("ExistsByUserNameOrEmail", mock.Anything, "testuser3", "test3@example.com").Return(false, nil)
				suite.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("create error"))
			},
			expectedError: "failed to create user: create error",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}
			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			user, err := suite.userService.CreateUser(context.Background(), tc.request)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}

			tc.validateResult(user)
		})
	}
}

func (suite *UserServiceTestSuite) TestGetUserByID() {
	testCases := []struct {
		name           string
		userID         uint
		setupMocks     func()
		expectedError  string
		validateResult func(*models.User)
	}{
		{
			name:   "Success",
			userID: 1,
			setupMocks: func() {
				expectedUser := &models.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(expectedUser, nil)
			},
			expectedError: "",
			validateResult: func(user *models.User) {
				suite.NotNil(user)
				suite.Equal(uint(1), user.ID)
				suite.Equal("testuser", user.Username)
			},
		},
		{
			name:   "UserNotFound",
			userID: 999,
			setupMocks: func() {
				suite.mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}
			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			user, err := suite.userService.GetUserByID(context.Background(), tc.userID)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}

			tc.validateResult(user)
		})
	}
}

func (suite *UserServiceTestSuite) TestGetUserByUsername() {
	testCases := []struct {
		name           string
		username       string
		setupMocks     func()
		expectedError  string
		validateResult func(*models.User)
	}{
		{
			name:     "Success",
			username: "testuser",
			setupMocks: func() {
				expectedUser := &models.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				suite.mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(expectedUser, nil)
			},
			expectedError: "",
			validateResult: func(user *models.User) {
				suite.NotNil(user)
				suite.Equal("testuser", user.Username)
			},
		},
		{
			name:     "UserNotFound",
			username: "nonexistent",
			setupMocks: func() {
				suite.mockRepo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}
			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			user, err := suite.userService.GetUserByUsername(context.Background(), tc.username)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}

			tc.validateResult(user)
		})
	}
}

func (suite *UserServiceTestSuite) TestDeleteUser() {
	testCases := []struct {
		name          string
		userID        uint
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "Success",
			userID: 1,
			setupMocks: func() {
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(user, nil)
				suite.mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "UserNotFound",
			userID: 999,
			setupMocks: func() {
				suite.mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:   "DeleteError",
			userID: 1,
			setupMocks: func() {
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(user, nil)
				suite.mockRepo.On("Delete", mock.Anything, uint(1)).Return(errors.New("delete error"))
			},
			expectedError: "delete error",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}
			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			err := suite.userService.DeleteUser(context.Background(), tc.userID)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *UserServiceTestSuite) TestListUsers() {
	testCases := []struct {
		name           string
		limit          int
		offset         int
		setupMocks     func()
		expectedError  string
		validateResult func([]models.User)
	}{
		{
			name:   "Success",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				expectedUsers := []models.User{
					{ID: 1, Username: "user1", Email: "user1@example.com"},
					{ID: 2, Username: "user2", Email: "user2@example.com"},
				}
				suite.mockRepo.On("List", mock.Anything, 10, 0).Return(expectedUsers, nil)
			},
			expectedError: "",
			validateResult: func(users []models.User) {
				suite.Len(users, 2)
				suite.Equal("user1", users[0].Username)
				suite.Equal("user2", users[1].Username)
			},
		},
		{
			name:   "EmptyResult",
			limit:  10,
			offset: 100,
			setupMocks: func() {
				suite.mockRepo.On("List", mock.Anything, 10, 100).Return([]models.User{}, nil)
			},
			expectedError: "",
			validateResult: func(users []models.User) {
				suite.Len(users, 0)
			},
		},
		{
			name:   "DatabaseError",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				suite.mockRepo.On("List", mock.Anything, 10, 0).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
			validateResult: func(users []models.User) {
				suite.Nil(users)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}

			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			users, err := suite.userService.ListUsers(context.Background(), tc.limit, tc.offset)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}

			tc.validateResult(users)
		})
	}
}

func (suite *UserServiceTestSuite) TestUpdateUser() {
	testCases := []struct {
		name           string
		userID         uint
		updateRequest  dto.UserUpdate
		setupMocks     func()
		expectedError  string
		validateResult func(*models.User)
	}{
		{
			name:   "Success",
			userID: 1,
			updateRequest: dto.UserUpdate{
				Username:  "updateduser",
				Email:     "updated@example.com",
				Password:  "newpassword123",
				FirstName: "Updated",
				LastName:  "User",
				IsActive:  true,
				Scopes:    []models.APIScope{"server:read"},
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:        1,
					Username:  "olduser",
					Email:     "old@example.com",
					FirstName: "Old",
					LastName:  "User",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
				suite.mockRepo.On("ExistsByUserNameOrEmail", mock.Anything, "updateduser", "updated@example.com").Return(false, nil)
				suite.mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: "",
			validateResult: func(user *models.User) {
				suite.NotNil(user)
				suite.Equal("updateduser", user.Username)
				suite.Equal("updated@example.com", user.Email)
			},
		},
		{
			name:   "UserNotFound",
			userID: 999,
			updateRequest: dto.UserUpdate{
				Username: "test",
				Email:    "test@example.com",
			},
			setupMocks: func() {
				suite.mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
		{
			name:   "UserExists",
			userID: 1,
			updateRequest: dto.UserUpdate{
				Username: "existinguser",
				Email:    "existing@example.com",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "olduser",
					Email:    "old@example.com",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
				suite.mockRepo.On("ExistsByUserNameOrEmail", mock.Anything, "existinguser", "existing@example.com").Return(true, nil)
			},
			expectedError: "user with username or email already exists",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}

			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			user, err := suite.userService.UpdateUser(context.Background(), tc.userID, tc.updateRequest)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}

			tc.validateResult(user)
		})
	}
}

func (suite *UserServiceTestSuite) TestUpdateProfile() {
	testCases := []struct {
		name           string
		userID         uint
		updateRequest  dto.ProfileUpdate
		setupMocks     func()
		expectedError  string
		validateResult func(*models.User)
	}{
		{
			name:   "Success",
			userID: 1,
			updateRequest: dto.ProfileUpdate{
				FirstName: "UpdatedFirst",
				LastName:  "UpdatedLast",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					FirstName: "Original",
					LastName:  "Name",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
				suite.mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: "",
			validateResult: func(user *models.User) {
				suite.NotNil(user)
				suite.Equal("UpdatedFirst", user.FirstName)
				suite.Equal("UpdatedLast", user.LastName)
			},
		},
		{
			name:   "UserNotFound",
			userID: 999,
			updateRequest: dto.ProfileUpdate{
				FirstName: "Test",
				LastName:  "User",
			},
			setupMocks: func() {
				suite.mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
		{
			name:   "UpdateError",
			userID: 1,
			updateRequest: dto.ProfileUpdate{
				FirstName: "Test",
				LastName:  "User",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
				suite.mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("update error"))
			},
			expectedError: "failed to update user profile: update error",
			validateResult: func(user *models.User) {
				suite.Nil(user)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}

			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			user, err := suite.userService.UpdateProfile(context.Background(), tc.userID, tc.updateRequest)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}

			tc.validateResult(user)
		})
	}
}

func (suite *UserServiceTestSuite) TestUpdatePassword() {
	testCases := []struct {
		name          string
		userID        uint
		updateRequest dto.PasswordUpdate
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "Success",
			userID: 1,
			updateRequest: dto.PasswordUpdate{
				OldPassword:    "oldpassword",
				NewPassword:    "newpassword",
				RepeatPassword: "newpassword",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				// Set old password
				existingUser.SetPassword("oldpassword")

				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
				suite.mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "UserNotFound",
			userID: 999,
			updateRequest: dto.PasswordUpdate{
				OldPassword: "oldpassword",
				NewPassword: "newpassword",
			},
			setupMocks: func() {
				suite.mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:   "WrongOldPassword",
			userID: 1,
			updateRequest: dto.PasswordUpdate{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				existingUser.SetPassword("correctpassword")
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
			},
			expectedError: "old password is incorrect",
		},
		{
			name:   "NewPasswordsDoNotMatch",
			userID: 1,
			updateRequest: dto.PasswordUpdate{
				OldPassword:    "oldpassword",
				NewPassword:    "newpassword",
				RepeatPassword: "differentpassword",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				existingUser.SetPassword("oldpassword")
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
			},
			expectedError: "new password and repeat password do not match",
		},
		{
			name:   "UpdateError",
			userID: 1,
			updateRequest: dto.PasswordUpdate{
				OldPassword:    "oldpassword",
				NewPassword:    "newpassword",
				RepeatPassword: "newpassword",
			},
			setupMocks: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				existingUser.SetPassword("oldpassword")
				suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingUser, nil)
				suite.mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("update error"))
			},
			expectedError: "failed to update user password: update error",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mocks
			suite.mockRepo = MockUserRepository{}

			suite.userService = services.NewUserService(&suite.mockRepo, zap.NewNop())

			// Setup mocks
			tc.setupMocks()

			// Execute
			err := suite.userService.UpdatePassword(context.Background(), tc.userID, tc.updateRequest)

			// Validate
			if tc.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.expectedError)
			} else {
				suite.NoError(err)
			}
		})
	}
}
