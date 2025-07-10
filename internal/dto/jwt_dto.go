package dto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/domain/scope"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type Claims struct {
	UserID    uint             `json:"user_id"`
	Username  string           `json:"username"`
	Email     string           `json:"email"`
	Role      scope.UserRole   `json:"role"`
	Scopes    []scope.APIScope `json:"scopes"`
	TokenType string           `json:"token_type"`
	jwt.RegisteredClaims
}
