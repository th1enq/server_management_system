package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserRepositorySuite struct {
	suite.Suite
	repo UserRepository
	db   *gorm.DB
}

func (suite *UserRepositorySuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	err = db.AutoMigrate(&models.User{})
	assert.NoError(suite.T(), err)

	suite.db = db
	suite.repo = NewUserRepository(db)
}

func (suite *UserRepositorySuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *UserRepositorySuite) TestCreateUser() {
	user := &models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "example@gmail.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), user.ID)
	assert.NotEqual(suite.T(), "password123", user.Password) // Password should be hashed

	// Verify password was hashed correctly
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
	assert.NoError(suite.T(), err)
}

func (suite *UserRepositorySuite) TestCreateUserWithDuplicateUsername() {
	user1 := &models.User{
		Username: "duplicate",
		Password: "password123",
		Email:    "user1@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	user2 := &models.User{
		Username: "duplicate",
		Password: "password456",
		Email:    "user2@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	err := suite.repo.Create(suite.T().Context(), user1)
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *UserRepositorySuite) TestCreateUserWithDuplicateEmail() {
	user1 := &models.User{
		Username: "user1",
		Password: "password123",
		Email:    "duplicate@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	user2 := &models.User{
		Username: "user2",
		Password: "password456",
		Email:    "duplicate@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}

	err := suite.repo.Create(suite.T().Context(), user1)
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.T().Context(), user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *UserRepositorySuite) TestGetByUsername() {
	// Create a test user
	user := &models.User{
		Username:  "findme",
		Password:  "password123",
		Email:     "findme@example.com",
		FirstName: "Find",
		LastName:  "Me",
		Role:      models.RoleAdmin,
		IsActive:  true,
	}

	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Test finding the user
	foundUser, err := suite.repo.GetByUsername(suite.T().Context(), "findme")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	assert.Equal(suite.T(), user.Email, foundUser.Email)
	assert.Equal(suite.T(), user.FirstName, foundUser.FirstName)
	assert.Equal(suite.T(), user.LastName, foundUser.LastName)
	assert.Equal(suite.T(), user.Role, foundUser.Role)
}

func (suite *UserRepositorySuite) TestGetByUsernameNotFound() {
	foundUser, err := suite.repo.GetByUsername(suite.T().Context(), "nonexistent")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositorySuite) TestGetByID() {
	// Create a test user
	user := &models.User{
		Username:  "testbyid",
		Password:  "password123",
		Email:     "testbyid@example.com",
		FirstName: "Test",
		LastName:  "ByID",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Test finding the user by ID
	foundUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	assert.Equal(suite.T(), user.Email, foundUser.Email)
}

func (suite *UserRepositorySuite) TestGetByIDNotFound() {
	foundUser, err := suite.repo.GetByID(suite.T().Context(), 999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositorySuite) TestGetByEmail() {
	// Create a test user
	user := &models.User{
		Username:  "testbyemail",
		Password:  "password123",
		Email:     "testbyemail@example.com",
		FirstName: "Test",
		LastName:  "ByEmail",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Test finding the user by email
	foundUser, err := suite.repo.GetByEmail(suite.T().Context(), "testbyemail@example.com")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	assert.Equal(suite.T(), user.Email, foundUser.Email)
}

func (suite *UserRepositorySuite) TestGetByEmailNotFound() {
	foundUser, err := suite.repo.GetByEmail(suite.T().Context(), "nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositorySuite) TestUpdate() {
	// Create a test user
	user := &models.User{
		Username:  "testupdate",
		Password:  "password123",
		Email:     "testupdate@example.com",
		FirstName: "Original",
		LastName:  "Name",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Update the user
	user.FirstName = "Updated"
	user.LastName = "User"
	user.Role = models.RoleAdmin
	user.IsActive = false

	err = suite.repo.Update(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Verify the update
	updatedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated", updatedUser.FirstName)
	assert.Equal(suite.T(), "User", updatedUser.LastName)
	assert.Equal(suite.T(), models.RoleAdmin, updatedUser.Role)
	assert.False(suite.T(), updatedUser.IsActive)
}

func (suite *UserRepositorySuite) TestDelete() {
	// Create a test user
	user := &models.User{
		Username:  "testdelete",
		Password:  "password123",
		Email:     "testdelete@example.com",
		FirstName: "Delete",
		LastName:  "Me",
		Role:      models.RoleUser,
		IsActive:  true,
	}

	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	// Delete the user
	err = suite.repo.Delete(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)

	// Verify the user is deleted
	foundUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositorySuite) TestList() {
	// Create multiple test users
	users := []*models.User{
		{
			Username:  "user1",
			Password:  "password123",
			Email:     "user1@example.com",
			FirstName: "User",
			LastName:  "One",
			Role:      models.RoleUser,
			IsActive:  true,
		},
		{
			Username:  "user2",
			Password:  "password123",
			Email:     "user2@example.com",
			FirstName: "User",
			LastName:  "Two",
			Role:      models.RoleAdmin,
			IsActive:  true,
		},
		{
			Username:  "user3",
			Password:  "password123",
			Email:     "user3@example.com",
			FirstName: "User",
			LastName:  "Three",
			Role:      models.RoleUser,
			IsActive:  false,
		},
	}

	for _, user := range users {
		err := suite.repo.Create(suite.T().Context(), user)
		assert.NoError(suite.T(), err)
	}

	// Test list with limit and offset
	foundUsers, err := suite.repo.List(suite.T().Context(), 2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 2)

	// Test list with offset
	foundUsers, err = suite.repo.List(suite.T().Context(), 2, 1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 2)

	// Test list all
	foundUsers, err = suite.repo.List(suite.T().Context(), 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 3)
}

func (suite *UserRepositorySuite) TestListEmpty() {
	foundUsers, err := suite.repo.List(suite.T().Context(), 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 0)
}

// TestUserRepositorySuite runs the test suite
func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}
