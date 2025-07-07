package handler

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
)

type MockServerService struct {
	mock.Mock
}

func (m *MockServerService) CreateServer(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerService) GetServer(ctx context.Context, id uint) (*models.Server, error) {
	args := m.Called(ctx, id)
	if server, ok := args.Get(0).(*models.Server); ok {
		return server, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerService) ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (*dto.ServerListResponse, error) {
	args := m.Called(ctx, filter, pagination)
	if response, ok := args.Get(0).(*dto.ServerListResponse); ok {
		return response, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerService) UpdateServer(ctx context.Context, id uint, updates dto.ServerUpdate) (*models.Server, error) {
	args := m.Called(ctx, id, updates)
	if server, ok := args.Get(0).(*models.Server); ok {
		return server, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerService) DeleteServer(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerService) ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error) {
	args := m.Called(ctx, filePath)
	if result, ok := args.Get(0).(*dto.ImportResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerService) ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error) {
	args := m.Called(ctx, filter, pagination)
	if filePath, ok := args.Get(0).(string); ok {
		return filePath, args.Error(1)
	}
	return "", args.Error(1)
}

func (m *MockServerService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	args := m.Called(ctx, serverID, status)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerService) GetServerStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if stats, ok := args.Get(0).(map[string]int64); ok {
		return stats, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerService) CheckServerStatus(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerService) CheckServer(ctx context.Context, server models.Server) error {
	args := m.Called(ctx, server)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerService) GetAllServers(ctx context.Context) ([]models.Server, error) {
	args := m.Called(ctx)
	if servers, ok := args.Get(0).([]models.Server); ok {
		return servers, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerService) ClearServerCaches(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*models.User, error) {
	args := m.Called(ctx, id, updates)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*models.User, error) {
	args := m.Called(ctx, id, updates)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error {
	args := m.Called(ctx, id, updates)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	args := m.Called(ctx, limit, offset)
	if users, ok := args.Get(0).([]models.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(user *models.User) (string, error) {
	args := m.Called(user)
	if token, ok := args.Get(0).(string); ok {
		return token, args.Error(1)
	}
	return "", args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken(user *models.User) (string, error) {
	args := m.Called(user)
	if token, ok := args.Get(0).(string); ok {
		return token, args.Error(1)
	}
	return "", args.Error(1)
}

func (m *MockTokenService) ValidateToken(tokenString string) (*dto.Claims, error) {
	args := m.Called(tokenString)
	if claims, ok := args.Get(0).(*dto.Claims); ok {
		return claims, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTokenService) ParseTokenClaims(tokenString string) (*dto.Claims, error) {
	args := m.Called(tokenString)
	if claims, ok := args.Get(0).(*dto.Claims); ok {
		return claims, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTokenService) AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	args := m.Called(ctx, token, userID, expiration)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockTokenService) IsTokenWhitelisted(ctx context.Context, token string) bool {
	args := m.Called(ctx, token)
	if ok, _ := args.Get(0).(bool); ok {
		return ok
	}
	return false
}

func (m *MockTokenService) RemoveTokenFromWhitelist(ctx context.Context, token string) {
	m.Called(ctx, token)
}

func (m *MockTokenService) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) {
	m.Called(ctx, userID)
}

type MockHealthCheckService struct {
	mock.Mock
}

func (m *MockHealthCheckService) CalculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (*models.DailyReport, error) {
	args := m.Called(ctx, startTime, endTime)
	if report, ok := args.Get(0).(*models.DailyReport); ok {
		return report, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockHealthCheckService) CalculateServerUpTime(ctx context.Context, serverID *string, startTime, endTime time.Time) (float64, error) {
	args := m.Called(ctx, serverID, startTime, endTime)
	if uptime, ok := args.Get(0).(float64); ok {
		return uptime, args.Error(1)
	}
	return 0, args.Error(1)
}

func (m *MockHealthCheckService) CountLogStats(ctx context.Context, serverID *string, stat string, startTime, endTime time.Time) (int64, error) {
	args := m.Called(ctx, serverID, stat, startTime, endTime)
	if count, ok := args.Get(0).(int64); ok {
		return count, args.Error(1)
	}
	return 0, args.Error(1)
}

func (m *MockHealthCheckService) ExportReportXLSX(ctx context.Context, report *models.DailyReport) (string, error) {
	args := m.Called(ctx, report)
	if filePath, ok := args.Get(0).(string); ok {
		return filePath, args.Error(1)
	}
	return "", args.Error(1)
}
