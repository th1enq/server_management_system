package repository

import (
	"context"

	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
)

type IUserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]models.User, error)
	ExistsByUserNameOrEmail(ctx context.Context, username, email string) (bool, error)
}

type userRepository struct {
	db db.IDatabaseClient
}

func NewUserRepository(db db.IDatabaseClient) IUserRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Create(ctx context.Context, user *models.User) error {
	return u.db.WithContext(ctx).Create(user)
}

func (u *userRepository) Delete(ctx context.Context, id uint) error {
	return u.db.WithContext(ctx).Delete(&models.User{}, id)
}

func (u *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).First(&user, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := u.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *userRepository) Update(ctx context.Context, user *models.User) error {
	return u.db.WithContext(ctx).Save(user)
}

func (u *userRepository) ExistsByUserNameOrEmail(ctx context.Context, username, email string) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
