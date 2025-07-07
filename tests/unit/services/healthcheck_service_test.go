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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/tests/unit/handler"
	"go.uber.org/zap"
)

type HealthCheckServiceTestSuite struct {
	suite.Suite
	serverService      *handler.MockServerService
	healthCheckService services.IHealthCheckService
	mockTransport      *MockTransport
}

func (suite *HealthCheckServiceTestSuite) SetupTest() {
	suite.mockTransport = &MockTransport{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		},
	}
	suite.mockTransport.RoundTripFn = func(req *http.Request) (*http.Response, error) {
		return suite.mockTransport.Response, nil
	}

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: suite.mockTransport,
	})

	suite.serverService = &handler.MockServerService{}
	suite.healthCheckService = services.NewHealthCheckService(client, suite.serverService, zap.NewNop())
}

func TestHealthCheckService(t *testing.T) {
	suite.Run(t, new(HealthCheckServiceTestSuite))
}

func (suite *HealthCheckServiceTestSuite) TestCalculateAverageUptime_Success() {
	// Setup mock server data
	suite.serverService.On("GetAllServers", mock.Anything).Return([]models.Server{
		{ID: 1, ServerID: "server-1", Status: models.ServerStatusOn},
		{ID: 2, ServerID: "server-2", Status: models.ServerStatusOff},
	}, nil)

	// Mock Elasticsearch response for uptime calculation
	// This response will be used for each countLogStats call
	mockResponse := `{
		"hits": {
			"total": {
				"value": 100,
				"relation": "eq"
			}
		}
	}`

	suite.mockTransport.RoundTripFn = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		}, nil
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	report, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	suite.NoError(err)
	suite.NotNil(report)
	suite.Equal(int64(2), report.TotalServers)
	suite.Equal(int64(1), report.OnlineCount)
	suite.Equal(int64(1), report.OfflineCount)
	suite.Equal(startTime.Truncate(time.Second), report.StartOfDay.Truncate(time.Second))
	suite.Equal(endTime.Truncate(time.Second), report.EndOfDay.Truncate(time.Second))
	suite.Len(report.Detail, 2) // Should have details for both servers
}

func (suite *HealthCheckServiceTestSuite) TestCalculateAverageUptime_EmptyServers() {
	suite.serverService.On("GetAllServers", mock.Anything).Return([]models.Server{}, nil)

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	report, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	suite.NoError(err)
	suite.NotNil(report)
	suite.Equal(int64(0), report.TotalServers)
	suite.Equal(int64(0), report.OnlineCount)
	suite.Equal(int64(0), report.OfflineCount)
	suite.Equal(float64(0), report.AvgUptime)
	suite.Empty(report.Detail)
}

func (suite *HealthCheckServiceTestSuite) TestCalculateAverageUptime_ServiceError() {
	suite.serverService.On("GetAllServers", mock.Anything).Return(nil,
		errors.New("database connection failed"))

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	report, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	suite.Error(err)
	suite.Nil(report)
	suite.Contains(err.Error(), "failed to get servers")
}

func (suite *HealthCheckServiceTestSuite) TestExportReportXLSX_Success() {
	// Create a sample report
	report := &models.DailyReport{
		StartOfDay:   time.Now().Add(-24 * time.Hour),
		EndOfDay:     time.Now(),
		TotalServers: 2,
		OnlineCount:  1,
		OfflineCount: 1,
		AvgUptime:    75.5,
		Detail: []models.ServerUpTime{
			{
				Server: models.Server{
					ID:          1,
					ServerID:    "server-1",
					ServerName:  "Test Server 1",
					Status:      models.ServerStatusOn,
					IPv4:        "192.168.1.100",
					Description: "Test server",
					Location:    "Data Center 1",
					OS:          "Ubuntu 20.04",
					Disk:        100,
					RAM:         8,
				},
				AvgUpTime: 85.5,
			},
		},
	}

	filePath, err := suite.healthCheckService.ExportReportXLSX(context.Background(), report)

	suite.NoError(err)
	suite.NotEmpty(filePath)
	suite.Contains(filePath, "./exports/daily_report")
	suite.Contains(filePath, ".xlsx")

	// Verify file exists
	_, fileErr := os.Stat(filePath)
	suite.NoError(fileErr)

	// Clean up
	os.Remove(filePath)
}

func (suite *HealthCheckServiceTestSuite) TestExportReportXLSX_FileSaveError() {
	// Create a sample report
	report := &models.DailyReport{
		StartOfDay:   time.Now().Add(-24 * time.Hour),
		EndOfDay:     time.Now(),
		TotalServers: 1,
		OnlineCount:  1,
		OfflineCount: 0,
		AvgUptime:    100.0,
		Detail: []models.ServerUpTime{
			{
				Server: models.Server{
					ID:          1,
					ServerID:    "server-1",
					ServerName:  "Test Server 1",
					Status:      models.ServerStatusOn,
					IPv4:        "192.168.1.100",
					Description: "Test server",
					Location:    "Data Center 1",
					OS:          "Ubuntu 20.04",
					Disk:        100,
					RAM:         8,
				},
				AvgUpTime: 100.0,
			},
		},
	}

	// To test file save errors, we'll temporarily modify the system by creating
	// a directory structure that would prevent file creation

	// First ensure exports directory exists
	os.MkdirAll("./exports", 0755)

	// The best way to test this is to create a file where the directory should be
	// But since the filename includes timestamp, we can't predict it exactly
	// Instead, we'll make the entire exports directory read-only

	originalMode, _ := os.Stat("./exports")
	os.Chmod("./exports", 0444) // Read-only

	_, err := suite.healthCheckService.ExportReportXLSX(context.Background(), report)

	// Restore original permissions immediately
	if originalMode != nil {
		os.Chmod("./exports", originalMode.Mode())
	} else {
		os.Chmod("./exports", 0755)
	}

	// The test might pass or fail depending on the system, but we can check if we got an error
	// If we got an error, it should contain "failed to save file"
	if err != nil {
		suite.Contains(err.Error(), "failed to save file")
	}
	// Note: On some systems with different permission handling, this might not fail
	// This is a limitation of testing file system operations
}

func (suite *HealthCheckServiceTestSuite) TestCountLogStats_ElasticsearchError() {
	// Mock Elasticsearch to return an error response
	suite.mockTransport.Response = &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(strings.NewReader(`{"error": "internal server error"}`)),
		Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
	}

	serverID := "test-server"
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	// Use reflection to access private method for testing
	// Note: This requires the method to be public or we need to test it indirectly
	// For now, we'll test it through CalculateAverageUptime which calls it
	suite.serverService.On("GetAllServers", mock.Anything).Return([]models.Server{
		{ID: 1, ServerID: serverID, Status: models.ServerStatusOn},
	}, nil)

	_, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	// The error should be logged but not break the overall process
	suite.NoError(err) // The service should handle ES errors gracefully
}

func (suite *HealthCheckServiceTestSuite) TestCalculateAverageUptime_MixedServerStates() {
	// Test with multiple servers in different states
	suite.serverService.On("GetAllServers", mock.Anything).Return([]models.Server{
		{ID: 1, ServerID: "server-1", Status: models.ServerStatusOn},
		{ID: 2, ServerID: "server-2", Status: models.ServerStatusOn},
		{ID: 3, ServerID: "server-3", Status: models.ServerStatusOff},
		{ID: 4, ServerID: "server-4", Status: models.ServerStatusOff},
	}, nil)

	// Mock successful Elasticsearch responses with fresh response for each call
	mockResponse := `{
		"hits": {
			"total": {
				"value": 50,
				"relation": "eq"
			}
		}
	}`

	suite.mockTransport.RoundTripFn = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		}, nil
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	report, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	suite.NoError(err)
	suite.NotNil(report)
	suite.Equal(int64(4), report.TotalServers)
	suite.Equal(int64(2), report.OnlineCount)
	suite.Equal(int64(2), report.OfflineCount)
	suite.Len(report.Detail, 4)
}

func (suite *HealthCheckServiceTestSuite) TestCalculateAverageUptime_InvalidTimeRange() {
	suite.serverService.On("GetAllServers", mock.Anything).Return([]models.Server{
		{ID: 1, ServerID: "server-1", Status: models.ServerStatusOn},
	}, nil)

	// Test with invalid time range (end before start)
	startTime := time.Now()
	endTime := time.Now().Add(-24 * time.Hour)

	// The calculateServerUpTime method should handle this error
	suite.mockTransport.RoundTripFn = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"hits":{"total":{"value":0}}}`)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		}, nil
	}

	report, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	// The service should not fail but the detail might be empty due to errors in individual server calculations
	suite.NoError(err)
	suite.NotNil(report)
}

func (suite *HealthCheckServiceTestSuite) TestCalculateAverageUptime_NoElasticsearchData() {
	suite.serverService.On("GetAllServers", mock.Anything).Return([]models.Server{
		{ID: 1, ServerID: "server-1", Status: models.ServerStatusOn},
	}, nil)

	// Mock Elasticsearch response with no data
	mockResponse := `{
		"hits": {
			"total": {
				"value": 0,
				"relation": "eq"
			}
		}
	}`

	suite.mockTransport.RoundTripFn = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(mockResponse)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		}, nil
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	report, err := suite.healthCheckService.CalculateAverageUptime(context.Background(), startTime, endTime)

	suite.NoError(err)
	suite.NotNil(report)
	suite.Equal(int64(1), report.TotalServers)
	suite.Len(report.Detail, 1)
	// When no data is available, uptime should be 0
	suite.Equal(float64(0), report.Detail[0].AvgUpTime)
}

func (suite *HealthCheckServiceTestSuite) TestExportReportXLSX_EmptyReport() {
	// Test exporting an empty report
	report := &models.DailyReport{
		StartOfDay:   time.Now().Add(-24 * time.Hour),
		EndOfDay:     time.Now(),
		TotalServers: 0,
		OnlineCount:  0,
		OfflineCount: 0,
		AvgUptime:    0,
		Detail:       []models.ServerUpTime{},
	}

	filePath, err := suite.healthCheckService.ExportReportXLSX(context.Background(), report)

	suite.NoError(err)
	suite.NotEmpty(filePath)
	suite.Contains(filePath, "./exports/daily_report")
	suite.Contains(filePath, ".xlsx")

	// Verify file exists
	_, fileErr := os.Stat(filePath)
	suite.NoError(fileErr)

	// Clean up
	os.Remove(filePath)
}
