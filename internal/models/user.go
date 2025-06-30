package models

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type APIScope string

const (
	// Server management scopes
	ScopeServerRead   APIScope = "server:read"
	ScopeServerWrite  APIScope = "server:write"
	ScopeServerDelete APIScope = "server:delete"
	ScopeServerImport APIScope = "server:import"
	ScopeServerExport APIScope = "server:export"

	// User management scopes
	ScopeUserRead   APIScope = "user:read"
	ScopeUserWrite  APIScope = "user:write"
	ScopeUserDelete APIScope = "user:delete"

	// Report scopes
	ScopeReportRead  APIScope = "report:read"
	ScopeReportWrite APIScope = "report:write"

	// Profile scopes
	ScopeProfileRead  APIScope = "profile:read"
	ScopeProfileWrite APIScope = "profile:write"

	// Admin scopes
	ScopeAdminAll APIScope = "admin:all"
)

// GetDefaultScopes returns default scopes for a given role
func GetDefaultScopes(role UserRole) []APIScope {
	switch role {
	case RoleAdmin:
		return []APIScope{
			ScopeServerRead, ScopeServerWrite, ScopeServerDelete, ScopeServerImport, ScopeServerExport,
			ScopeUserRead, ScopeUserWrite, ScopeUserDelete,
			ScopeReportRead, ScopeReportWrite,
			ScopeProfileRead, ScopeProfileWrite,
			ScopeAdminAll,
		}
	case RoleUser:
		return []APIScope{
			ScopeServerRead, ScopeServerExport,
			ScopeReportRead,
			ScopeProfileRead, ScopeProfileWrite,
		}
	default:
		return []APIScope{ScopeProfileRead}
	}
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      UserRole  `gorm:"default:user" json:"role"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (u *User) SetPassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
