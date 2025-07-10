package models

import (
	"time"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/scope"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey"`
	Username     string         `gorm:"uniqueIndex;not null"`
	Email        string         `gorm:"uniqueIndex;not null"`
	Password     string         `gorm:"not null"`
	Role         scope.UserRole `gorm:"not null;default:'USER'"`
	FirstName    string
	LastName     string
	HashedScopes int64     `gorm:"not null;default:0"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt
}

func FromUserEntity(u *entity.User) *User {
	return &User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		Password:     u.Password,
		Role:         u.Role,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		HashedScopes: u.HashedScopes,
		IsActive:     u.IsActive,
	}
}

func ToUserEntity(u *User) *entity.User {
	return &entity.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		Password:     u.Password,
		Role:         u.Role,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		HashedScopes: u.HashedScopes,
		IsActive:     u.IsActive,
	}
}

func ToUserEntities(users []User) []*entity.User {
	var entities []*entity.User
	for _, u := range users {
		entities = append(entities, ToUserEntity(&u))
	}
	return entities
}
