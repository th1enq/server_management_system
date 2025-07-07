package services

import (
	"context"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	args := m.Called(ctx, limit, offset)
	if users, ok := args.Get(0).([]models.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) ExistsByUserNameOrEmail(ctx context.Context, username, email string) (bool, error) {
	args := m.Called(ctx, username, email)
	if exists, ok := args.Get(0).(bool); ok {
		return exists, args.Error(1)
	}
	return false, args.Error(1)
}

type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) Create(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerRepository) GetByID(ctx context.Context, id uint) (*models.Server, error) {
	args := m.Called(ctx, id)
	if server, ok := args.Get(0).(*models.Server); ok {
		return server, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerRepository) GetByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	args := m.Called(ctx, serverID)
	if server, ok := args.Get(0).(*models.Server); ok {
		return server, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerRepository) GetByServerName(ctx context.Context, serverName string) (*models.Server, error) {
	args := m.Called(ctx, serverName)
	if server, ok := args.Get(0).(*models.Server); ok {
		return server, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerRepository) Update(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerRepository) List(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]models.Server, int64, error) {
	args := m.Called(ctx, filter, pagination)
	if servers, ok := args.Get(0).([]models.Server); ok {
		return servers, args.Get(1).(int64), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockServerRepository) BatchCreate(ctx context.Context, servers []models.Server) error {
	args := m.Called(ctx, servers)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerRepository) UpdateStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	args := m.Called(ctx, serverID, status)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockServerRepository) CountByStatus(ctx context.Context, status models.ServerStatus) (int64, error) {
	args := m.Called(ctx, status)
	if count, ok := args.Get(0).(int64); ok {
		return count, args.Error(1)
	}
	return 0, args.Error(1)
}

func (m *MockServerRepository) CountAll(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	if count, ok := args.Get(0).(int64); ok {
		return count, args.Error(1)
	}
	return 0, args.Error(1)
}

func (m *MockServerRepository) GetAll(ctx context.Context) ([]models.Server, error) {
	args := m.Called(ctx)
	if servers, ok := args.Get(0).([]models.Server); ok {
		return servers, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServerRepository) ExistsByServerIDOrServerName(ctx context.Context, serverID string, serverName string) (bool, error) {
	args := m.Called(ctx, serverID, serverName)
	if exists, ok := args.Get(0).(bool); ok {
		return exists, args.Error(1)
	}
	return false, args.Error(1)
}

type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	args := m.Called(ctx, key, data, ttl)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockCacheClient) Get(ctx context.Context, key string, dest any) error {
	args := m.Called(ctx, key, dest)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockCacheClient) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockCacheClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	args := m.Called(ctx, pattern)
	if keys, ok := args.Get(0).([]string); ok {
		return keys, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCacheClient) SADD(ctx context.Context, key string, members ...string) error {
	args := m.Called(ctx, key, members)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *MockCacheClient) SMEMBERS(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	if members, ok := args.Get(0).([]string); ok {
		return members, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockTransport struct {
	Response    *http.Response
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.RoundTripFn(req)
}

type MockElasticSearch struct {
	mock.Mock
}
