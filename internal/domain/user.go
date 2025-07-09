package domain

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      UserRole  `gorm:"default:user" json:"role"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Scopes    int64     `gorm:"default:0" json:"scopes"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// GetScopes returns the API scopes for the user
func (u *User) GetScopes() []APIScope {
	scopes := []APIScope{}

	for i, scope := range AllScopes {
		if (u.Scopes>>i)&1 == 1 {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}
