package repositories

import (
	"context"
	"time"

	"github.com/th1enq/server_management_system/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	UpdateLastLogin(ctx context.Context, userID uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create implements UserRepository.
func (u *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := user.SetPassword(user.Password); err != nil {
		return err
	}
	return u.db.WithContext(ctx).Create(user).Error
}

// Delete implements UserRepository.
func (u *userRepository) Delete(ctx context.Context, id uint) error {
	return u.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// GetByUsername implements UserRepository.
func (u *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID implements UserRepository.
func (u *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail implements UserRepository.
func (u *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List implements UserRepository.
func (u *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := u.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Update implements UserRepository.
func (u *userRepository) Update(ctx context.Context, user *models.User) error {
	return u.db.WithContext(ctx).Save(user).Error
}

// UpdateLastLogin implements UserRepository.
func (u *userRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	return u.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login", time.Now()).Error
}
