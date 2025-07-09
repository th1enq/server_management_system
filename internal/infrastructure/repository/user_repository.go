package repository

import (
	"context"

	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]domain.User, error)
	ExistsByUserNameOrEmail(ctx context.Context, username, email string) (bool, error)
}

type userRepository struct {
	db database.DatabaseClient
}

func NewUserRepository(db database.DatabaseClient) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Create(ctx context.Context, user *domain.User) error {
	return u.db.WithContext(ctx).Create(user)
}

func (u *userRepository) Delete(ctx context.Context, id uint) error {
	return u.db.WithContext(ctx).Delete(&domain.User{}, id)
}

func (u *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := u.db.WithContext(ctx).First(&user, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	err := u.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *userRepository) Update(ctx context.Context, user *domain.User) error {
	return u.db.WithContext(ctx).Save(user)
}

func (u *userRepository) ExistsByUserNameOrEmail(ctx context.Context, username, email string) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&domain.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
