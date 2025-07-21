package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/scope"
	"github.com/th1enq/server_management_system/internal/domain/services"
	"github.com/th1enq/server_management_system/internal/dto"
)

type jwtService struct {
	jwtConfig configs.JWT
}

func NewJWTService(cfg configs.JWT) services.TokenServices {
	return &jwtService{
		jwtConfig: cfg,
	}
}

func (j *jwtService) GenerateServerAccessToken(ctx context.Context, server *entity.Server) (string, error) {
	claims := &dto.ServerClaims{
		ServerID:   server.ServerID,
		ServerName: server.ServerName,
		TokenType:  "server_access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.jwtConfig.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   server.ServerID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.jwtConfig.Secret))
}

func (j *jwtService) GenerateServerRefreshToken(ctx context.Context, server *entity.Server) (string, error) {
	claims := &dto.ServerClaims{
		ServerID:   server.ServerID,
		ServerName: server.ServerName,
		TokenType:  "server_refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.jwtConfig.Expiration * 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   server.ServerID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.jwtConfig.Secret))
}

func (j *jwtService) GenerateAccessToken(user *entity.User) (string, error) {
	userScopes := scope.ToArray(user.HashedScopes)
	claims := &dto.Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Scopes:    userScopes,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.jwtConfig.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.jwtConfig.Secret))
}

func (j *jwtService) GenerateRefreshToken(user *entity.User) (string, error) {
	userScopes := scope.ToArray(user.HashedScopes)
	claims := &dto.Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Scopes:    userScopes,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.jwtConfig.Expiration * 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.jwtConfig.Secret))
}

func (j *jwtService) ValidateToken(tokenString string) (*dto.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*dto.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
