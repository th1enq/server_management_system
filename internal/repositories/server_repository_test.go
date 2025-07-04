package repositories

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockPgxPool implements a simple mock for testing BatchCreate
type MockPgxPool struct {
	copyFromFunc func(ctx context.Context, tableName []string, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

func (m *MockPgxPool) CopyFrom(ctx context.Context, tableName []string, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	if m.copyFromFunc != nil {
		return m.copyFromFunc(ctx, tableName, columnNames, rowSrc)
	}
	return 0, nil
}

// ServerRepositoryRealTestSuite for testing with real server repository
type ServerRepositoryRealTestSuite struct {
	suite.Suite
	repo ServerRepository
	db   *gorm.DB
}

func (suite *ServerRepositoryRealTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	err = db.AutoMigrate(&models.Server{})
	assert.NoError(suite.T(), err)

	suite.db = db
	// Test with real serverRepository instead of TestServerRepository
	suite.repo = NewServerRepository(db, nil)
}

func (suite *ServerRepositoryRealTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// Test real server repository methods
func (suite *ServerRepositoryRealTestSuite) TestRealCreate() {
	server := &models.Server{
		ServerID:   "real-test-server",
		ServerName: "Real Test Server",
		IPv4:       "192.168.1.100",
		Status:     models.ServerStatusOn,
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), server.ID)
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetByID() {
	server := &models.Server{
		ServerID:   "real-test-server-2",
		ServerName: "Real Test Server 2",
		IPv4:       "192.168.1.101",
		Status:     models.ServerStatusOn,
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	result, err := suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, result.ServerID)
	assert.Equal(suite.T(), server.ServerName, result.ServerName)
}

func (suite *ServerRepositoryRealTestSuite) TestRealBatchCreate() {
	// Test BatchCreate with real repository (will use GORM fallback since no pgx pool)
	servers := []models.Server{
		{
			ServerID:   "real-batch-1",
			ServerName: "Real Batch Server 1",
			IPv4:       "10.0.0.10",
			Status:     models.ServerStatusOn,
		},
		{
			ServerID:   "real-batch-2",
			ServerName: "Real Batch Server 2",
			IPv4:       "10.0.0.11",
			Status:     models.ServerStatusOff,
		},
	}

	err := suite.repo.BatchCreate(suite.T().Context(), servers)
	assert.NoError(suite.T(), err)

	// Verify servers were created
	result1, err := suite.repo.GetByServerID(suite.T().Context(), "real-batch-1")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Real Batch Server 1", result1.ServerName)

	result2, err := suite.repo.GetByServerID(suite.T().Context(), "real-batch-2")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Real Batch Server 2", result2.ServerName)
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetByServerID() {
	server := &models.Server{
		ServerID:   "real-get-server-id",
		ServerName: "Real Get Server",
		IPv4:       "192.168.1.102",
		Status:     models.ServerStatusOn,
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	result, err := suite.repo.GetByServerID(suite.T().Context(), "real-get-server-id")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerName, result.ServerName)
}

func (suite *ServerRepositoryRealTestSuite) TestRealUpdate() {
	server := &models.Server{
		ServerID:   "real-update-server",
		ServerName: "Real Update Server",
		IPv4:       "192.168.1.103",
		Status:     models.ServerStatusOn,
		Location:   "DC1",
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Update the server
	server.ServerName = "Real Updated Server"
	server.Location = "DC2"
	server.Status = models.ServerStatusOff

	err = suite.repo.Update(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Verify update
	result, err := suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Real Updated Server", result.ServerName)
	assert.Equal(suite.T(), "DC2", result.Location)
	assert.Equal(suite.T(), models.ServerStatusOff, result.Status)
}

func (suite *ServerRepositoryRealTestSuite) TestRealDelete() {
	server := &models.Server{
		ServerID:   "real-delete-server",
		ServerName: "Real Delete Server",
		IPv4:       "192.168.1.104",
		Status:     models.ServerStatusOn,
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Delete the server
	err = suite.repo.Delete(suite.T().Context(), server.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	result, err := suite.repo.GetByID(suite.T().Context(), server.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *ServerRepositoryRealTestSuite) TestRealCountByStatus() {
	// Create test servers
	servers := []*models.Server{
		{ServerID: "real-count-1", ServerName: "Real Count 1", IPv4: "10.0.1.1", Status: models.ServerStatusOn},
		{ServerID: "real-count-2", ServerName: "Real Count 2", IPv4: "10.0.1.2", Status: models.ServerStatusOn},
		{ServerID: "real-count-3", ServerName: "Real Count 3", IPv4: "10.0.1.3", Status: models.ServerStatusOff},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), server)
		assert.NoError(suite.T(), err)
	}

	// Count ON servers
	count, err := suite.repo.CountByStatus(suite.T().Context(), models.ServerStatusOn)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), count, int64(2))

	// Count OFF servers
	count, err = suite.repo.CountByStatus(suite.T().Context(), models.ServerStatusOff)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), count, int64(1))
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetAll() {
	// Create some test servers
	servers := []*models.Server{
		{ServerID: "real-getall-1", ServerName: "Real GetAll 1", IPv4: "10.0.2.1", Status: models.ServerStatusOn},
		{ServerID: "real-getall-2", ServerName: "Real GetAll 2", IPv4: "10.0.2.2", Status: models.ServerStatusOff},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), server)
		assert.NoError(suite.T(), err)
	}

	// Get all servers
	result, err := suite.repo.GetAll(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(result), 2)
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetServersIP() {
	// Create test servers
	servers := []*models.Server{
		{ServerID: "real-ip-1", ServerName: "Real IP 1", IPv4: "10.0.3.1", Status: models.ServerStatusOn},
		{ServerID: "real-ip-2", ServerName: "Real IP 2", IPv4: "10.0.3.2", Status: models.ServerStatusOff},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), server)
		assert.NoError(suite.T(), err)
	}

	// Get all IPs
	ips, err := suite.repo.GetServersIP(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(ips), 2)

	// Check that our test IPs are in the result
	ipMap := make(map[string]bool)
	for _, ip := range ips {
		ipMap[ip] = true
	}
	assert.True(suite.T(), ipMap["10.0.3.1"])
	assert.True(suite.T(), ipMap["10.0.3.2"])
}

func (suite *ServerRepositoryRealTestSuite) TestRealList() {
	// Create test servers
	servers := []*models.Server{
		{
			ServerID:   "real-list-1",
			ServerName: "Real List Production 1",
			IPv4:       "10.0.4.1",
			Status:     models.ServerStatusOn,
			Location:   "US-East",
			OS:         "Ubuntu",
		},
		{
			ServerID:   "real-list-2",
			ServerName: "Real List Development 1",
			IPv4:       "10.0.4.2",
			Status:     models.ServerStatusOff,
			Location:   "US-West",
			OS:         "CentOS",
		},
	}

	for _, server := range servers {
		err := suite.repo.Create(suite.T().Context(), server)
		assert.NoError(suite.T(), err)
	}

	// Test list without filters
	filter := dto.ServerFilter{}
	pagination := dto.Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "id",
		Order:    "asc",
	}

	result, total, err := suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(result), 2)
	assert.GreaterOrEqual(suite.T(), total, int64(2))

	// Test list with status filter
	filter.Status = models.ServerStatusOn
	result, total, err = suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(result), 1)
	for _, server := range result {
		assert.Equal(suite.T(), models.ServerStatusOn, server.Status)
	}
}

func (suite *ServerRepositoryRealTestSuite) TestRealUpdateStatus() {
	server := &models.Server{
		ServerID:   "real-update-status",
		ServerName: "Real Update Status Server",
		IPv4:       "192.168.1.105",
		Status:     models.ServerStatusOn,
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Update status to OFF
	err = suite.repo.UpdateStatus(suite.T().Context(), "real-update-status", models.ServerStatusOff)
	assert.NoError(suite.T(), err)

	// Verify status update
	result, err := suite.repo.GetByServerID(suite.T().Context(), "real-update-status")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.ServerStatusOff, result.Status)
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetByServerName() {
	server := &models.Server{
		ServerID:   "real-get-by-name",
		ServerName: "Real Unique Server Name Test",
		IPv4:       "192.168.1.106",
		Status:     models.ServerStatusOn,
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Test successful retrieval
	result, err := suite.repo.GetByServerName(suite.T().Context(), "Real Unique Server Name Test")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), server.ServerID, result.ServerID)
	assert.Equal(suite.T(), server.ServerName, result.ServerName)

	// Test non-existent server name
	result, err = suite.repo.GetByServerName(suite.T().Context(), "Non-existent Server Name")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *ServerRepositoryRealTestSuite) TestRealBatchCreateWithPgxPool() {
	// Create a real server repository with mock pgx pool to test the pgx path
	mockPg := &MockPgxPool{}

	// Set up mock
	callCount := 0
	mockPg.copyFromFunc = func(ctx context.Context, tableName []string, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
		callCount++
		return 2, nil
	}

	// Create server repository with mock pgx pool
	realRepoWithPgx := &serverRepository{
		db: suite.db,
		pg: mockPg,
	}

	servers := []models.Server{
		{
			ServerID:   "pgx-batch-1",
			ServerName: "PGX Batch Server 1",
			IPv4:       "10.0.5.1",
			Status:     models.ServerStatusOn,
		},
		{
			ServerID:   "pgx-batch-2",
			ServerName: "PGX Batch Server 2",
			IPv4:       "10.0.5.2",
			Status:     models.ServerStatusOff,
		},
	}

	// Execute batch create (should use pgx path)
	err := realRepoWithPgx.BatchCreate(suite.T().Context(), servers)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, callCount) // Verify mock was called
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetByServerIDNotFound() {
	// Test error case for GetByServerID
	result, err := suite.repo.GetByServerID(suite.T().Context(), "non-existent-server-id")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *ServerRepositoryRealTestSuite) TestRealGetServersIPEmpty() {
	// Clean the database for this test
	suite.db.Exec("DELETE FROM servers")

	ips, err := suite.repo.GetServersIP(suite.T().Context())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(ips))
}

func (suite *ServerRepositoryRealTestSuite) TestRealListWithAllFilters() {
	// Create a server with specific attributes
	server := &models.Server{
		ServerID:   "filtered-server",
		ServerName: "Filtered Production Server",
		IPv4:       "10.0.6.1",
		Status:     models.ServerStatusOn,
		Location:   "US-Central",
		OS:         "Ubuntu 22.04",
	}

	err := suite.repo.Create(suite.T().Context(), server)
	assert.NoError(suite.T(), err)

	// Test filtering by all fields
	filter := dto.ServerFilter{
		ServerID:   "filtered",
		ServerName: "Production",
		IPv4:       "10.0.6",
		Status:     models.ServerStatusOn,
		Location:   "Central",
		OS:         "Ubuntu",
	}

	pagination := dto.Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "server_name",
		Order:    "desc",
	}

	result, total, err := suite.repo.List(suite.T().Context(), filter, pagination)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(result), 1)
	assert.GreaterOrEqual(suite.T(), total, int64(1))

	// Verify the result matches our filters
	found := false
	for _, s := range result {
		if s.ServerID == "filtered-server" {
			found = true
			assert.Contains(suite.T(), s.ServerName, "Production")
			assert.Equal(suite.T(), models.ServerStatusOn, s.Status)
			break
		}
	}
	assert.True(suite.T(), found)
}
func TestServerRepositoryRealTestSuite(t *testing.T) {
	suite.Run(t, new(ServerRepositoryRealTestSuite))
}
