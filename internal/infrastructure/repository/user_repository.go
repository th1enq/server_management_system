package repository

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/query"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/models"
)

type userRepository struct {
	db database.DatabaseClient
}

func NewUserRepository(db database.DatabaseClient) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Create(ctx context.Context, user *entity.User) error {
	model := models.FromUserEntity(user)
	return u.db.WithContext(ctx).CreateWithErr(model)
}

func (u *userRepository) Delete(ctx context.Context, id uint) error {
	return u.db.WithContext(ctx).Delete(&models.User{}, id)
}

func (u *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if err != nil {
		return nil, err
	}
	return models.ToUserEntity(&user), nil
}

func (u *userRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).First(&user, id)
	if err != nil {
		return nil, err
	}
	return models.ToUserEntity(&user), nil
}

func (u *userRepository) List(ctx context.Context, pagination query.Pagination) ([]*entity.User, error) {
	var users []models.User

	query := u.db.WithContext(ctx).Model(&models.User{})

	offset := pagination.Offset()
	orderBy := fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)

	err := query.
		Order(orderBy).
		Limit(pagination.PageSize).
		Offset(offset).
		Find(&users)

	if err != nil {
		return nil, err
	}

	return models.ToUserEntities(users), nil
}

func (u *userRepository) Update(ctx context.Context, user *entity.User) error {
	model := models.FromUserEntity(user)
	return u.db.WithContext(ctx).Save(model)
}

func (u *userRepository) ExistsByUserNameOrEmail(ctx context.Context, username, email string, id uint) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&models.User{}).
		Where("(username = ? OR email = ?) AND id != ?", username, email, id).Count(&count)
	if err != nil {
		return false, err

	}
	return count > 0, nil
}
