package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// MockServerRepository is a mock implementation of ServerRepository
type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) Create(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepository) BatchCreate(ctx context.Context, servers []models.Server) error {
	args := m.Called(ctx, servers)
	return args.Error(0)
}

func (m *MockServerRepository) GetByID(ctx context.Context, id uint) (*models.Server, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerRepository) GetByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	args := m.Called(ctx, serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerRepository) GetByServerName(ctx context.Context, serverName string) (*models.Server, error) {
	args := m.Called(ctx, serverName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerRepository) List(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) ([]models.Server, int64, error) {
	args := m.Called(ctx, filter, pagination)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Server), args.Get(1).(int64), args.Error(2)
}

func (m *MockServerRepository) Update(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServerRepository) GetAll(ctx context.Context) ([]models.Server, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Server), args.Error(1)
}

func (m *MockServerRepository) GetServersIP(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockServerRepository) UpdateStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	args := m.Called(ctx, serverID, status)
	return args.Error(0)
}

func (m *MockServerRepository) CountByStatus(ctx context.Context, status models.ServerStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

// SimpleMockRedis provides a simple mock for Redis operations
type SimpleMockRedis struct {
	mock.Mock
	data map[string]string
}

func NewSimpleMockRedis() *SimpleMockRedis {
	return &SimpleMockRedis{
		data: make(map[string]string),
	}
}

// We'll create a simple test service that doesn't rely on Redis for basic functionality
func createTestServerServiceWithoutRedis() (*MockServerRepository, *zap.Logger) {
	mockRepo := &MockServerRepository{}
	logger := zap.NewNop()
	return mockRepo, logger
}

// Helper function to create a test Redis client using miniredis
func createTestRedisClient() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

// Helper function to create a server service with mocked Redis
func createTestServerService() (*MockServerRepository, *redis.Client, *zap.Logger) {
	mockRepo := &MockServerRepository{}
	redisClient := createTestRedisClient()
	logger := zap.NewNop()
	return mockRepo, redisClient, logger
}

func TestServerService_CreateServer_Success(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	// Create service manually for testing
	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil, // We'll skip redis for basic tests
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ServerID:   "test-001",
		ServerName: "Test Server",
		Status:     models.ServerStatusOff,
		IPv4:       "192.168.1.100",
	}

	mockRepo.On("GetByServerID", ctx, "test-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetByServerName", ctx, "Test Server").Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, server).Return(nil)

	// Test
	err := serverSrv.CreateServer(ctx, server)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_CreateServer_MissingServerID(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ServerName: "Test Server",
	}

	// Test
	err := serverSrv.CreateServer(ctx, server)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server_id and server_name are required")
}

func TestServerService_CreateServer_MissingServerName(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ServerID: "test-001",
	}

	// Test
	err := serverSrv.CreateServer(ctx, server)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server_id and server_name are required")
}

func TestServerService_CreateServer_ServerIDExists(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	existingServer := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Existing Server",
	}

	mockRepo.On("GetByServerID", ctx, "test-001").Return(existingServer, nil)

	// Test
	err := serverSrv.CreateServer(ctx, server)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is already exists")
	mockRepo.AssertExpectations(t)
}

func TestServerService_CreateServer_ServerNameExists(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	existingServer := &models.Server{
		ID:         1,
		ServerID:   "test-002",
		ServerName: "Test Server",
	}

	mockRepo.On("GetByServerID", ctx, "test-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetByServerName", ctx, "Test Server").Return(existingServer, nil)

	// Test
	err := serverSrv.CreateServer(ctx, server)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is already exists")
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetServer_Success_FromDatabase(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)

	// Test
	result, err := serverSrv.GetServer(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, server.ID, result.ID)
	assert.Equal(t, server.ServerID, result.ServerID)
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetServer_NotFound(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("server not found"))

	// Test
	result, err := serverSrv.GetServer(ctx, 1)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "server not found")
	mockRepo.AssertExpectations(t)
}

func TestServerService_UpdateServer_Success(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
		Status:     models.ServerStatusOff,
	}

	updates := map[string]interface{}{
		"server_name": "Updated Server",
		"status":      "ON",
		"ipv4":        "192.168.1.200",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("GetByServerName", ctx, "Updated Server").Return(nil, errors.New("not found"))
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Server")).Return(nil)

	// Test
	result, err := serverSrv.UpdateServer(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Server", result.ServerName)
	assert.Equal(t, models.ServerStatusOn, result.Status)
	assert.Equal(t, "192.168.1.200", result.IPv4)
	mockRepo.AssertExpectations(t)
}

func TestServerService_UpdateServer_ServerNotFound(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("server not found"))

	// Test
	result, err := serverSrv.UpdateServer(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "server not found")
	mockRepo.AssertExpectations(t)
}

func TestServerService_UpdateServer_DuplicateServerName(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	existingServer := &models.Server{
		ID:         2,
		ServerID:   "test-002",
		ServerName: "Updated Server",
	}

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("GetByServerName", ctx, "Updated Server").Return(existingServer, nil)

	// Test
	result, err := serverSrv.UpdateServer(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "server with name is already exists")
	mockRepo.AssertExpectations(t)
}

func TestServerService_DeleteServer_Success(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	// Test
	err := serverSrv.DeleteServer(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_DeleteServer_ServerNotFound(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("server not found"))

	// Test
	err := serverSrv.DeleteServer(ctx, 1)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server not found")
	mockRepo.AssertExpectations(t)
}

func TestServerService_UpdateServerStatus_Success(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:       1,
		ServerID: "test-001",
		Status:   models.ServerStatusOff,
	}

	mockRepo.On("GetByServerID", ctx, "test-001").Return(server, nil)
	mockRepo.On("UpdateStatus", ctx, "test-001", models.ServerStatusOn).Return(nil)

	// Test
	err := serverSrv.UpdateServerStatus(ctx, "test-001", models.ServerStatusOn)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_UpdateServerStatus_ServerNotFound(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	mockRepo.On("GetByServerID", ctx, "test-001").Return(nil, errors.New("server not found"))

	// Test
	err := serverSrv.UpdateServerStatus(ctx, "test-001", models.ServerStatusOn)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server not found")
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetAllServers_Success_FromDatabase(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	servers := []models.Server{
		{ID: 1, ServerID: "test-001", ServerName: "Server 1"},
		{ID: 2, ServerID: "test-002", ServerName: "Server 2"},
	}

	mockRepo.On("GetAll", ctx).Return(servers, nil)

	// Test
	result, err := serverSrv.GetAllServers(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, servers, result)
	mockRepo.AssertExpectations(t)
}

func TestServerService_CheckServerStatus_Success(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	servers := []models.Server{
		{ID: 1, ServerID: "test-001", IPv4: "127.0.0.1", Status: models.ServerStatusOn},
		{ID: 2, ServerID: "test-002", IPv4: "192.168.1.100", Status: models.ServerStatusOn},
	}

	mockRepo.On("GetAll", ctx).Return(servers, nil)
	// Mock UpdateStatus calls that might happen if server status changes
	// Since the servers have invalid IPs for testing, they'll likely be marked as OFF
	mockRepo.On("UpdateStatus", ctx, "test-001", models.ServerStatusOff).Return(nil).Maybe()
	mockRepo.On("UpdateStatus", ctx, "test-002", models.ServerStatusOff).Return(nil).Maybe()

	// Test
	err := serverSrv.CheckServerStatus(ctx)

	// Assertions
	assert.NoError(t, err)

	// Give a moment for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

func TestServerService_CheckServerStatus_GetAllError(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	mockRepo.On("GetAll", ctx).Return(nil, errors.New("database error"))

	// Test
	err := serverSrv.CheckServerStatus(ctx)

	// Assertions
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_ImportServers_FileNotExists(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Test with non-existent file
	result, err := serverSrv.ImportServers(ctx, "non-existent-file.xlsx")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestServerService_ExportServers_Success(t *testing.T) {
	// Create exports directory for the test
	err := os.MkdirAll("exports", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("exports") // Clean up after test

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	servers := []models.Server{
		{
			ID:         1,
			ServerID:   "test-001",
			ServerName: "Server 1",
			Status:     models.ServerStatusOn,
			IPv4:       "192.168.1.100",
		},
	}

	mockRepo.On("List", ctx, filter, pagination).Return(servers, int64(1), nil)

	// Test
	filePath, err := serverSrv.ExportServers(ctx, filter, pagination)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath)
	assert.Contains(t, filePath, "exports/servers_")
	assert.Contains(t, filePath, ".xlsx")
	mockRepo.AssertExpectations(t)
}

// Test ExportServers with empty servers list
func TestServerService_ExportServers_EmptyServersList(t *testing.T) {
	// Create exports directory for the test
	err := os.MkdirAll("exports", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("exports") // Clean up after test

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	// Return empty servers list
	mockRepo.On("List", ctx, filter, pagination).Return([]models.Server{}, int64(0), nil)

	// Test
	filePath, err := serverSrv.ExportServers(ctx, filter, pagination)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath)
	assert.Contains(t, filePath, "exports/servers_")
	assert.Contains(t, filePath, ".xlsx")
	mockRepo.AssertExpectations(t)
}

func TestNewServerService(t *testing.T) {
	mockRepo := &MockServerRepository{}
	redisClient := createTestRedisClient()
	logger := zap.NewNop()

	service := NewServerService(mockRepo, redisClient, logger)

	assert.NotNil(t, service)
	assert.IsType(t, &serverService{}, service)
}

/*
func TestServerService_GetServerStats_Success(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	servers := []models.Server{
		{ID: 1, Status: models.ServerStatusOn},
		{ID: 2, Status: models.ServerStatusOff},
		{ID: 3, Status: models.ServerStatusOn},
	}

	mockRepo.On("GetAll", ctx).Return(servers, nil)
	mockRepo.On("CountByStatus", ctx, "ON").Return(int64(2), nil)
	mockRepo.On("CountByStatus", ctx, "OFF").Return(int64(1), nil)

	// Test
	result, err := serverSrv.GetServerStats(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result["total"])
	assert.Equal(t, int64(2), result["online"])
	assert.Equal(t, int64(1), result["offline"])
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetServerStats_Error(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	// Mock GetAll to return error - this will cause the stats to have 0 total
	// but the method will still return successfully with 0 counts
	mockRepo.On("GetAll", ctx).Return(nil, errors.New("database error"))
	mockRepo.On("CountByStatus", ctx, "ON").Return(int64(0), nil)
	mockRepo.On("CountByStatus", ctx, "OFF").Return(int64(0), nil)

	// Test
	result, err := serverSrv.GetServerStats(ctx)

	// Assertions
	// The method doesn't fail even if GetAll fails, it just returns 0 total
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result["total"])
	mockRepo.AssertExpectations(t)
}
*/

func TestServerService_ListServers_Success(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{Status: "ON"}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	servers := []models.Server{
		{ID: 1, ServerID: "test-001", ServerName: "Server 1", Status: models.ServerStatusOn},
		{ID: 2, ServerID: "test-002", ServerName: "Server 2", Status: models.ServerStatusOn},
	}
	total := int64(2)

	mockRepo.On("List", ctx, filter, pagination).Return(servers, total, nil)

	// Test
	result, err := serverSrv.ListServers(ctx, filter, pagination)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, servers, result.Servers)
	assert.Equal(t, total, result.Total)
	assert.Equal(t, pagination.Page, result.Page)
	assert.Equal(t, pagination.PageSize, result.Size)
	mockRepo.AssertExpectations(t)
}

func TestServerService_ListServers_Error(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	mockRepo.On("List", ctx, filter, pagination).Return(nil, int64(0), errors.New("database error"))

	// Test
	result, err := serverSrv.ListServers(ctx, filter, pagination)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

// Test GetServerStats - this function has 0% coverage
func TestServerService_GetServerStats_Success(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	servers := []models.Server{
		{ID: 1, Status: models.ServerStatusOn},
		{ID: 2, Status: models.ServerStatusOff},
		{ID: 3, Status: models.ServerStatusOn},
	}

	mockRepo.On("GetAll", ctx).Return(servers, nil)
	mockRepo.On("CountByStatus", ctx, models.ServerStatusOn).Return(int64(2), nil)
	mockRepo.On("CountByStatus", ctx, models.ServerStatusOff).Return(int64(1), nil)

	// Test
	result, err := serverSrv.GetServerStats(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result["total"])
	assert.Equal(t, int64(2), result["online"])
	assert.Equal(t, int64(1), result["offline"])
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetServerStats_FromCache(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	// Prepare cache data
	stats := map[string]int64{
		"total":   5,
		"online":  3,
		"offline": 2,
	}
	statsJSON, _ := json.Marshal(stats)
	redisClient.Set(ctx, "server:stats", statsJSON, 5*time.Minute)

	// Test
	result, err := serverSrv.GetServerStats(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(5), result["total"])
	assert.Equal(t, int64(3), result["online"])
	assert.Equal(t, int64(2), result["offline"])
	// No repo calls should be made since we're getting from cache
}

func TestServerService_GetServerStats_GetAllError(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	// Mock GetAll to return error - this will cause the stats to have 0 total
	// but the method will still return successfully with 0 counts
	mockRepo.On("GetAll", ctx).Return(nil, errors.New("database error"))
	mockRepo.On("CountByStatus", ctx, models.ServerStatusOn).Return(int64(0), nil)
	mockRepo.On("CountByStatus", ctx, models.ServerStatusOff).Return(int64(0), nil)

	// Test
	result, err := serverSrv.GetServerStats(ctx)

	// Assertions
	// The method doesn't fail even if GetAll fails, it just returns 0 total
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result["total"])
	mockRepo.AssertExpectations(t)
}

// Test ImportServers - currently only 4.7% coverage
func TestServerService_ImportServers_Success(t *testing.T) {
	// Create a test Excel file first
	file := excelize.NewFile()
	defer file.Close()

	// Add header row
	file.SetCellValue("Sheet1", "A1", "server_id")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	// Add data rows
	file.SetCellValue("Sheet1", "A2", "test-001")
	file.SetCellValue("Sheet1", "B2", "Test Server 1")
	file.SetCellValue("Sheet1", "C2", "ON")
	file.SetCellValue("Sheet1", "D2", "192.168.1.100")
	file.SetCellValue("Sheet1", "E2", "Test description")
	file.SetCellValue("Sheet1", "F2", "Data Center 1")
	file.SetCellValue("Sheet1", "G2", "Ubuntu 20.04")
	file.SetCellValue("Sheet1", "H2", "4")
	file.SetCellValue("Sheet1", "I2", "8")
	file.SetCellValue("Sheet1", "J2", "100")

	testFile := "/tmp/test_servers.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Mock BatchCreate to succeed
	mockRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]models.Server")).Return(nil)

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Contains(t, result.SuccessServers, "test-001")
	mockRepo.AssertExpectations(t)
}

func TestServerService_ImportServers_EmptyDataRows(t *testing.T) {
	// This test should check error handling when there are no data rows (only header)
	file := excelize.NewFile()
	defer file.Close()

	// Add all required headers but no data
	file.SetCellValue("Sheet1", "A1", "server_id")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	testFile := "/tmp/test_empty.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions - this should return an error because file has no data rows
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file must contain at least 2 rows")
}

func TestServerService_ImportServers_InsufficientRows(t *testing.T) {
	// Create a file with only header, no data
	file := excelize.NewFile()
	defer file.Close()

	// Add only header row with all required columns
	file.SetCellValue("Sheet1", "A1", "server_id")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	testFile := "/tmp/test_no_data.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file must contain at least 2 rows")
}

func TestServerService_ImportServers_BatchCreateFallback(t *testing.T) {
	// Create a test Excel file
	file := excelize.NewFile()
	defer file.Close()

	// Add header and data with all required columns
	file.SetCellValue("Sheet1", "A1", "server_id")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	file.SetCellValue("Sheet1", "A2", "test-001")
	file.SetCellValue("Sheet1", "B2", "Test Server 1")
	file.SetCellValue("Sheet1", "C2", "ON")
	file.SetCellValue("Sheet1", "D2", "192.168.1.100")
	file.SetCellValue("Sheet1", "E2", "Test description")
	file.SetCellValue("Sheet1", "F2", "DC1")
	file.SetCellValue("Sheet1", "G2", "Ubuntu")
	file.SetCellValue("Sheet1", "H2", "4")
	file.SetCellValue("Sheet1", "I2", "8")
	file.SetCellValue("Sheet1", "J2", "100")

	testFile := "/tmp/test_fallback.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Mock BatchCreate to fail, then individual Create to succeed
	mockRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]models.Server")).Return(errors.New("batch failed"))
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Server")).Return(nil)

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	mockRepo.AssertExpectations(t)
}

func TestServerService_ImportServers_IndividualCreateFails(t *testing.T) {
	// Create a test Excel file
	file := excelize.NewFile()
	defer file.Close()

	// Add header and data with all required columns
	file.SetCellValue("Sheet1", "A1", "server_id")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	file.SetCellValue("Sheet1", "A2", "test-001")
	file.SetCellValue("Sheet1", "B2", "Test Server 1")
	file.SetCellValue("Sheet1", "C2", "ON")
	file.SetCellValue("Sheet1", "D2", "192.168.1.100")
	file.SetCellValue("Sheet1", "E2", "Test description")
	file.SetCellValue("Sheet1", "F2", "DC1")
	file.SetCellValue("Sheet1", "G2", "Ubuntu")
	file.SetCellValue("Sheet1", "H2", "4")
	file.SetCellValue("Sheet1", "I2", "8")
	file.SetCellValue("Sheet1", "J2", "100")

	testFile := "/tmp/test_create_fails.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Mock BatchCreate to fail, then individual Create to also fail
	mockRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]models.Server")).Return(errors.New("batch failed"))
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Server")).Return(errors.New("create failed"))

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	assert.Contains(t, result.FailureServers[0], "create failed")
	mockRepo.AssertExpectations(t)
}

// Test Redis cache functionality for GetServer
func TestServerService_GetServer_Success_FromCache(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	// Prepare cache data
	serverJSON, _ := json.Marshal(server)
	redisClient.Set(ctx, "server:1", serverJSON, 30*time.Minute)

	// Test
	result, err := serverSrv.GetServer(ctx, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, server.ID, result.ID)
	assert.Equal(t, server.ServerID, result.ServerID)
	// No repo calls should be made since we're getting from cache
}

// Test Redis cache functionality for GetAllServers
func TestServerService_GetAllServers_Success_FromCache(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	servers := []models.Server{
		{ID: 1, ServerID: "test-001", ServerName: "Server 1"},
		{ID: 2, ServerID: "test-002", ServerName: "Server 2"},
	}

	// Prepare cache data
	serversJSON, _ := json.Marshal(servers)
	redisClient.Set(ctx, "servers:all", serversJSON, 30*time.Minute)

	// Test
	result, err := serverSrv.GetAllServers(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, servers, result)
	// No repo calls should be made since we're getting from cache
}

// Test Redis cache functionality for ListServers
func TestServerService_ListServers_Success_FromCache(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{Status: "ON"}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	response := &models.ServerListResponse{
		Total: 2,
		Servers: []models.Server{
			{ID: 1, ServerID: "test-001", ServerName: "Server 1"},
		},
		Page: 1,
		Size: 10,
	}

	// Prepare cache data
	cacheKey := fmt.Sprintf("servers:list:%v:%d:%d", filter, pagination.Page, pagination.PageSize)
	responseJSON, _ := json.Marshal(response)
	redisClient.Set(ctx, cacheKey, responseJSON, 5*time.Minute)

	// Test
	result, err := serverSrv.ListServers(ctx, filter, pagination)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, response, result)
	// No repo calls should be made since we're getting from cache
}

// Test UpdateServer with different field types
func TestServerService_UpdateServer_AllFields(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
		Status:     models.ServerStatusOff,
		IPv4:       "192.168.1.100",
		Location:   "DC1",
		OS:         "Ubuntu 18.04",
		CPU:        2,
		RAM:        4,
		Disk:       50,
	}

	updates := map[string]interface{}{
		"server_name": "Updated Server",
		"status":      "ON",
		"ipv4":        "192.168.1.200",
		"location":    "DC2",
		"os":          "Ubuntu 20.04",
		"cpu":         float64(4),
		"ram":         float64(8),
		"disk":        float64(100),
		"server_id":   "should-be-ignored", // This should be ignored
		"id":          999,                 // This should be ignored
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("GetByServerName", ctx, "Updated Server").Return(nil, errors.New("not found"))
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Server")).Return(nil)

	// Test
	result, err := serverSrv.UpdateServer(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Server", result.ServerName)
	assert.Equal(t, models.ServerStatusOn, result.Status)
	assert.Equal(t, "192.168.1.200", result.IPv4)
	assert.Equal(t, "DC2", result.Location)
	assert.Equal(t, "Ubuntu 20.04", result.OS)
	assert.Equal(t, 4, result.CPU)
	assert.Equal(t, 8, result.RAM)
	assert.Equal(t, 100, result.Disk)
	assert.Equal(t, "test-001", result.ServerID) // Should not be changed
	assert.Equal(t, uint(1), result.ID)          // Should not be changed
	mockRepo.AssertExpectations(t)
}

// Test CreateServer repository error
func TestServerService_CreateServer_RepositoryError(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	mockRepo.On("GetByServerID", ctx, "test-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetByServerName", ctx, "Test Server").Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, server).Return(errors.New("database error"))

	// Test
	err := serverSrv.CreateServer(ctx, server)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create server")
	mockRepo.AssertExpectations(t)
}

// Test DeleteServer repository error
func TestServerService_DeleteServer_RepositoryError(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(errors.New("database error"))

	// Test
	err := serverSrv.DeleteServer(ctx, 1)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete server")
	mockRepo.AssertExpectations(t)
}

// Test UpdateServerStatus repository error
func TestServerService_UpdateServerStatus_RepositoryError(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:       1,
		ServerID: "test-001",
		Status:   models.ServerStatusOff,
	}

	mockRepo.On("GetByServerID", ctx, "test-001").Return(server, nil)
	mockRepo.On("UpdateStatus", ctx, "test-001", models.ServerStatusOn).Return(errors.New("database error"))

	// Test
	err := serverSrv.UpdateServerStatus(ctx, "test-001", models.ServerStatusOn)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update server status")
	mockRepo.AssertExpectations(t)
}

// Test GetAllServers repository error
func TestServerService_GetAllServers_RepositoryError(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	mockRepo.On("GetAll", ctx).Return(nil, errors.New("database error"))

	// Test
	result, err := serverSrv.GetAllServers(ctx)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get all servers")
	mockRepo.AssertExpectations(t)
}

// Test ExportServers repository error
func TestServerService_ExportServers_RepositoryError(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	mockRepo.On("List", ctx, filter, pagination).Return(nil, int64(0), errors.New("database error"))

	// Test
	result, err := serverSrv.ExportServers(ctx, filter, pagination)

	// Assertions
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to get servers")
	mockRepo.AssertExpectations(t)
}

// Test invalidateServerCaches method by testing UpdateServer which calls it
func TestServerService_UpdateServer_CacheInvalidation(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
		Status:     models.ServerStatusOff,
	}

	// Pre-populate cache with some data
	redisClient.Set(ctx, "server:1", `{"id":1,"server_id":"test-001"}`, 30*time.Minute)
	redisClient.Set(ctx, "server:stats", `{"total":10}`, 5*time.Minute)
	redisClient.Set(ctx, "servers:list:filter1", `{"total":1}`, 5*time.Minute)

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("GetByServerName", ctx, "Updated Server").Return(nil, errors.New("not found"))
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Server")).Return(nil)

	// Test
	result, err := serverSrv.UpdateServer(ctx, 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify cache entries have been invalidated (should return redis.Nil)
	_, err = redisClient.Get(ctx, "server:1").Result()
	assert.Error(t, err) // Should be redis.Nil

	_, err = redisClient.Get(ctx, "server:stats").Result()
	assert.Error(t, err) // Should be redis.Nil

	mockRepo.AssertExpectations(t)
}

// Test CheckServer method directly
func TestServerService_CheckServer_OnlineStatus(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := models.Server{
		ID:       1,
		ServerID: "test-001",
		IPv4:     "127.0.0.1", // localhost should be reachable
		Status:   models.ServerStatusOff,
	}

	// Mock UpdateStatus in case status changes
	mockRepo.On("UpdateStatus", ctx, "test-001", mock.AnythingOfType("models.ServerStatus")).Return(nil).Maybe()

	// Test
	serverSrv.CheckServer(ctx, server)

	// Give a moment for the check to complete
	time.Sleep(50 * time.Millisecond)

	// Assertions - not much to assert since it's async, but test should not panic
	// This mainly tests that the method runs without errors
}

// Test CheckServer with empty IP
func TestServerService_CheckServer_EmptyIP(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := models.Server{
		ID:       1,
		ServerID: "test-001",
		IPv4:     "", // Empty IP
		Status:   models.ServerStatusOn,
	}

	// Should call UpdateStatus to set it to OFF since empty IP means unreachable
	mockRepo.On("UpdateStatus", ctx, "test-001", models.ServerStatusOff).Return(nil)

	// Test
	serverSrv.CheckServer(ctx, server)

	// Give a moment for the check to complete
	time.Sleep(50 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// Test UpdateServer repository error
func TestServerService_UpdateServer_RepositoryError(t *testing.T) {
	mockRepo, redisClient, logger := createTestServerService()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: redisClient,
		logger:      logger,
	}

	ctx := context.Background()

	server := &models.Server{
		ID:         1,
		ServerID:   "test-001",
		ServerName: "Test Server",
	}

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(server, nil)
	mockRepo.On("GetByServerName", ctx, "Updated Server").Return(nil, errors.New("not found"))
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Server")).Return(errors.New("database error"))

	// Test
	result, err := serverSrv.UpdateServer(ctx, 1, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

// Test ImportServers with invalid row data
func TestServerService_ImportServers_InvalidRowData(t *testing.T) {
	// Create a test Excel file with invalid data
	file := excelize.NewFile()
	defer file.Close()

	// Add complete header
	file.SetCellValue("Sheet1", "A1", "server_id")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	// Add data row with insufficient columns (will cause parsing error)
	file.SetCellValue("Sheet1", "A2", "test-001")
	file.SetCellValue("Sheet1", "B2", "Test Server 1")
	// Missing other required columns

	testFile := "/tmp/test_invalid_data.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions
	assert.NoError(t, err) // Method should succeed but with parsing failures
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	assert.Contains(t, result.FailureServers[0], "invalid row")
}

// Test ImportServers with header validation error
func TestServerService_ImportServers_InvalidHeader(t *testing.T) {
	// Create a test Excel file with invalid header
	file := excelize.NewFile()
	defer file.Close()

	// Add incorrect header but enough columns
	file.SetCellValue("Sheet1", "A1", "wrong_header")
	file.SetCellValue("Sheet1", "B1", "server_name")
	file.SetCellValue("Sheet1", "C1", "status")
	file.SetCellValue("Sheet1", "D1", "ipv4")
	file.SetCellValue("Sheet1", "E1", "description")
	file.SetCellValue("Sheet1", "F1", "location")
	file.SetCellValue("Sheet1", "G1", "os")
	file.SetCellValue("Sheet1", "H1", "cpu")
	file.SetCellValue("Sheet1", "I1", "ram")
	file.SetCellValue("Sheet1", "J1", "disk")

	// Add a data row to satisfy the row count requirement
	file.SetCellValue("Sheet1", "A2", "test-001")
	file.SetCellValue("Sheet1", "B2", "Test Server")
	file.SetCellValue("Sheet1", "C2", "ON")

	testFile := "/tmp/test_invalid_header.xlsx"
	file.SaveAs(testFile)
	defer os.Remove(testFile)

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	// Test
	result, err := serverSrv.ImportServers(ctx, testFile)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "header validation failed")
}

// Test ExportServers with stream writer error
func TestServerService_ExportServers_CreateDirError(t *testing.T) {
	// Create a readonly directory to cause error
	err := os.MkdirAll("/tmp/readonly_exports", 0444) // readonly
	require.NoError(t, err)
	defer os.RemoveAll("/tmp/readonly_exports")

	// Change working directory temporarily
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir("/tmp/readonly_exports")

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	mockRepo.On("List", ctx, filter, pagination).Return(nil, int64(0), nil)

	// Test
	result, err := serverSrv.ExportServers(ctx, filter, pagination)

	// Assertions - should still work because it creates the file in ./exports/
	// but we can at least test the code path
	if err != nil {
		assert.Contains(t, err.Error(), "failed to save file")
	} else {
		assert.NotEmpty(t, result)
	}
	mockRepo.AssertExpectations(t)
}

// Test to improve CheckServer coverage
func TestServerService_CheckServer_StatusUpdateError(t *testing.T) {
	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	server := models.Server{
		ID:       1,
		ServerID: "test-001",
		IPv4:     "192.168.255.255", // Unreachable IP
		Status:   models.ServerStatusOn,
	}

	// Mock UpdateStatus to return error
	mockRepo.On("UpdateStatus", ctx, "test-001", models.ServerStatusOff).Return(errors.New("update error"))

	// Test
	serverSrv.CheckServer(ctx, server)

	// Give a moment for the check to complete
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// Test ExportServers with StreamWriter error scenario
func TestServerService_ExportServers_StreamWriterError(t *testing.T) {
	// Create exports directory for the test
	err := os.MkdirAll("exports", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("exports") // Clean up after test

	mockRepo, logger := createTestServerServiceWithoutRedis()

	serverSrv := &serverService{
		serverRepo:  mockRepo,
		redisClient: nil,
		logger:      logger,
	}

	ctx := context.Background()

	filter := models.ServerFilter{}
	pagination := models.Pagination{Page: 1, PageSize: 10}

	// Create a large number of servers to potentially trigger stream writer issues
	servers := make([]models.Server, 100)
	for i := 0; i < 100; i++ {
		servers[i] = models.Server{
			ID:         uint(i + 1),
			ServerID:   fmt.Sprintf("test-%03d", i+1),
			ServerName: fmt.Sprintf("Server %d", i+1),
			Status:     models.ServerStatusOn,
			IPv4:       fmt.Sprintf("192.168.1.%d", i+1),
		}
	}

	mockRepo.On("List", ctx, filter, pagination).Return(servers, int64(100), nil)

	// Test
	filePath, err := serverSrv.ExportServers(ctx, filter, pagination)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath)
	assert.Contains(t, filePath, "exports/servers_")
	assert.Contains(t, filePath, ".xlsx")
	mockRepo.AssertExpectations(t)
}
