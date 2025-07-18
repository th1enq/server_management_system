package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/query"
	repoInterFace "github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/models"
	"github.com/th1enq/server_management_system/internal/infrastructure/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo repoInterFace.UserRepository
	db   database.DatabaseClient
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	assert.NoError(suite.T(), err)

	gormDB.AutoMigrate(&models.User{})

	suite.db = database.NewDatabaseWithGorm(gormDB)
	suite.repo = repository.NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TestDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), sqlDB.Close())
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TestCreateUser_Success() {
	user := &entity.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, fetchedUser.Username)
	assert.Equal(suite.T(), user.Email, fetchedUser.Email)
}

func (suite *UserRepositoryTestSuite) TestListUsers() {
	pagination := query.Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "created_at",
		Order:    "desc",
	}
	user := &entity.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	users, err := suite.repo.List(suite.T().Context(), pagination)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 1)
	assert.Equal(suite.T(), "testuser", users[0].Username)
	assert.Equal(suite.T(), "test@example.com", users[0].Email)
}

func (suite *UserRepositoryTestSuite) TestDeleteUser() {
	user := &entity.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	err = suite.repo.Delete(suite.T().Context(), user.ID)
	assert.NoError(suite.T(), err)

	fetchedUser, err := suite.repo.GetByID(suite.T().Context(), user.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), fetchedUser)
}

func (suite *UserRepositoryTestSuite) TestExistsByUsernameOrEmail() {
	user := &entity.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	exists, err := suite.repo.ExistsByUserNameOrEmail(suite.T().Context(), user.Username, user.Email, 0)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	exists, err = suite.repo.ExistsByUserNameOrEmail(suite.T().Context(), "nonexistent", "nonexistent@example.com", 1)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestUpdate() {
	user := &entity.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := suite.repo.Update(suite.T().Context(), user)
	assert.NoError(suite.T(), err)

	fetchedUser, err := suite.repo.GetByUsername(suite.T().Context(), user.Username)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, fetchedUser.Username)
	assert.Equal(suite.T(), user.Email, fetchedUser.Email)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername_Fail() {
	_, err := suite.repo.GetByUsername(suite.T().Context(), "nonexistent")
	assert.Error(suite.T(), err)
}
