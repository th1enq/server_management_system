package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
)

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

func (m *MockServerService) ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error) {
	args := m.Called(ctx, filter, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServerListResponse), args.Error(1)
}

func (m *MockServerService) UpdateServer(ctx context.Context, id uint, updates map[string]interface{}) (*models.Server, error) {
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

func (m *MockServerService) ImportServers(ctx context.Context, filePath string) (*models.ImportResult, error) {
	args := m.Called(ctx, filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImportResult), args.Error(1)
}

func (m *MockServerService) ExportServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (string, error) {
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

func createTestServerHandler() (*ServerHandler, *MockServerService) {
	mockService := new(MockServerService)
	logger := zap.NewNop()
	serverHandler := NewServerHandler(mockService, logger)
	return serverHandler, mockService
}

func setupGinTestContext(method, url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	c.Request = req
	return c, w
}

func setupGinTestContextWithParams(method, url string, body interface{}, params map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	c.Request = req

	// Set URL params
	for key, value := range params {
		c.Params = append(c.Params, gin.Param{Key: key, Value: value})
	}

	return c, w
}

func createTestRequestBody(data interface{}) io.ReadCloser {
	jsonData, _ := json.Marshal(data)
	return io.NopCloser(bytes.NewBuffer(jsonData))
}

// Tests for CreateServer
func TestServerHandler_CreateServer_Success(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	server := models.Server{
		ServerID:   "test-server-1",
		ServerName: "Test Server",
		IPv4:       "192.168.1.1",
		Status:     models.ServerStatusOff,
	}

	mockService.On("CreateServer", mock.Anything, mock.AnythingOfType("*models.Server")).Return(nil)

	c, w := setupGinTestContext("POST", "/api/v1/servers", server)
	serverHandler.CreateServer(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_CreateServer_InvalidJSON(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	c, w := setupGinTestContext("POST", "/api/v1/servers", "invalid json")
	serverHandler.CreateServer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_CreateServer_ValidationError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	server := models.Server{
		ServerID:   "",
		ServerName: "",
	}

	mockService.On("CreateServer", mock.Anything, mock.AnythingOfType("*models.Server")).Return(errors.New("server_id and server_name are required"))

	c, w := setupGinTestContext("POST", "/api/v1/servers", server)
	serverHandler.CreateServer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_CreateServer_ConflictError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	server := models.Server{
		ServerID:   "existing-server",
		ServerName: "Existing Server",
	}

	mockService.On("CreateServer", mock.Anything, mock.AnythingOfType("*models.Server")).Return(errors.New("server is already exists"))

	c, w := setupGinTestContext("POST", "/api/v1/servers", server)
	serverHandler.CreateServer(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_CreateServer_InternalError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	server := models.Server{
		ServerID:   "test-server",
		ServerName: "Test Server",
	}

	mockService.On("CreateServer", mock.Anything, mock.AnythingOfType("*models.Server")).Return(errors.New("database error"))

	c, w := setupGinTestContext("POST", "/api/v1/servers", server)
	serverHandler.CreateServer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for ListServer
func TestServerHandler_ListServer_Success(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	response := &models.ServerListResponse{
		Total: 1,
		Servers: []models.Server{
			{
				ID:         1,
				ServerID:   "test-server-1",
				ServerName: "Test Server",
			},
		},
		Page: 1,
		Size: 10,
	}

	mockService.On("ListServers", mock.Anything, mock.AnythingOfType("models.ServerFilter"), mock.AnythingOfType("models.Pagination")).Return(response, nil)

	c, w := setupGinTestContext("GET", "/api/v1/servers", nil)
	serverHandler.ListServer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_ListServer_WithFilters(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	response := &models.ServerListResponse{
		Total:   0,
		Servers: []models.Server{},
		Page:    1,
		Size:    10,
	}

	mockService.On("ListServers", mock.Anything, mock.AnythingOfType("models.ServerFilter"), mock.AnythingOfType("models.Pagination")).Return(response, nil)

	c, w := setupGinTestContext("GET", "/api/v1/servers?server_id=test&status=ON&page=2&page_size=20", nil)
	serverHandler.ListServer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_ListServer_ServiceError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	mockService.On("ListServers", mock.Anything, mock.AnythingOfType("models.ServerFilter"), mock.AnythingOfType("models.Pagination")).Return(nil, errors.New("database error"))

	c, w := setupGinTestContext("GET", "/api/v1/servers", nil)
	serverHandler.ListServer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for UpdateServer
func TestServerHandler_UpdateServer_Success(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	updatedServer := &models.Server{
		ID:         1,
		ServerID:   "test-server-1",
		ServerName: "Updated Server",
	}

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockService.On("UpdateServer", mock.Anything, uint(1), updates).Return(updatedServer, nil)

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/servers/1", updates, map[string]string{"id": "1"})
	serverHandler.UpdateServer(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_UpdateServer_InvalidID(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/servers/invalid", nil, map[string]string{"id": "invalid"})
	serverHandler.UpdateServer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_UpdateServer_InvalidJSON(t *testing.T) {
	serverHandler, _ := createTestServerHandler() // Not using mockService since no service methods should be called

	// In this case, when JSON is invalid, the handler should return early
	// and not call the service's UpdateServer method
	// We don't need to set expectations on the mock service

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/servers/1", "invalid json", map[string]string{"id": "1"})
	serverHandler.UpdateServer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Don't assert expectations since no method should be called
}

func TestServerHandler_UpdateServer_NotFound(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockService.On("UpdateServer", mock.Anything, uint(999), updates).Return(nil, errors.New("server not found"))

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/servers/999", updates, map[string]string{"id": "999"})
	serverHandler.UpdateServer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_UpdateServer_ConflictError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	updates := map[string]interface{}{
		"server_name": "Existing Server",
	}

	mockService.On("UpdateServer", mock.Anything, uint(1), updates).Return(nil, errors.New("server with name is already exists"))

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/servers/1", updates, map[string]string{"id": "1"})
	serverHandler.UpdateServer(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_UpdateServer_InternalError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	updates := map[string]interface{}{
		"server_name": "Updated Server",
	}

	mockService.On("UpdateServer", mock.Anything, uint(1), updates).Return(nil, errors.New("database error"))

	c, w := setupGinTestContextWithParams("PUT", "/api/v1/servers/1", updates, map[string]string{"id": "1"})
	serverHandler.UpdateServer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for DeleteServer
func TestServerHandler_DeleteServer_Success(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	mockService.On("DeleteServer", mock.Anything, uint(1)).Return(nil)

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/servers/1", nil, map[string]string{"id": "1"})
	serverHandler.DeleteServer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_DeleteServer_InvalidID(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/servers/invalid", nil, map[string]string{"id": "invalid"})
	serverHandler.DeleteServer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_DeleteServer_NotFound(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	mockService.On("DeleteServer", mock.Anything, uint(999)).Return(errors.New("server not found"))

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/servers/999", nil, map[string]string{"id": "999"})
	serverHandler.DeleteServer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_DeleteServer_InternalError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	mockService.On("DeleteServer", mock.Anything, uint(1)).Return(errors.New("database error"))

	c, w := setupGinTestContextWithParams("DELETE", "/api/v1/servers/1", nil, map[string]string{"id": "1"})
	serverHandler.DeleteServer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for ImportServers
func TestServerHandler_ImportServers_Success(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	// Create a test file
	testFile := createTestExcelFile()
	defer os.Remove(testFile)

	result := &models.ImportResult{
		SuccessCount:   2,
		SuccessServers: []string{"server1", "server2"},
		FailureCount:   0,
		FailureServers: []string{},
	}

	mockService.On("ImportServers", mock.Anything, mock.MatchedBy(func(filePath string) bool {
		return strings.Contains(filePath, "test.xlsx")
	})).Return(result, nil)

	// Create multipart form data
	body, contentType := createMultipartFormData("file", testFile)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("POST", "/api/v1/servers/import", body)
	req.Header.Set("Content-Type", contentType)
	c.Request = req

	serverHandler.ImportServers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_ImportServers_NoFile(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	c, w := setupGinTestContext("POST", "/api/v1/servers/import", nil)
	serverHandler.ImportServers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_ImportServers_ServiceError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	// Create a test file
	testFile := createTestExcelFile()
	defer os.Remove(testFile)

	mockService.On("ImportServers", mock.Anything, mock.MatchedBy(func(filePath string) bool {
		return strings.Contains(filePath, "test.xlsx")
	})).Return(nil, errors.New("import failed"))

	// Create multipart form data
	body, contentType := createMultipartFormData("file", testFile)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("POST", "/api/v1/servers/import", body)
	req.Header.Set("Content-Type", contentType)
	c.Request = req

	serverHandler.ImportServers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Tests for ExportServers
func TestServerHandler_ExportServers_Success(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	testFilePath := "/tmp/test_export.xlsx"

	// Create a test file for export
	file, err := os.Create(testFilePath)
	assert.NoError(t, err)
	file.Close()
	defer os.Remove(testFilePath)

	mockService.On("ExportServers", mock.Anything, mock.AnythingOfType("models.ServerFilter"), mock.AnythingOfType("models.Pagination")).Return(testFilePath, nil)

	c, w := setupGinTestContext("GET", "/api/v1/servers/export", nil)
	serverHandler.ExportServers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	mockService.AssertExpectations(t)
}

func TestServerHandler_ExportServers_WithFilters(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	testFilePath := "/tmp/test_export_filtered.xlsx"

	// Create a test file for export
	file, err := os.Create(testFilePath)
	assert.NoError(t, err)
	file.Close()
	defer os.Remove(testFilePath)

	mockService.On("ExportServers", mock.Anything, mock.AnythingOfType("models.ServerFilter"), mock.AnythingOfType("models.Pagination")).Return(testFilePath, nil)

	c, w := setupGinTestContext("GET", "/api/v1/servers/export?server_id=test&status=ON&page_size=5000", nil)
	serverHandler.ExportServers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServerHandler_ExportServers_ServiceError(t *testing.T) {
	serverHandler, mockService := createTestServerHandler()

	mockService.On("ExportServers", mock.Anything, mock.AnythingOfType("models.ServerFilter"), mock.AnythingOfType("models.Pagination")).Return("", errors.New("export failed"))

	c, w := setupGinTestContext("GET", "/api/v1/servers/export", nil)
	serverHandler.ExportServers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Helper functions
func createTestExcelFile() string {
	file, _ := os.Create("/tmp/test.xlsx")
	file.WriteString("test content") // This would be actual Excel content in real scenario
	file.Close()
	return "/tmp/test.xlsx"
}

func createMultipartFormData(fieldName, fileName string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, _ := os.Open(fileName)
	defer file.Close()

	part, _ := writer.CreateFormFile(fieldName, "test.xlsx")
	io.Copy(part, file)
	writer.Close()

	return body, writer.FormDataContentType()
}
