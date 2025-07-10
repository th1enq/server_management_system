package entity

import (
	"github.com/th1enq/server_management_system/internal/domain/scope"
)

type User struct {
	ID           uint
	Username     string
	Email        string
	Password     string
	Role         scope.UserRole
	FirstName    string
	LastName     string
	HashedScopes int64
	IsActive     bool
}

func (u *User) GetScopes() []scope.APIScope {
	scopes := []scope.APIScope{}

	for i, scope := range scope.AllScopes {
		if (u.HashedScopes>>i)&1 == 1 {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}
