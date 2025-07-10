package dto

import (
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/scope"
)

type CreateUserRequest struct {
	Username  string           `json:"username" binding:"required,min=5,max=20"`
	Email     string           `json:"email" binding:"required,email"`
	Password  string           `json:"password" binding:"required,min=6"`
	FirstName string           `json:"first_name" binding:"required"`
	LastName  string           `json:"last_name" binding:"required"`
	Role      scope.UserRole   `json:"role" binding:"oneof=user admin"`
	Scopes    []scope.APIScope `json:"scopes" binding:"omitempty,valid_scope"`
}

type UserResponse struct {
	UserName  string         `json:"username"`
	Email     string         `json:"email"`
	Role      scope.UserRole `json:"role"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
}

type PasswordUpdate struct {
	OldPassword    string `json:"old_password" binding:"required,min=6,max=100"`
	RepeatPassword string `json:"repeat_password " binding:"required,min=6,max=100"`
	NewPassword    string `json:"new_password " binding:"required,min=6,max=100"`
}

type UserUpdate struct {
	Email     string           `json:"email" binding:"omitempty,email"`
	Role      scope.UserRole   `json:"role" binding:"omitempty,oneof=user admin"`
	FirstName string           `json:"first_name" binding:"omitempty"`
	LastName  string           `json:"last_name" binding:"omitempty"`
	IsActive  *bool            `json:"is_active" binding:"omitempty"`
	Scopes    []scope.APIScope `json:"scopes" binding:"omitempty,valid_scope"`
}

type ProfileUpdate struct {
	FirstName string `json:"first_name" binding:"omitempty"`
	LastName  string `json:"last_name" binding:"omitempty"`
}

type UserPagination struct {
	Page     int    `form:"page" binding:"gte=1"`
	PageSize int    `form:"page_size" binding:"gte=1,lte=100"`
	Sort     string `form:"sort"`
	Order    string `form:"order" binding:"oneof=asc desc"`
}

func FromEntityToUserResponse(user *entity.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		UserName:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func FromEntityListToUserResponseList(users []*entity.User) []UserResponse {
	if users == nil {
		return nil
	}
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = *FromEntityToUserResponse(user)
	}
	return responses
}
