package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServerRepositoryTestSuite struct {
	suite.Suite
	repo repository.IServerRepository
	db   db.IDatabaseClient
}

func (suite *ServerRepositoryTestSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(suite.T(), err)

	gormDB.AutoMigrate(&models.Server{})

	suite.db = db.NewDatabaseWithGorm(gormDB)
	suite.repo = repository.NewServerRepository(suite.db)
}

func (suite *ServerRepositoryTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), sqlDB.Close())
}

func TestServerRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ServerRepositoryTestSuite))
}

func (suite *ServerRepositoryTestSuite) TestCreateServer() {
	server := &models.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)
	assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)
	assert.Equal(suite.T(), server.IPv4, fetchedServer.IPv4)
}

func (suite *ServerRepositoryTestSuite) TestGetByID() {
	server := &models.Server{
		ServerID:    "test-server-1",
		ServerName:  "Test Server 1",
		IPv4:        "192.168.1.1",
		Status:      models.ServerStatusOn,
		Description: "Test description",
		Location:    "Test Location",
		OS:          "Ubuntu 20.04",
		CPU:         4,
		RAM:         8,
		Disk:        100,
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)
	assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)
	assert.Equal(suite.T(), server.IPv4, fetchedServer.IPv4)
	assert.Equal(suite.T(), server.Status, fetchedServer.Status)

	// Test not found
	_, err = suite.repo.GetByID(suite.T().Context(), 9999)
	assert.Error(suite.T(), err)
}

func (suite *ServerRepositoryTestSuite) TestGetByServerID() {
	server := &models.Server{
		ServerID:   "unique-server-id",
		ServerName: "Test Server",
		IPv4:       "192.168.1.2",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Test successful retrieval
	fetchedServer, err := suite.repo.GetByServerID(suite.T().Context(), "unique-server-id")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)

	// Test not found
	_, err = suite.repo.GetByServerID(suite.T().Context(), "non-existent-id")
	assert.Error(suite.T(), err)
}

func (suite *ServerRepositoryTestSuite) TestGetByServerName() {
	server := &models.Server{
		ServerID:   "test-server-2",
		ServerName: "Unique Server Name",
		IPv4:       "192.168.1.3",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Test successful retrieval
	fetchedServer, err := suite.repo.GetByServerName(suite.T().Context(), "Unique Server Name")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)

	// Test not found
	_, err = suite.repo.GetByServerName(suite.T().Context(), "Non-existent Name")
	assert.Error(suite.T(), err)
}

func (suite *ServerRepositoryTestSuite) TestUpdateServer() {
	server := &models.Server{
		ServerID:   "test-server-3",
		ServerName: "Original Server Name",
		IPv4:       "192.168.1.4",
		Status:     models.ServerStatusOff,
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Update server
	server.ServerName = "Updated Server Name"
	server.Status = models.ServerStatusOn
	server.Description = "Updated description"
	err = suite.repo.Update(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Verify update
	fetchedServer, err := suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Server Name", fetchedServer.ServerName)
	assert.Equal(suite.T(), models.ServerStatusOn, fetchedServer.Status)
	assert.Equal(suite.T(), "Updated description", fetchedServer.Description)
}

func (suite *ServerRepositoryTestSuite) TestUpdateStatus() {
	server := &models.Server{
		ServerID:   "test-server-4",
		ServerName: "Test Server 4",
		IPv4:       "192.168.1.5",
		Status:     models.ServerStatusOff,
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Update status
	err = suite.repo.UpdateStatus(suite.T().Context(), server.ServerID, models.ServerStatusOn)
	assert.NoError(suite.T(), err)

	// Verify status update
	fetchedServer, err := suite.repo.GetByServerID(suite.T().Context(), server.ServerID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.ServerStatusOn, fetchedServer.Status)
}

func (suite *ServerRepositoryTestSuite) TestDeleteServer() {
	server := &models.Server{
		ServerID:   "test-server-5",
		ServerName: "Test Server 5",
		IPv4:       "192.168.1.6",
	}
	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Delete server
	err = suite.repo.Delete(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.Error(suite.T(), err)
}

func (suite *ServerRepositoryTestSuite) TestBatchCreate() {
	servers := []models.Server{
		{
			ServerID:   "batch-server-1",
			ServerName: "Batch Server 1",
			IPv4:       "192.168.2.1",
		},
		{
			ServerID:   "batch-server-2",
			ServerName: "Batch Server 2",
			IPv4:       "192.168.2.2",
		},
		{
			ServerID:   "batch-server-3",
			ServerName: "Batch Server 3",
			IPv4:       "192.168.2.3",
		},
	}

	err := suite.repo.BatchCreate(suite.T().Context(), servers)
	assert.NoError(suite.T(), err)

	// Verify all servers were created
	for _, server := range servers {
		fetchedServer, err := suite.repo.GetByServerID(suite.T().Context(), server.ServerID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), server.ServerID, fetchedServer.ServerID)
		assert.Equal(suite.T(), server.ServerName, fetchedServer.ServerName)
	}
}

func (suite *ServerRepositoryTestSuite) TestCountByStatus() {
	// Create servers with different statuses
	servers := []models.Server{
		{ServerID: "on-server-1", ServerName: "On Server 1", IPv4: "192.168.3.1", Status: models.ServerStatusOn},
		{ServerID: "on-server-2", ServerName: "On Server 2", IPv4: "192.168.3.2", Status: models.ServerStatusOn},
		{ServerID: "off-server-1", ServerName: "Off Server 1", IPv4: "192.168.3.3", Status: models.ServerStatusOff},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), &server)
		assert.NoError(suite.T(), err)
	}

	// Count ON servers
	countOn, err := suite.repo.CountByStatus(suite.T().Context(), models.ServerStatusOn)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), countOn)

	// Count OFF servers
	countOff, err := suite.repo.CountByStatus(suite.T().Context(), models.ServerStatusOff)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), countOff)
}

func (suite *ServerRepositoryTestSuite) TestCountAll() {
	// Create some test servers
	servers := []models.Server{
		{ServerID: "count-server-1", ServerName: "Count Server 1", IPv4: "192.168.4.1"},
		{ServerID: "count-server-2", ServerName: "Count Server 2", IPv4: "192.168.4.2"},
		{ServerID: "count-server-3", ServerName: "Count Server 3", IPv4: "192.168.4.3"},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), &server)
		assert.NoError(suite.T(), err)
	}

	count, err := suite.repo.CountAll(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), count, int64(3))
}

func (suite *ServerRepositoryTestSuite) TestGetAll() {
	// Clear any existing data by getting current count
	initialCount, err := suite.repo.CountAll(suite.T().Context())
	assert.NoError(suite.T(), err)

	// Create test servers
	servers := []models.Server{
		{ServerID: "all-server-1", ServerName: "All Server 1", IPv4: "192.168.5.1"},
		{ServerID: "all-server-2", ServerName: "All Server 2", IPv4: "192.168.5.2"},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), &server)
		assert.NoError(suite.T(), err)
	}

	allServers, err := suite.repo.GetAll(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int(initialCount+2), len(allServers))

	// Close the database connection
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), sqlDB.Close())

	allServers, err = suite.repo.GetAll(suite.T().Context())
	assert.Error(suite.T(), err)      // Should return an error since the DB is closed
	assert.Nil(suite.T(), allServers) // Should return nil
}

func (suite *ServerRepositoryTestSuite) TestList() {
	// Create test servers with different attributes for filtering
	servers := []models.Server{
		{
			ServerID:   "list-server-1",
			ServerName: "Web Server 1",
			IPv4:       "192.168.6.1",
			Status:     models.ServerStatusOn,
			Location:   "DataCenter A",
			OS:         "Ubuntu 20.04",
		},
		{
			ServerID:   "list-server-2",
			ServerName: "Database Server",
			IPv4:       "192.168.6.2",
			Status:     models.ServerStatusOff,
			Location:   "DataCenter B",
			OS:         "CentOS 7",
		},
		{
			ServerID:   "list-server-3",
			ServerName: "Web Server 2",
			IPv4:       "192.168.6.3",
			Status:     models.ServerStatusOn,
			Location:   "DataCenter A",
			OS:         "Ubuntu 22.04",
		},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), &server)
		assert.NoError(suite.T(), err)
	}

	// Test with no filter
	filter := dto.ServerFilter{}
	pagination := dto.Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "created_time",
		Order:    "desc",
	}

	result, total, err := suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), total, int64(3))
	assert.GreaterOrEqual(suite.T(), len(result), 3)

	// Test with ServerID filter
	filter.ServerID = "list-server-1"
	result, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), 1, len(result))
	assert.Equal(suite.T(), "list-server-1", result[0].ServerID)

	// Test with ServerName filter
	filter = dto.ServerFilter{ServerName: "Web Server"}
	result, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), total)
	assert.Equal(suite.T(), 2, len(result))

	// Test with Status filter
	filter = dto.ServerFilter{Status: models.ServerStatusOn}
	_, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), total, int64(2))

	// Test with IPv4 filter
	filter = dto.ServerFilter{IPv4: "192.168.6.3"}
	_, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)

	// Test with Location filter
	filter = dto.ServerFilter{Location: "DataCenter A"}
	_, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), total)

	// Test with OS filter
	filter = dto.ServerFilter{OS: "Ubuntu"}
	_, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), total)

	// Test pagination
	filter = dto.ServerFilter{}
	pagination = dto.Pagination{
		Page:     1,
		PageSize: 1,
		Sort:     "server_id",
		Order:    "asc",
	}
	result, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), total, int64(3))
	assert.Equal(suite.T(), 1, len(result))
}

func (suite *ServerRepositoryTestSuite) TestCreateServerWithDuplicateServerID() {
	server1 := &models.Server{
		ServerID:   "duplicate-id",
		ServerName: "Server 1",
		IPv4:       "192.168.7.1",
	}
	err := suite.repo.Create(suite.T().Context(), server1)
	assert.NoError(suite.T(), err)

	// Try to create another server with the same ServerID
	server2 := &models.Server{
		ServerID:   "duplicate-id",
		ServerName: "Server 2",
		IPv4:       "192.168.7.2",
	}
	err = suite.repo.Create(suite.T().Context(), server2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *ServerRepositoryTestSuite) TestCreateServerWithDuplicateServerName() {
	server1 := &models.Server{
		ServerID:   "unique-id-1",
		ServerName: "Duplicate Name",
		IPv4:       "192.168.8.1",
	}
	err := suite.repo.Create(suite.T().Context(), server1)
	assert.NoError(suite.T(), err)

	// Try to create another server with the same ServerName
	server2 := &models.Server{
		ServerID:   "unique-id-2",
		ServerName: "Duplicate Name",
		IPv4:       "192.168.8.2",
	}
	err = suite.repo.Create(suite.T().Context(), server2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}
