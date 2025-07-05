package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo repository.IUserRepository
	db   db.IDatabaseClient
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(suite.T(), err)

	gormDB.AutoMigrate(&models.User{})

	suite.db = db.NewDatabaseWithGorm(gormDB)
	suite.repo = repository.NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), sqlDB.Close())
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TestCreateUser() {
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, fetchedUser.Username)
	assert.Equal(suite.T(), user.Email, fetchedUser.Email)
	assert.Equal(suite.T(), user.FirstName, fetchedUser.FirstName)
	assert.Equal(suite.T(), user.LastName, fetchedUser.LastName)
	assert.Equal(suite.T(), user.Role, fetchedUser.Role)
	assert.Equal(suite.T(), user.IsActive, fetchedUser.IsActive)
}

func (suite *UserRepositoryTestSuite) TestGetByID() {
	user := &models.User{
		Username:  "testuser1",
		Email:     "testuser1@example.com",
		FirstName: "Test",
		LastName:  "User1",
		Role:      models.RoleUser,
		IsActive:  true,
		Scopes:    123,
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, fetchedUser.Username)
	assert.Equal(suite.T(), user.Email, fetchedUser.Email)
	assert.Equal(suite.T(), user.Scopes, fetchedUser.Scopes)

	// Test not found
	_, err = suite.repo.GetByID(suite.T().Context(), 9999)
	assert.Error(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername() {
	user := &models.User{
		Username:  "uniqueuser",
		Email:     "unique@example.com",
		FirstName: "Unique",
		LastName:  "User",
		Role:      models.RoleAdmin,
		IsActive:  true,
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Test successful retrieval
	fetchedUser, err := suite.repo.GetByUsername(suite.T().Context(), "uniqueuser")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, fetchedUser.Username)
	assert.Equal(suite.T(), user.Role, fetchedUser.Role)

	// Test not found
	_, err = suite.repo.GetByUsername(suite.T().Context(), "nonexistentuser")
	assert.Error(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestUpdateUser() {
	user := &models.User{
		Username:  "updateuser",
		Email:     "update@example.com",
		FirstName: "Original",
		LastName:  "Name",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Update user
	user.FirstName = "Updated"
	user.LastName = "User"
	user.Role = models.RoleAdmin
	user.IsActive = false
	user.Scopes = 456
	err = suite.repo.Update(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Verify update
	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated", fetchedUser.FirstName)
	assert.Equal(suite.T(), "User", fetchedUser.LastName)
	assert.Equal(suite.T(), models.RoleAdmin, fetchedUser.Role)
	assert.Equal(suite.T(), false, fetchedUser.IsActive)
	assert.Equal(suite.T(), int64(456), fetchedUser.Scopes)
}

func (suite *UserRepositoryTestSuite) TestDeleteUser() {
	user := &models.User{
		Username:  "deleteuser",
		Email:     "delete@example.com",
		FirstName: "Delete",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Delete user
	err = suite.repo.Delete(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.Error(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestList() {
	// Create test users
	users := []models.User{
		{
			Username:  "listuser1",
			Email:     "listuser1@example.com",
			FirstName: "List",
			LastName:  "User1",
			Role:      models.RoleUser,
			IsActive:  true,
		},
		{
			Username:  "listuser2",
			Email:     "listuser2@example.com",
			FirstName: "List",
			LastName:  "User2",
			Role:      models.RoleAdmin,
			IsActive:  true,
		},
		{
			Username:  "listuser3",
			Email:     "listuser3@example.com",
			FirstName: "List",
			LastName:  "User3",
			Role:      models.RoleUser,
			IsActive:  false,
		},
	}

	for i := range users {
		err := users[i].SetPassword("password123")
		assert.NoError(suite.T(), err)
		err = suite.repo.Create(suite.T().Context(), &users[i])
		assert.NoError(suite.T(), err)
	}

	// Test list with limit and offset
	result, err := suite.repo.List(suite.T().Context(), 10, 0)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(result), 3)

	// Test pagination
	result, err = suite.repo.List(suite.T().Context(), 2, 0)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(result))

	result, err = suite.repo.List(suite.T().Context(), 2, 1)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(result), 1)
}

func (suite *UserRepositoryTestSuite) TestExistsByUserNameOrEmail() {
	user := &models.User{
		Username:  "existsuser",
		Email:     "exists@example.com",
		FirstName: "Exists",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Test username exists
	exists, err := suite.repo.ExistsByUserNameOrEmail(suite.T().Context(), "existsuser", "other@example.com")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test email exists
	exists, err = suite.repo.ExistsByUserNameOrEmail(suite.T().Context(), "otheruser", "exists@example.com")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test both exist
	exists, err = suite.repo.ExistsByUserNameOrEmail(suite.T().Context(), "existsuser", "exists@example.com")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test neither exists
	exists, err = suite.repo.ExistsByUserNameOrEmail(suite.T().Context(), "nonexistent", "nonexistent@example.com")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestCreateUserWithDuplicateUsername() {
	user1 := &models.User{
		Username:  "duplicateuser",
		Email:     "user1@example.com",
		FirstName: "User",
		LastName:  "One",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err := user1.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user1)
	assert.NoError(suite.T(), err)

	// Try to create another user with the same username
	user2 := &models.User{
		Username:  "duplicateuser",
		Email:     "user2@example.com",
		FirstName: "User",
		LastName:  "Two",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err = user2.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *UserRepositoryTestSuite) TestCreateUserWithDuplicateEmail() {
	user1 := &models.User{
		Username:  "user1",
		Email:     "duplicate@example.com",
		FirstName: "User",
		LastName:  "One",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err := user1.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user1)
	assert.NoError(suite.T(), err)

	// Try to create another user with the same email
	user2 := &models.User{
		Username:  "user2",
		Email:     "duplicate@example.com",
		FirstName: "User",
		LastName:  "Two",
		Role:      models.RoleUser,
		IsActive:  true,
	}
	err = user2.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *UserRepositoryTestSuite) TestUserPasswordMethods() {
	user := &models.User{
		Username:  "passworduser",
		Email:     "password@example.com",
		FirstName: "Password",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	// Test setting password
	err := user.SetPassword("securepassword123")
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), user.Password)

	// Test empty password
	err = user.SetPassword("")
	assert.Error(suite.T(), err)

	// Reset to valid password
	err = user.SetPassword("securepassword123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Test password checking
	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)

	err = fetchedUser.CheckPassword("securepassword123")
	assert.NoError(suite.T(), err)

	err = fetchedUser.CheckPassword("wrongpassword")
	assert.Error(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestUserScopes() {
	user := &models.User{
		Username:  "scopeuser",
		Email:     "scope@example.com",
		FirstName: "Scope",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
		Scopes:    7, // Binary: 111 (first 3 scopes)
	}
	err := user.SetPassword("password123")
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)

	scopes := fetchedUser.GetScopes()
	assert.GreaterOrEqual(suite.T(), len(scopes), 3) // Should have at least 3 scopes based on value 7
}
