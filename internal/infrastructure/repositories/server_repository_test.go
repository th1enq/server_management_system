package repositories

import (
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"gorm.io/gorm"
)

type ServerRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.ServerRepository
}

// func (suite *ServerRepositoryTestSuite) SetupTest() {
// 	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent),
// 	})
// 	assert.NoError(suite.T(), err)
// 	err = gormDB.AutoMigrate(&models.Server{})
// 	assert.NoError(suite.T(), err)

// 	suite.db = gormDB
// 	suite.repo = NewServerRepository(gormDB)
// }
