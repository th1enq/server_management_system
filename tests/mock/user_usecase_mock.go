package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/dto"
)

type UserUseCaseMock struct {
	mock.Mock
}

func (u *UserUseCaseMock) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*entity.User, error) {
	args := u.Called(ctx, req)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (u *UserUseCaseMock) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	args := u.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (u *UserUseCaseMock) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := u.Called(ctx, username)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (u *UserUseCaseMock) UpdateUser(ctx context.Context, id uint, updates dto.UserUpdate) (*entity.User, error) {
	args := u.Called(ctx, id, updates)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (u *UserUseCaseMock) UpdateProfile(ctx context.Context, id uint, updates dto.ProfileUpdate) (*entity.User, error) {
	args := u.Called(ctx, id, updates)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (u *UserUseCaseMock) UpdatePassword(ctx context.Context, id uint, updates dto.PasswordUpdate) error {
	args := u.Called(ctx, id, updates)
	return args.Error(0)
}

func (u *UserUseCaseMock) DeleteUser(ctx context.Context, id uint) error {
	args := u.Called(ctx, id)
	return args.Error(0)
}

func (u *UserUseCaseMock) ListUsers(ctx context.Context, pagination dto.UserPagination) ([]*entity.User, error) {
	args := u.Called(ctx, pagination)
	return args.Get(0).([]*entity.User), args.Error(1)
}
