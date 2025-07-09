package dto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/domain"
)

type AuthResponse struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	TokenType    string            `json:"token_type"`
	ExpiresIn    int64             `json:"expires_in"`
	User         *domain.User      `json:"user"`
	Scopes       []domain.APIScope `json:"scopes"`
}

type Claims struct {
	UserID    uint              `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Role      domain.UserRole   `json:"role"`
	Scopes    []domain.APIScope `json:"scopes"`
	TokenType string            `json:"token_type"`
	jwt.RegisteredClaims
}
