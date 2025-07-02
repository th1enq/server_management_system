package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) GenerateReport(ctx context.Context, startOfDay, endOfDay time.Time) (*models.DailyReport, error) {
	args := m.Called(ctx, startOfDay, endOfDay)
	return args.Get(0).(*models.DailyReport), args.Error(1)
}

func (m *MockReportService) SendReportToEmail(ctx context.Context, report *models.DailyReport, emailTo, msg string) error {
	args := m.Called(ctx, report, emailTo, msg)
	return args.Error(0)
}

func (m *MockReportService) SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error {
	args := m.Called(ctx, startDate, endDate, emailTo)
	return args.Error(0)
}

func (m *MockReportService) SendReportForDaily(ctx context.Context, date time.Time) error {
	args := m.Called(ctx, date)
	return args.Error(0)
}

func createTestReportHandler() (*ReportHandler, *MockReportService) {
	mockService := new(MockReportService)
	logger := zap.NewNop()

	reportHander := NewReportHandler(mockService, logger)

	return reportHander, mockService
}

func TestReportHandler_SendReportDaily_Success(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Mock the service method
	mockService.On("SendReportForDaily", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	// Setup test context
	c, _ := setupGinTestContext("POST", "/api/v1/reports/daily", nil)

	// Call the handler
	reportHandler.SendReportDaily(c)

	// Assert response
	if c.Writer.Status() != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, c.Writer.Status())
	}

	mockService.AssertExpectations(t)
}

func TestReportHandler_SendReportDaily_Failure(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Mock the service method to return an error
	mockService.On("SendReportForDaily", mock.Anything, mock.AnythingOfType("time.Time")).Return(errors.New("some error"))

	// Setup test context
	c, _ := setupGinTestContext("POST", "/api/v1/reports/daily", nil)

	// Call the handler
	reportHandler.SendReportDaily(c)

	// Assert response
	if c.Writer.Status() != http.StatusInternalServerError {
		t.Errorf("Expected status %d but got %d", http.StatusInternalServerError, c.Writer.Status())
	}

	mockService.AssertExpectations(t)
}

func TestReportHandler_SendReportByDate_Success(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Mock the service method
	mockService.On("SendReportForDateRange", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Return(nil)

	// Setup test context with proper request struct
	requestBody := ReportRequest{
		StartDate: "2023-01-01 00:00:00",
		EndDate:   "2023-01-31 23:59:59",
		Email:     "test@example.com",
	}
	c, _ := setupGinTestContext("POST", "/api/v1/reports", requestBody)

	// Call the handler
	reportHandler.SendReportByDate(c)
	if c.Writer.Status() != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, c.Writer.Status())
	}
	mockService.AssertExpectations(t)
}

func TestReportHandler_SendReportByDate_InvalidJSON(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Setup test context with invalid email
	requestBody := ReportRequest{
		StartDate: "2023-01-01 00:00:00",
		EndDate:   "2023-01-31 23:59:59",
		Email:     "invalid-email", // Invalid email format
	}
	c, _ := setupGinTestContext("POST", "/api/v1/reports", requestBody)

	// Call the handler
	reportHandler.SendReportByDate(c)

	// Assert response - should fail due to email validation
	if c.Writer.Status() != http.StatusBadRequest {
		t.Errorf("Expected status %d but got %d", http.StatusBadRequest, c.Writer.Status())
	}

	mockService.AssertExpectations(t)
}

func TestReportHandler_SendReportByDate_InvalidStartDateFormat(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Setup test context with invalid date format
	requestBody := ReportRequest{
		StartDate: "2023-01-01", // Invalid date format
		EndDate:   "2023-01-31",
		Email:     "test@example.com",
	}
	c, _ := setupGinTestContext("POST", "/api/v1/reports", requestBody)

	// Call the handler
	reportHandler.SendReportByDate(c)

	// Assert response - should fail due to date format
	if c.Writer.Status() != http.StatusBadRequest {
		t.Errorf("Expected status %d but got %d", http.StatusBadRequest, c.Writer.Status())
	}

	mockService.AssertExpectations(t)
}

func TestReportHandler_SendReportByDate_InvalidEndDateFormat(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Setup test context with invalid date format
	requestBody := ReportRequest{
		StartDate: "2023-01-01 00:00:00",
		EndDate:   "2023-01-01",
		Email:     "test@example.com",
	}
	c, _ := setupGinTestContext("POST", "/api/v1/reports", requestBody)

	// Call the handler
	reportHandler.SendReportByDate(c)

	// Assert response - should fail due to date format
	if c.Writer.Status() != http.StatusBadRequest {
		t.Errorf("Expected status %d but got %d", http.StatusBadRequest, c.Writer.Status())
	}

	mockService.AssertExpectations(t)
}

func TestReportHandler_SendReportByDate_ServiceError(t *testing.T) {
	reportHandler, mockService := createTestReportHandler()

	// Mock the service method to return an error
	mockService.On("SendReportForDateRange", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Return(errors.New("service error"))

	// Setup test context with proper request struct
	requestBody := ReportRequest{
		StartDate: "2023-01-01 00:00:00",
		EndDate:   "2023-01-31 23:59:59",
		Email:     "test@example.com",
	}
	c, _ := setupGinTestContext("POST", "/api/v1/reports", requestBody)

	// Call the handler
	reportHandler.SendReportByDate(c)

	// Assert response - should fail due to service error
	if c.Writer.Status() != http.StatusInternalServerError {
		t.Errorf("Expected status %d but got %d", http.StatusInternalServerError, c.Writer.Status())
	}
	mockService.AssertExpectations(t)
}
