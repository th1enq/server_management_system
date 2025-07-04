package dto

import (
	"github.com/th1enq/server_management_system/internal/models"
)

type PasswordUpdate struct {
	OldPassword    string `json:"old_password" binding:"required,min=6,max=100"`
	RepeatPassword string `json:"repeat_password " binding:"required,min=6,max=100"`
	NewPassword    string `json:"new_password " binding:"required,min=6,max=100"`
}

type UserUpdate struct {
	Username  string            `json:"username" binding:"min=3,max=20"`
	Email     string            `json:"email" binding:"email"`
	Password  string            `json:"password" binding:"omitempty,min=6,max=100"`
	Role      models.UserRole   `json:"role" binding:"oneof=user admin"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	IsActive  bool              `json:"is_active" binding:"omitempty"`
	Scopes    []models.APIScope `json:"scopes" binding:"omitempty"`
}

type ProfileUpdate struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type CreateUserRequest struct {
	Username  string            `json:"username" binding:"required,min=3,max=20"`
	Email     string            `json:"email" binding:"required,email"`
	Password  string            `json:"password" binding:"required,min=6"`
	FirstName string            `json:"first_name" binding:"required"`
	LastName  string            `json:"last_name" binding:"required"`
	Role      models.UserRole   `json:"role" binding:"oneof=user admin" default:"user"`
	Scopes    []models.APIScope `json:"scopes" binding:"omitempty"`
}
