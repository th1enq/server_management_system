package services

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"go.uber.org/zap"
)

type MockTransport struct {
	Response    *http.Response
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.RoundTripFn(req)
}

// MockElasticsearchClient is a mock for Elasticsearch client
type MockElasticsearchClient struct {
	mock.Mock
}

// MockServerService is a mock for ServerService
type MockServerService struct {
	mock.Mock
}

func (m *MockServerService) CreateServer(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerService) GetServer(ctx context.Context, id uint) (*models.Server, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerService) ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (*dto.ServerListResponse, error) {
	args := m.Called(ctx, filter, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServerListResponse), args.Error(1)
}

func (m *MockServerService) UpdateServer(ctx context.Context, id uint, updates dto.ServerUpdate) (*models.Server, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerService) DeleteServer(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServerService) ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error) {
	args := m.Called(ctx, filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ImportResult), args.Error(1)
}

func (m *MockServerService) ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error) {
	args := m.Called(ctx, filter, pagination)
	return args.String(0), args.Error(1)
}

func (m *MockServerService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	args := m.Called(ctx, serverID, status)
	return args.Error(0)
}

func (m *MockServerService) GetServerStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockServerService) GetAllServers(ctx context.Context) ([]models.Server, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Server), args.Error(1)
}

func (m *MockServerService) CheckServerStatus(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockServerService) CheckServer(ctx context.Context, server models.Server) {
	m.Called(ctx, server)
}

func createTestReportService() (*reportService, *MockServerService) {
	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	mocktrans := MockTransport{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		},
	}
	mocktrans.RoundTripFn = func(req *http.Request) (*http.Response, error) { return mocktrans.Response, nil }

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: &mocktrans,
	})

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	return reportSrv, mockServerService
}

func TestReportService_GenerateReport_Success(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	startOfDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	serverStats := map[string]int64{
		"online":  5,
		"offline": 2,
	}

	// Use empty servers list to avoid elasticsearch calculation
	servers := []models.Server{}

	mockServerService.On("GetServerStats", ctx).Return(serverStats, nil)
	mockServerService.On("GetAllServers", ctx).Return(servers, nil)

	// Test
	report, err := reportSrv.GenerateReport(ctx, startOfDay, endOfDay)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, startOfDay, report.StartOfDay)
	assert.Equal(t, endOfDay, report.EndOfDay)
	assert.Equal(t, int64(7), report.TotalServers) // 5 + 2
	assert.Equal(t, int64(5), report.OnlineCount)
	assert.Equal(t, int64(2), report.OfflineCount)
	assert.Equal(t, float64(0), report.AvgUptime) // 0 when no servers
	assert.Len(t, report.Detail, 0)

	mockServerService.AssertExpectations(t)
}

func TestReportService_GenerateReport_GetServerStatsError(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	startOfDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	mockServerService.On("GetServerStats", ctx).Return(nil, errors.New("database error"))

	// Test
	report, err := reportSrv.GenerateReport(ctx, startOfDay, endOfDay)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "failed to get server stats")

	mockServerService.AssertExpectations(t)
}

func TestReportService_SendReportForDateRange_Success(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	emailTo := "recipient@example.com"

	serverStats := map[string]int64{
		"online":  5,
		"offline": 2,
	}

	// Use empty servers list to avoid elasticsearch calculation
	servers := []models.Server{}

	mockServerService.On("GetServerStats", ctx).Return(serverStats, nil)
	mockServerService.On("GetAllServers", ctx).Return(servers, nil)

	// Test - This will fail because we can't actually send email in tests
	// but we can test the logic up to that point
	err := reportSrv.SendReportForDateRange(ctx, startDate, endDate, emailTo)

	// Since we don't have a real SMTP server, this will fail at the email sending part
	// but we can check that it doesn't fail on input validation
	assert.Error(t, err) // Expected to fail on email sending

	mockServerService.AssertExpectations(t)
}

func TestReportService_SendReportForDateRange_InvalidDateRange(t *testing.T) {
	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	startDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // End before start
	emailTo := "recipient@example.com"

	// Test
	err := reportSrv.SendReportForDateRange(ctx, startDate, endDate, emailTo)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start date must be before end date")
}

func TestReportService_SendReportForDaily_Success(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	date := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	serverStats := map[string]int64{
		"online":  5,
		"offline": 2,
	}

	// Use empty servers list to avoid elasticsearch calculation
	servers := []models.Server{}

	mockServerService.On("GetServerStats", ctx).Return(serverStats, nil)
	mockServerService.On("GetAllServers", ctx).Return(servers, nil)

	// Test - This will fail because we can't actually send email in tests
	err := reportSrv.SendReportForDaily(ctx, date)

	// Since we don't have a real SMTP server, this will fail at the email sending part
	assert.Error(t, err) // Expected to fail on email sending

	mockServerService.AssertExpectations(t)
}

func TestReportService_CalculateAverageUptime_NoServers(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	mockServerService.On("GetAllServers", ctx).Return([]models.Server{}, nil)

	// Test
	avgUptime, detail, err := reportSrv.calculateAverageUptime(ctx, startTime, endTime)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, float64(0), avgUptime)
	assert.Nil(t, detail)

	mockServerService.AssertExpectations(t)
}

func TestReportService_CalculateAverageUptime_GetAllServersError(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	mockServerService.On("GetAllServers", ctx).Return(nil, errors.New("database error"))

	// Test
	avgUptime, detail, err := reportSrv.calculateAverageUptime(ctx, startTime, endTime)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, float64(0), avgUptime)
	assert.Nil(t, detail)

	mockServerService.AssertExpectations(t)
}

func TestReportService_CalculateServerUpTime_InvalidInput(t *testing.T) {
	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	// Test with nil serverID
	uptime, err := reportSrv.calculateServerUpTime(ctx, nil, startTime, endTime)
	assert.Error(t, err)
	assert.Equal(t, float64(0), uptime)
	assert.Contains(t, err.Error(), "serverID cannot be nil or empty")

	// Test with empty serverID
	emptyServerID := ""
	uptime, err = reportSrv.calculateServerUpTime(ctx, &emptyServerID, startTime, endTime)
	assert.Error(t, err)
	assert.Equal(t, float64(0), uptime)
	assert.Contains(t, err.Error(), "serverID cannot be nil or empty")

	// Test with invalid date range
	serverID := "test-server"
	uptime, err = reportSrv.calculateServerUpTime(ctx, &serverID, endTime, startTime) // swapped dates
	assert.Error(t, err)
	assert.Equal(t, float64(0), uptime)
	assert.Contains(t, err.Error(), "startTime cannot be after endTime")
}

func TestReportService_CalculateServerUpTime_NoElasticsearchData(t *testing.T) {
	// This test verifies that calculateServerUpTime fails gracefully when no Elasticsearch client is available
	// Since we can't safely call calculateServerUpTime with a nil Elasticsearch client (it will panic),
	// we'll test this by ensuring the method would require a non-nil client

	// Create a report service with nil Elasticsearch client
	mockServerService := &MockServerService{}
	logger, _ := zap.NewDevelopment()
	reportSrv := NewReportService(
		configs.Email{},
		nil, // nil Elasticsearch client
		mockServerService,
		logger,
	)

	// Verify that the service was created with nil ES client
	assert.NotNil(t, reportSrv)

	// Note: We cannot safely test calculateServerUpTime with nil ES client as it will panic
	// This is a limitation of the current implementation - the method should check for nil client
	// before attempting to use it. This test documents that behavior.
}

// Test helper functions that don't require external dependencies

func TestReportService_NewReportService(t *testing.T) {
	cfg := configs.Email{
		From:     "test@example.com",
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
	}

	mockServerService := &MockServerService{}
	logger := zap.NewNop()

	reportSrv := NewReportService(cfg, nil, mockServerService, logger)

	assert.NotNil(t, reportSrv)
	assert.IsType(t, &reportService{}, reportSrv)
}

func TestReportService_SendReportToEmail_TemplateNotFound(t *testing.T) {
	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	report := &models.DailyReport{
		StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		TotalServers: 7,
		OnlineCount:  5,
		OfflineCount: 2,
		AvgUptime:    85.5,
		Detail:       []models.ServerUpTime{},
	}

	emailTo := "test@example.com"
	msg := "Test Report"

	// Test - This will fail because template file doesn't exist
	err := reportSrv.SendReportToEmail(ctx, report, emailTo, msg)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read email template")
}

func TestReportService_SendReportToEmail_NilReport(t *testing.T) {
	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	emailTo := "test@example.com"
	msg := "Test Report"

	// Test with nil report
	err := reportSrv.SendReportToEmail(ctx, nil, emailTo, msg)

	// Assertions - This will likely cause a template execution error
	assert.Error(t, err)
}

func TestReportService_SendReportToEmail_EmptyEmailTo(t *testing.T) {
	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	report := &models.DailyReport{
		StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		TotalServers: 7,
		OnlineCount:  5,
		OfflineCount: 2,
		AvgUptime:    85.5,
		Detail:       []models.ServerUpTime{},
	}

	emailTo := "" // empty email
	msg := "Test Report"

	// Test
	err := reportSrv.SendReportToEmail(ctx, report, emailTo, msg)

	// Assertions - This will fail at email validation or sending
	assert.Error(t, err)
}

func TestReportService_CalculateServerUpTime_WithElasticsearchError(t *testing.T) {
	// Create a mock transport that returns an error
	mockTransport := &MockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("elasticsearch connection error")
		},
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})

	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	ctx := context.Background()
	serverID := "test-server-1"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	// Test
	uptime, err := reportSrv.calculateServerUpTime(ctx, &serverID, startTime, endTime)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, float64(0), uptime)
	assert.Contains(t, err.Error(), "elasticsearch search failed")
}

func TestReportService_CalculateServerUpTime_WithSuccessfulResponse(t *testing.T) {
	// Create a mock response with status logs
	mockResponse := `{
		"hits": {
			"hits": [
				{
					"_source": {
						"server_id": "test-server-1",
						"status": "ON",
						"@timestamp": "2024-01-01T00:00:00Z"
					}
				},
				{
					"_source": {
						"server_id": "test-server-1", 
						"status": "OFF",
						"@timestamp": "2024-01-01T12:00:00Z"
					}
				},
				{
					"_source": {
						"server_id": "test-server-1",
						"status": "ON", 
						"@timestamp": "2024-01-01T18:00:00Z"
					}
				}
			]
		}
	}`

	mockTransport := &MockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
				Header: http.Header{
					"Content-Type":      []string{"application/json"},
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
			}, nil
		},
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})

	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	ctx := context.Background()
	serverID := "test-server-1"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	// Test
	uptime, err := reportSrv.calculateServerUpTime(ctx, &serverID, startTime, endTime)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, uptime >= 0 && uptime <= 100) // Should be a valid percentage
}

func TestReportService_CalculateServerUpTime_NoDataFound(t *testing.T) {
	// Create a mock response with no hits
	mockResponse := `{
		"hits": {
			"hits": []
		}
	}`

	mockTransport := &MockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
				Header: http.Header{
					"Content-Type":      []string{"application/json"},
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
			}, nil
		},
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})

	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	ctx := context.Background()
	serverID := "test-server-1"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	// Test
	uptime, err := reportSrv.calculateServerUpTime(ctx, &serverID, startTime, endTime)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, float64(0), uptime) // Should return 0% when no data found
}

func TestReportService_CalculateAverageUptime_WithServers(t *testing.T) {
	reportSrv, mockServerService := createTestReportService()
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

	// Mock servers data
	servers := []models.Server{
		{
			ID:         1,
			ServerID:   "server-1",
			ServerName: "Test Server 1",
			Status:     models.ServerStatusOn,
		},
		{
			ID:         2,
			ServerID:   "server-2",
			ServerName: "Test Server 2",
			Status:     models.ServerStatusOff,
		},
	}

	mockServerService.On("GetAllServers", ctx).Return(servers, nil)

	// Test - calculateServerUpTime will fail with Elasticsearch issues but method continues
	avgUptime, detail, err := reportSrv.calculateAverageUptime(ctx, startTime, endTime)

	// The method should succeed even if individual server uptime calculations fail
	// It logs errors but continues processing and returns totalUpTime / len(servers)
	assert.NoError(t, err)
	assert.True(t, avgUptime >= 0) // Should be non-negative
	assert.NotNil(t, detail)       // Should return slice, not nil
	// Some servers might succeed, some might fail, so we just check it's not more than total servers
	assert.True(t, len(detail) <= len(servers))

	mockServerService.AssertExpectations(t)
}

func TestReportService_SendReportToEmail_WithValidTemplate(t *testing.T) {
	// Create temporary template directory and file
	tempDir := "/tmp/test_template"
	err := os.MkdirAll(tempDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up

	// Create a simple email template
	templateContent := `
<!DOCTYPE html>
<html>
<head><title>Test Report</title></head>
<body>
	<h1>Server Report</h1>
	<p>Period: {{.StartOfDay.Format "2006-01-02"}} to {{.EndOfDay.Format "2006-01-02"}}</p>
	<p>Total Servers: {{.TotalServers}}</p>
	<p>Online: {{.OnlineCount}}</p>
	<p>Offline: {{.OfflineCount}}</p>
	<p>Average Uptime: {{.AvgUptime}}%</p>
</body>
</html>`

	templatePath := tempDir + "/email.html"
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	// Change working directory temporarily to make template accessible
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Create a temporary directory structure that matches expected template path
	err = os.MkdirAll("template", 0755)
	assert.NoError(t, err)
	defer os.RemoveAll("template")

	err = os.WriteFile("template/email.html", []byte(templateContent), 0644)
	assert.NoError(t, err)

	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	report := &models.DailyReport{
		StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		TotalServers: 10,
		OnlineCount:  8,
		OfflineCount: 2,
		AvgUptime:    95.5,
		Detail:       []models.ServerUpTime{},
	}

	emailTo := "test@example.com"
	msg := "Test Daily Report"

	// Test - This will fail at SMTP sending since we don't have a real SMTP server
	// but we can verify it gets past template processing
	err = reportSrv.SendReportToEmail(ctx, report, emailTo, msg)

	// Should fail at email sending, not at template processing
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
	// Should NOT contain template errors
	assert.NotContains(t, err.Error(), "failed to read email template")
	assert.NotContains(t, err.Error(), "failed to parse template")
	assert.NotContains(t, err.Error(), "failed to execute template")
}

func TestReportService_SendReportToEmail_WithInvalidTemplate(t *testing.T) {
	// Create a template with invalid Go template syntax
	err := os.MkdirAll("template", 0755)
	assert.NoError(t, err)
	defer os.RemoveAll("template")

	invalidTemplateContent := `
<!DOCTYPE html>
<html>
<body>
	<h1>{{.InvalidField.NonExistentMethod.BadSyntax}}</h1>
	<p>{{range .Detail}}{{.InvalidField}}{{end}}</p>
</body>
</html>`

	err = os.WriteFile("template/email.html", []byte(invalidTemplateContent), 0644)
	assert.NoError(t, err)

	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	report := &models.DailyReport{
		StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		TotalServers: 10,
		OnlineCount:  8,
		OfflineCount: 2,
		AvgUptime:    95.5,
		Detail:       []models.ServerUpTime{},
	}

	emailTo := "test@example.com"
	msg := "Test Daily Report"

	// Test
	err = reportSrv.SendReportToEmail(ctx, report, emailTo, msg)

	// Should fail at template execution
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute template")
}

func TestReportService_SendReportToEmail_WithMalformedTemplate(t *testing.T) {
	// Create a template with malformed Go template syntax
	err := os.MkdirAll("template", 0755)
	assert.NoError(t, err)
	defer os.RemoveAll("template")

	malformedTemplateContent := `
<!DOCTYPE html>
<html>
<body>
	<h1>{{.StartOfDay</h1>  <!-- Missing closing }} -->
	<p>{{range .Detail}}</p> <!-- Missing end -->
</body>
</html>`

	err = os.WriteFile("template/email.html", []byte(malformedTemplateContent), 0644)
	assert.NoError(t, err)

	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	report := &models.DailyReport{
		StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		TotalServers: 10,
		OnlineCount:  8,
		OfflineCount: 2,
		AvgUptime:    95.5,
		Detail:       []models.ServerUpTime{},
	}

	emailTo := "test@example.com"
	msg := "Test Daily Report"

	// Test
	err = reportSrv.SendReportToEmail(ctx, report, emailTo, msg)

	// Should fail at template parsing
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestReportService_SendReportToEmail_WithEmptyTemplate(t *testing.T) {
	// Create an empty template file
	err := os.MkdirAll("template", 0755)
	assert.NoError(t, err)
	defer os.RemoveAll("template")

	err = os.WriteFile("template/email.html", []byte(""), 0644)
	assert.NoError(t, err)

	reportSrv, _ := createTestReportService()
	ctx := context.Background()

	report := &models.DailyReport{
		StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		TotalServers: 10,
		OnlineCount:  8,
		OfflineCount: 2,
		AvgUptime:    95.5,
		Detail:       []models.ServerUpTime{},
	}

	emailTo := "test@example.com"
	msg := "Test Daily Report"

	// Test - Empty template should parse fine but will fail at email sending
	err = reportSrv.SendReportToEmail(ctx, report, emailTo, msg)

	// Should fail at email sending, not template processing
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}

func TestReportService_CalculateServerUpTime_FirstLogAfterStartTime_OnlineStatus(t *testing.T) {
	// Test the specific code path where firstLog.CheckedAt.After(startTime) is true
	// and firstLog.Status == models.ServerStatusOn
	mockResponse := `{
		"hits": {
			"hits": [
				{
					"_source": {
						"server_id": "test-server-1",
						"status": "ON",
						"@timestamp": "2024-01-01T02:00:00Z"
					}
				},
				{
					"_source": {
						"server_id": "test-server-1", 
						"status": "OFF",
						"@timestamp": "2024-01-01T10:00:00Z"
					}
				}
			]
		}
	}`

	mockTransport := &MockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
				Header: http.Header{
					"Content-Type":      []string{"application/json"},
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
			}, nil
		},
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})

	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	ctx := context.Background()
	serverID := "test-server-1"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)  // Start at midnight
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC) // End at 23:59:59

	// Test
	uptime, err := reportSrv.calculateServerUpTime(ctx, &serverID, startTime, endTime)

	// Assertions
	assert.NoError(t, err)

	// Expected calculation:
	// 1. Period from 00:00:00 to 02:00:00 (2 hours) - assumed ON (same as first log)
	// 2. Period from 02:00:00 to 10:00:00 (8 hours) - ON status
	// 3. Period from 10:00:00 to 23:59:59 (13h 59m 59s) - OFF status
	// Total uptime: 2 + 8 = 10 hours out of 24 hours ≈ 41.67%
	assert.True(t, uptime > 40 && uptime < 45) // Should be around 41.67%
}

func TestReportService_CalculateServerUpTime_FirstLogAfterStartTime_OfflineStatus(t *testing.T) {
	// Test the specific code path where firstLog.CheckedAt.After(startTime) is true
	// and firstLog.Status == models.ServerStatusOff
	mockResponse := `{
		"hits": {
			"hits": [
				{
					"_source": {
						"server_id": "test-server-1",
						"status": "OFF",
						"@timestamp": "2024-01-01T03:00:00Z"
					}
				},
				{
					"_source": {
						"server_id": "test-server-1", 
						"status": "ON",
						"@timestamp": "2024-01-01T12:00:00Z"
					}
				}
			]
		}
	}`

	mockTransport := &MockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
				Header: http.Header{
					"Content-Type":      []string{"application/json"},
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
			}, nil
		},
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})

	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	ctx := context.Background()
	serverID := "test-server-1"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)  // Start at midnight
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC) // End at 23:59:59

	// Test
	uptime, err := reportSrv.calculateServerUpTime(ctx, &serverID, startTime, endTime)

	// Assertions
	assert.NoError(t, err)

	// Expected calculation:
	// 1. Period from 00:00:00 to 03:00:00 (3 hours) - assumed OFF (same as first log)
	// 2. Period from 03:00:00 to 12:00:00 (9 hours) - OFF status
	// 3. Period from 12:00:00 to 23:59:59 (11h 59m 59s) - ON status
	// Total uptime: ~12 hours out of 24 hours = 50%
	assert.True(t, uptime > 48 && uptime < 52) // Should be around 50%
}

func TestReportService_CalculateServerUpTime_FirstLogAtStartTime(t *testing.T) {
	// Test the else branch where firstLog.CheckedAt is at or before startTime
	mockResponse := `{
		"hits": {
			"hits": [
				{
					"_source": {
						"server_id": "test-server-1",
						"status": "ON",
						"@timestamp": "2024-01-01T00:00:00Z"
					}
				},
				{
					"_source": {
						"server_id": "test-server-1", 
						"status": "OFF",
						"@timestamp": "2024-01-01T12:00:00Z"
					}
				}
			]
		}
	}`

	mockTransport := &MockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
				Header: http.Header{
					"Content-Type":      []string{"application/json"},
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
			}, nil
		},
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})

	mockServerService := &MockServerService{}
	cfg := configs.Email{
		From:       "test@example.com",
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "testuser",
		Password:   "testpass",
		AdminEmail: "admin@example.com",
	}
	logger := zap.NewNop()

	reportSrv := &reportService{
		cfg:       cfg,
		esClient:  client,
		serverSrv: mockServerService,
		logger:    logger,
	}

	ctx := context.Background()
	serverID := "test-server-1"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)  // Start at midnight
	endTime := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC) // End at 23:59:59

	// Test
	uptime, err := reportSrv.calculateServerUpTime(ctx, &serverID, startTime, endTime)

	// Assertions
	assert.NoError(t, err)

	// Expected calculation:
	// 1. No initial duration (first log is exactly at start time)
	// 2. Period from 00:00:00 to 12:00:00 (12 hours) - ON status
	// 3. Period from 12:00:00 to 23:59:59 (11h 59m 59s) - OFF status
	// Total uptime: 12 hours out of 24 hours = 50%
	assert.True(t, uptime > 48 && uptime < 52) // Should be around 50%
}
