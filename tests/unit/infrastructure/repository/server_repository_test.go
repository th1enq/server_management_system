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

type ServerRepositoryTestSuite struct {
	suite.Suite
	repo repoInterFace.ServerRepository
	db   database.DatabaseClient
}

func (suite *ServerRepositoryTestSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	assert.NoError(suite.T(), err)

	gormDB.AutoMigrate(&models.Server{})

	suite.db = database.NewDatabaseWithGorm(gormDB)
	suite.repo = repository.NewServerRepository(suite.db)
}

func (suite *ServerRepositoryTestSuite) TestDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), sqlDB.Close())
}

func TestServerRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ServerRepositoryTestSuite))
}

func (suite *ServerRepositoryTestSuite) TestCreateServer_Success() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)
	assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)
	assert.Equal(suite.T(), server.IPv4, fetchedServer.IPv4)
}

func (suite *ServerRepositoryTestSuite) TestCreateServer_Fail() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), 0)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), fetchedServer)
}

func (suite *ServerRepositoryTestSuite) TestBatchCreateServers() {
	servers := []entity.Server{
		{ServerID: "server1", ServerName: "Server One", IPv4: "192.168.1.1"},
		{ServerID: "server2", ServerName: "Server Two", IPv4: "192.168.1.2"},
	}
	err := suite.repo.BatchCreate(suite.T().Context(), servers)
	assert.NoError(suite.T(), err)
	fetchedServers, err := suite.repo.GetAll(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), fetchedServers, 2)
	assert.Equal(suite.T(), "server1", fetchedServers[0].ServerID)
	assert.Equal(suite.T(), "server2", fetchedServers[1].ServerID)
}

func (suite *ServerRepositoryTestSuite) TestCountByStatus() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountByStatus(suite.T().Context(), entity.ServerStatusOff)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *ServerRepositoryTestSuite) TestDeleteServer() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	err = suite.repo.Delete(suite.T().Context(), 1)
	assert.NoError(suite.T(), err)
	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), 1)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), fetchedServer)
}

func (suite *ServerRepositoryTestSuite) TestCountAllServers() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	count, err := suite.repo.CountAll(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *ServerRepositoryTestSuite) TestGetByID() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)
	assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)
	assert.Equal(suite.T(), server.IPv4, fetchedServer.IPv4)
}

func (suite *ServerRepositoryTestSuite) TestGetByServerName() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	fetchedServer, err := suite.repo.GetByServerName(suite.T().Context(), "Test Server")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)
	assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)
	assert.Equal(suite.T(), server.IPv4, fetchedServer.IPv4)
}

func (suite *ServerRepositoryTestSuite) TestListServers() {
	server1 := &entity.Server{
		ServerID:   "server1",
		ServerName: "Server One",
		IPv4:       "192.168.1.1",
		Location:   "Location One",
		OS:         "Linux",
		Status:     entity.ServerStatusOff,
		CPU:        2,
		RAM:        4,
		Disk:       100,
	}
	err := suite.repo.Create(suite.T().Context(), server1)
	assert.NoError(suite.T(), err)

	filter := query.ServerFilter{
		ServerID:   "server1",
		ServerName: "Server One",
		IPv4:       "192.168.1.1",
		Status:     entity.ServerStatusOff,
		Location:   "Location One",
		OS:         "Linux",
		CPU:        2,
		RAM:        4,
		Disk:       100,
	}
	pagination := query.Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "server_name",
		Order:    "asc",
	}
	servers, total, err := suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Len(suite.T(), servers, 1)
	assert.Equal(suite.T(), server1.ServerID, servers[0].ServerID)
	assert.Equal(suite.T(), server1.ServerName, servers[0].ServerName)
	assert.Equal(suite.T(), server1.IPv4, servers[0].IPv4)
	assert.Equal(suite.T(), server1.Location, servers[0].Location)
	assert.Equal(suite.T(), server1.OS, servers[0].OS)
	assert.Equal(suite.T(), server1.Status, servers[0].Status)
	assert.Equal(suite.T(), server1.CPU, servers[0].CPU)
	assert.Equal(suite.T(), server1.RAM, servers[0].RAM)
	assert.Equal(suite.T(), server1.Disk, servers[0].Disk)
}

func (suite *ServerRepositoryTestSuite) TestUpdateServer() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	fetchedServer, err := suite.repo.GetByServerID(suite.T().Context(), "test-server")
	assert.NoError(suite.T(), err)
	fetchedServer.ServerName = "Updated Server"
	err = suite.repo.Update(suite.T().Context(), fetchedServer)
	assert.NoError(suite.T(), err)
	updatedServer, err := suite.repo.GetByID(suite.T().Context(), fetchedServer.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Server", updatedServer.ServerName)
}

func (suite *ServerRepositoryTestSuite) TestUpdateStatus() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	err = suite.repo.UpdateStatus(suite.T().Context(), server.ServerID, entity.ServerStatusOn)
	assert.NoError(suite.T(), err)
	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), entity.ServerStatusOn, fetchedServer.Status)
}

func (suite *ServerRepositoryTestSuite) TestExistsByServerIDOrServerName() {
	server := &entity.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	exists, err := suite.repo.ExistsByServerIDOrServerName(suite.T().Context(), "test-server", "Test Server")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	exists, err = suite.repo.ExistsByServerIDOrServerName(suite.T().Context(), "non-existent-server", "Non Existent Server")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}
