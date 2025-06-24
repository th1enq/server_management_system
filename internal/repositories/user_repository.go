package repositories

import (
	"context"

	"github.com/th1enq/server_management_system/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
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
	panic("unimplemented")
}

// Delete implements UserRepository.
func (u *userRepository) Delete(ctx context.Context, id uint) error {
	panic("unimplemented")
}

// GetByID implements UserRepository.
func (u *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	panic("unimplemented")
}

func (u *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	panic("unimplemented")
}

// Update implements UserRepository.
func (u *userRepository) Update(ctx context.Context, user *models.User) error {
	panic("unimplemented")
}
