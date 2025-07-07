package services

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type ServerServiceTestSuite struct {
	suite.Suite
	serverService services.IServerService
	mockRepo      MockServerRepository
	mockCache     MockCacheClient
}

func (suite *ServerServiceTestSuite) SetupTest() {
	suite.mockRepo = MockServerRepository{}
	suite.mockCache = MockCacheClient{}
	suite.serverService = services.NewServerService(&suite.mockRepo, &suite.mockCache, zap.NewNop())
}

func TestServerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServerServiceTestSuite))
}

func (suite *ServerServiceTestSuite) TestCreateServer() {
	server := &models.Server{
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	suite.mockRepo.On("ExistsByServerIDOrServerName", mock.Anything, server.ServerID, server.ServerName).Return(false, nil)
	suite.mockRepo.On("Create", mock.Anything, server).Return(nil)

	err := suite.serverService.CreateServer(context.Background(), server)
	suite.NoError(err)

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestCreateServerExists() {
	server := &models.Server{
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	suite.mockRepo.On("ExistsByServerIDOrServerName", mock.Anything, server.ServerID, server.ServerName).Return(true, nil)
	err := suite.serverService.CreateServer(context.Background(), server)
	suite.Error(err)
	suite.EqualError(err, "server with ID 'server-123' or name 'Test Server' already exists")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestCreateServerFail() {
	server := &models.Server{
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	suite.mockRepo.On("ExistsByServerIDOrServerName", mock.Anything, server.ServerID, server.ServerName).Return(false, nil)
	suite.mockRepo.On("Create", mock.Anything, server).Return(errors.New("mock error"))
	err := suite.serverService.CreateServer(context.Background(), server)
	suite.Error(err)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestCreateServerErrCheck() {
	server := &models.Server{
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}
	suite.mockRepo.On("ExistsByServerIDOrServerName", mock.Anything, server.ServerID, server.ServerName).Return(false, errors.New("mock error"))
	err := suite.serverService.CreateServer(context.Background(), server)
	suite.Error(err)
	suite.EqualError(err, "failed to check if server exists: mock error")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServer() {
	server := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	suite.mockCache.On("Get", mock.Anything, "server:1", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(server, nil)
	suite.mockCache.On("Set", mock.Anything, "server:1", server, 30*time.Minute).Return(nil)

	result, err := suite.serverService.GetServer(context.Background(), 1)

	suite.NoError(err)
	suite.Equal(server, result)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerFromCache() {
	server := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	suite.mockCache.On("Get", mock.Anything, "server:1", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(**models.Server)
		*arg = server
	}).Return(nil)

	result, err := suite.serverService.GetServer(context.Background(), 1)

	suite.NoError(err)
	suite.Equal(server, result)
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerNotFound() {
	suite.mockCache.On("Get", mock.Anything, "server:1", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("server not found"))

	result, err := suite.serverService.GetServer(context.Background(), 1)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "server not found")
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestDeleteServer() {
	server := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(server, nil)
	suite.mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:1").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:stats").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:all").Return(nil)

	err := suite.serverService.DeleteServer(context.Background(), 1)

	suite.NoError(err)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestDeleteServerNotFound() {
	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("server not found"))

	err := suite.serverService.DeleteServer(context.Background(), 1)

	suite.Error(err)
	suite.Contains(err.Error(), "server not found")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestDeleteServerFail() {
	server := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(server, nil)
	suite.mockRepo.On("Delete", mock.Anything, uint(1)).Return(errors.New("delete failed"))

	err := suite.serverService.DeleteServer(context.Background(), 1)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete server")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServer() {
	existingServer := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Old Server",
		IPv4:       "192.168.1.1",
	}

	updates := dto.ServerUpdate{
		ServerName:  "New Server",
		Status:      models.ServerStatusOn,
		IPv4:        "192.168.1.2",
		Description: "Updated description",
		Location:    "DC1",
		OS:          "Ubuntu",
		CPU:         4,
		RAM:         8,
		Disk:        100,
	}

	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingServer, nil)
	suite.mockRepo.On("GetByServerName", mock.Anything, "New Server").Return(nil, errors.New("not found"))
	suite.mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(server *models.Server) bool {
		return server.ServerName == "New Server" && server.Status == models.ServerStatusOn
	})).Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:1").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:stats").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:all").Return(nil)

	result, err := suite.serverService.UpdateServer(context.Background(), 1, updates)

	suite.NoError(err)
	suite.Equal("New Server", result.ServerName)
	suite.Equal(models.ServerStatusOn, result.Status)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServerNotFound() {
	updates := dto.ServerUpdate{
		ServerName: "New Server",
	}

	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("server not found"))

	result, err := suite.serverService.UpdateServer(context.Background(), 1, updates)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "server not found")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServerNameExists() {
	existingServer := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Old Server",
	}

	conflictServer := &models.Server{
		ID:         2,
		ServerID:   "server-456",
		ServerName: "New Server",
	}

	updates := dto.ServerUpdate{
		ServerName: "New Server",
	}

	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingServer, nil)
	suite.mockRepo.On("GetByServerName", mock.Anything, "New Server").Return(conflictServer, nil)

	result, err := suite.serverService.UpdateServer(context.Background(), 1, updates)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "server name already exists")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestListServers() {
	servers := []models.Server{
		{ID: 1, ServerID: "server-1", ServerName: "Server 1"},
		{ID: 2, ServerID: "server-2", ServerName: "Server 2"},
	}

	filter := dto.ServerFilter{
		Status: models.ServerStatusOn,
	}
	pagination := dto.Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "created_time",
		Order:    "desc",
	}

	suite.mockRepo.On("List", mock.Anything, filter, pagination).Return(servers, int64(2), nil)

	result, err := suite.serverService.ListServers(context.Background(), filter, pagination)

	suite.NoError(err)
	suite.Equal(int64(2), result.Total)
	suite.Equal(len(servers), len(result.Servers))
	suite.Equal(1, result.Page)
	suite.Equal(10, result.Size)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestListServersFail() {
	filter := dto.ServerFilter{}
	pagination := dto.Pagination{}

	suite.mockRepo.On("List", mock.Anything, filter, pagination).Return(nil, int64(0), errors.New("database error"))

	result, err := suite.serverService.ListServers(context.Background(), filter, pagination)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to list servers")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetAllServers() {
	servers := []models.Server{
		{ID: 1, ServerID: "server-1", ServerName: "Server 1"},
		{ID: 2, ServerID: "server-2", ServerName: "Server 2"},
	}

	suite.mockCache.On("Get", mock.Anything, "server:all", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("GetAll", mock.Anything).Return(servers, nil)
	suite.mockCache.On("Set", mock.Anything, "server:all", servers, 30*time.Minute).Return(nil)

	result, err := suite.serverService.GetAllServers(context.Background())

	suite.NoError(err)
	suite.Equal(len(servers), len(result))
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetAllServersFromCache() {
	servers := []models.Server{
		{ID: 1, ServerID: "server-1", ServerName: "Server 1"},
		{ID: 2, ServerID: "server-2", ServerName: "Server 2"},
	}

	suite.mockCache.On("Get", mock.Anything, "server:all", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*[]models.Server)
		*arg = servers
	}).Return(nil)

	result, err := suite.serverService.GetAllServers(context.Background())

	suite.NoError(err)
	suite.Equal(len(servers), len(result))
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetAllServersFail() {
	suite.mockCache.On("Get", mock.Anything, "server:all", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))

	result, err := suite.serverService.GetAllServers(context.Background())

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to get all servers")
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerStats() {
	expectedStats := map[string]int64{
		"total":   10,
		"online":  7,
		"offline": 3,
	}

	suite.mockCache.On("Get", mock.Anything, "server:stats", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("CountAll", mock.Anything).Return(int64(10), nil)
	suite.mockRepo.On("CountByStatus", mock.Anything, models.ServerStatusOn).Return(int64(7), nil)
	suite.mockRepo.On("CountByStatus", mock.Anything, models.ServerStatusOff).Return(int64(3), nil)
	suite.mockCache.On("Set", mock.Anything, "server:stats", mock.MatchedBy(func(stats map[string]int64) bool {
		return stats["total"] == 10 && stats["online"] == 7 && stats["offline"] == 3
	}), 30*time.Minute).Return(nil)

	result, err := suite.serverService.GetServerStats(context.Background())

	suite.NoError(err)
	suite.Equal(expectedStats, result)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerStatsFromCache() {
	expectedStats := map[string]int64{
		"total":   10,
		"online":  7,
		"offline": 3,
	}

	suite.mockCache.On("Get", mock.Anything, "server:stats", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*map[string]int64)
		*arg = expectedStats
	}).Return(nil)

	result, err := suite.serverService.GetServerStats(context.Background())

	suite.NoError(err)
	suite.Equal(expectedStats, result)
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerStatsFail() {
	suite.mockCache.On("Get", mock.Anything, "server:stats", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("CountAll", mock.Anything).Return(int64(0), errors.New("database error"))

	result, err := suite.serverService.GetServerStats(context.Background())

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to count servers")
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServerStatus() {
	server := &models.Server{
		ID:       1,
		ServerID: "server-123",
	}

	suite.mockRepo.On("GetByServerID", mock.Anything, "server-123").Return(server, nil)
	suite.mockRepo.On("UpdateStatus", mock.Anything, "server-123", models.ServerStatusOn).Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:1").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:stats").Return(nil)
	suite.mockCache.On("Del", mock.Anything, "server:all").Return(nil)

	err := suite.serverService.UpdateServerStatus(context.Background(), "server-123", models.ServerStatusOn)

	suite.NoError(err)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServerStatusNotFound() {
	suite.mockRepo.On("GetByServerID", mock.Anything, "server-123").Return(nil, errors.New("server not found"))

	err := suite.serverService.UpdateServerStatus(context.Background(), "server-123", models.ServerStatusOn)

	suite.Error(err)
	suite.Contains(err.Error(), "server not found")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServerStatusFail() {
	server := &models.Server{
		ID:       1,
		ServerID: "server-123",
	}

	suite.mockRepo.On("GetByServerID", mock.Anything, "server-123").Return(server, nil)
	suite.mockRepo.On("UpdateStatus", mock.Anything, "server-123", models.ServerStatusOn).Return(errors.New("update failed"))

	err := suite.serverService.UpdateServerStatus(context.Background(), "server-123", models.ServerStatusOn)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update server status")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestExportServers() {
	// Note: This test calls the actual ExportServers method which creates real files
	// For a purely isolated unit test, the file creation logic should be extracted
	// into a separate interface that can be mocked

	servers := []models.Server{
		{
			ID:          1,
			ServerID:    "server-1",
			ServerName:  "Server 1",
			Status:      models.ServerStatusOn,
			Description: "Test server",
			IPv4:        "192.168.1.1",
			Disk:        100,
			RAM:         8,
			Location:    "DC1",
			OS:          "Ubuntu",
		},
	}

	filter := dto.ServerFilter{}
	pagination := dto.Pagination{}

	suite.mockRepo.On("List", mock.Anything, filter, pagination).Return(servers, int64(1), nil)

	// Create exports directory if it doesn't exist (for test environment)
	os.MkdirAll("./exports", 0755)

	result, err := suite.serverService.ExportServers(context.Background(), filter, pagination)

	suite.NoError(err)
	suite.Contains(result, "servers_")
	suite.Contains(result, ".xlsx")
	suite.mockRepo.AssertExpectations(suite.T())

	// Clean up created file
	if result != "" {
		os.Remove(result)
	}
}

func (suite *ServerServiceTestSuite) TestExportServersFail() {
	filter := dto.ServerFilter{}
	pagination := dto.Pagination{}

	suite.mockRepo.On("List", mock.Anything, filter, pagination).Return(nil, int64(0), errors.New("database error"))

	result, err := suite.serverService.ExportServers(context.Background(), filter, pagination)

	suite.Error(err)
	suite.Empty(result)
	suite.Contains(err.Error(), "failed to get servers")
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestCheckServerStatusFail() {
	suite.mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))

	err := suite.serverService.CheckServerStatus(context.Background())

	suite.Error(err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// func (suite *ServerServiceTestSuite) TestCheckServerSuccess() {
// 	servers := []models.Server{
// 		{
// 			ID:         1,
// 			ServerID:   "server-1",
// 			ServerName: "Server 1",
// 			IPv4:       "192.168.1.1",
// 		},
// 		{
// 			ID:         2,
// 			ServerID:   "server-2",
// 			ServerName: "Server 2",
// 			IPv4:       "192.168.1.2",
// 		},
// 	}

// 	suite.mockRepo.On("GetAll", mock.Anything).Return(servers, nil)
// 	suite.mockRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
// 	err := suite.serverService.CheckServerStatus(context.Background())
// 	suite.NoError(err)
// 	suite.mockRepo.AssertExpectations(suite.T())
// }

// func (suite *ServerServiceTestSuite) TestCheckServerFailed() {
// 	servers := []models.Server{
// 		{
// 			ID:         1,
// 			ServerID:   "server-1",
// 			ServerName: "Server 1",
// 			IPv4:       "192.168.1.1",
// 		},
// 		{
// 			ID:         2,
// 			ServerID:   "server-2",
// 			ServerName: "Server 2",
// 			IPv4:       "192.168.1.2",
// 		},
// 	}

// 	suite.mockRepo.On("GetAll", mock.Anything).Return(servers, nil)
// 	suite.mockRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update failed"))
// 	err := suite.serverService.CheckServerStatus(context.Background())
// 	suite.Error(err)
// }

func (suite *ServerServiceTestSuite) TestImportServersFileNotFound() {
	result, err := suite.serverService.ImportServers(context.Background(), "./non_existent_file.xlsx")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to open file")
}

func (suite *ServerServiceTestSuite) TestImportServersSuccess() {
	suite.mockRepo.On("BatchCreate", mock.Anything, mock.AnythingOfType("[]models.Server")).Return(nil)

	result, err := suite.serverService.ImportServers(context.Background(), "./imports/servers_10000.xlsx")

	suite.NoError(err)
	suite.NotNil(result)
}

func (suite *ServerServiceTestSuite) TestImportServersFailBatch() {
	suite.mockRepo.On("BatchCreate", mock.Anything, mock.AnythingOfType("[]models.Server")).Return(errors.New("batch create failed"))
	suite.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Server")).Return(nil).Maybe()

	result, err := suite.serverService.ImportServers(context.Background(), "./imports/servers_10000.xlsx")

	suite.NoError(err)
	suite.NotNil(result)
}

// Additional edge case tests
func (suite *ServerServiceTestSuite) TestGetServerCacheError() {
	server := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	// Mock cache to return a non-cache-miss error
	suite.mockCache.On("Get", mock.Anything, "server:1", mock.Anything).Return(errors.New("redis connection error"))
	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(server, nil)
	suite.mockCache.On("Set", mock.Anything, "server:1", server, 30*time.Minute).Return(nil)

	result, err := suite.serverService.GetServer(context.Background(), 1)

	suite.NoError(err)
	suite.Equal(server, result)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerCacheSetFail() {
	server := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
	}

	suite.mockCache.On("Get", mock.Anything, "server:1", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(server, nil)
	suite.mockCache.On("Set", mock.Anything, "server:1", server, 30*time.Minute).Return(errors.New("cache set failed"))

	result, err := suite.serverService.GetServer(context.Background(), 1)

	suite.NoError(err) // Should still succeed even if cache set fails
	suite.Equal(server, result)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestGetServerStatsCacheSetFail() {
	suite.mockCache.On("Get", mock.Anything, "server:stats", mock.Anything).Return(db.ErrCacheMiss)
	suite.mockRepo.On("CountAll", mock.Anything).Return(int64(10), nil)
	suite.mockRepo.On("CountByStatus", mock.Anything, models.ServerStatusOn).Return(int64(7), nil)
	suite.mockRepo.On("CountByStatus", mock.Anything, models.ServerStatusOff).Return(int64(3), nil)
	suite.mockCache.On("Set", mock.Anything, "server:stats", mock.AnythingOfType("map[string]int64"), 30*time.Minute).Return(errors.New("cache set failed"))

	result, err := suite.serverService.GetServerStats(context.Background())

	suite.Error(err) // Should fail if cache set fails in this case
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to cache server stats")
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

func (suite *ServerServiceTestSuite) TestUpdateServerUpdateFail() {
	existingServer := &models.Server{
		ID:         1,
		ServerID:   "server-123",
		ServerName: "Old Server",
	}

	updates := dto.ServerUpdate{
		ServerName: "New Server",
		Status:     models.ServerStatusOn,
	}

	suite.mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existingServer, nil)
	suite.mockRepo.On("GetByServerName", mock.Anything, "New Server").Return(nil, errors.New("not found"))
	suite.mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Server")).Return(errors.New("update failed"))

	result, err := suite.serverService.UpdateServer(context.Background(), 1, updates)

	suite.Error(err)
	suite.Nil(result)
	suite.mockRepo.AssertExpectations(suite.T())
}
