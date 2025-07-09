package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/infrastructure/repository"
	"go.uber.org/zap"
)

type TokenUseCase interface {
	GenerateAccessToken(user *domain.User) (string, error)
	GenerateRefreshToken(user *domain.User) (string, error)
	ValidateToken(tokenString string) (*dto.Claims, error)
	AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error
	IsTokenWhitelisted(ctx context.Context, token string) bool
	RemoveTokenFromWhitelist(ctx context.Context, token string) error
	RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) error
}

type tokenUseCase struct {
	jwtConfig configs.JWT
	logger    *zap.Logger
	tokenRepo repository.TokenRepository
}

func NewTokenUseCase(jwtConfig configs.JWT, logger *zap.Logger, tokenRepo repository.TokenRepository) TokenUseCase {
	return &tokenUseCase{
		jwtConfig: jwtConfig,
		logger:    logger,
		tokenRepo: tokenRepo,
	}
}

func (t *tokenUseCase) GenerateAccessToken(user *domain.User) (string, error) {
	userScopes := domain.ToArray(user.Scopes)
	claims := &dto.Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Scopes:    userScopes,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.jwtConfig.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(t.jwtConfig.Secret))
}

func (t *tokenUseCase) GenerateRefreshToken(user *domain.User) (string, error) {
	userScopes := domain.ToArray(user.Scopes)
	claims := &dto.Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Scopes:    userScopes,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.jwtConfig.Expiration * 7)), // 7x longer than access token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "server_management_system",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.jwtConfig.Secret))
}

func (t *tokenUseCase) ValidateToken(tokenString string) (*dto.Claims, error) {
	if !t.tokenRepo.IsTokenWhitelisted(context.Background(), tokenString) {
		return nil, fmt.Errorf("token is not valid or has been revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &dto.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*dto.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (t *tokenUseCase) AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	if err := t.tokenRepo.AddTokenToWhitelist(ctx, token, userID, expiration); err != nil {
		t.logger.Error("Failed to add token to whitelist", zap.Error(err))
		return fmt.Errorf("failed to add token to whitelist: %w", err)
	}
	return nil
}

func (t *tokenUseCase) IsTokenWhitelisted(ctx context.Context, token string) bool {
	return t.tokenRepo.IsTokenWhitelisted(ctx, token)
}

func (t *tokenUseCase) RemoveTokenFromWhitelist(ctx context.Context, token string) error {
	if err := t.tokenRepo.RemoveTokenFromWhitelist(ctx, token); err != nil {
		t.logger.Error("Failed to remove token from whitelist", zap.Error(err))
		return fmt.Errorf("failed to remove token from whitelist: %w", err)
	}
	return nil
}

func (t *tokenUseCase) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) error {
	if err := t.tokenRepo.RemoveUserTokensFromWhitelist(ctx, userID); err != nil {
		t.logger.Error("Failed to remove user tokens from whitelist", zap.Uint("user_id", userID), zap.Error(err))
		return fmt.Errorf("failed to remove user tokens from whitelist: %w", err)
	}
	return nil
}
