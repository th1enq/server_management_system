package dto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/models"
)

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *models.User `json:"user"`
	Scopes       []string     `json:"scopes"`
}

type Claims struct {
	UserID    uint              `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Role      models.UserRole   `json:"role"`
	Scopes    []models.APIScope `json:"scopes"`
	TokenType string            `json:"token_type"`
	jwt.RegisteredClaims
}
