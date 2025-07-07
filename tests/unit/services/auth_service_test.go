package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/tests/unit/handler"
	"go.uber.org/zap"
)

type AuthServiceTestSuite struct {
	suite.Suite
	userService  *handler.MockUserService
	tokenService *handler.MockTokenService
	authService  services.IAuthService
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.userService = &handler.MockUserService{}
	suite.tokenService = &handler.MockTokenService{}
	suite.authService = services.NewAuthService(suite.userService, suite.tokenService, zap.NewNop())
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (suite *AuthServiceTestSuite) TestLogin_Success() {
	username := "testuser"
	password := "password123"

	user := &models.User{
		ID:       1,
		Username: username,
		IsActive: true,
	}
	// Set the password so CheckPassword will work
	user.SetPassword(password)

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", mock.Anything).Return("refresh_token", nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.NoError(err)
	suite.NotNil(authResponse)
	suite.Equal("access_token", authResponse.AccessToken)
	suite.Equal("refresh_token", authResponse.RefreshToken)

	suite.userService.AssertExpectations(suite.T())
	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_UserNotFound() {
	username := "nonexistent"
	password := "password123"

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(nil,
		fmt.Errorf("user not found"))

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("invalid credentials", err.Error())

	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_InactiveUser() {
	username := "inactive_user"
	password := "password123"

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(&models.User{
		ID:       1,
		Username: username,
		IsActive: false,
	}, nil)

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("account is disabled", err.Error())

	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	username := "testuser"
	password := "wrongpassword"
	user := &models.User{
		ID:       1,
		Username: username,
		IsActive: true,
	}
	user.SetPassword("correctpassword")

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(user, nil)

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("invalid credentials", err.Error())

	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_AccessTokenGenerationError() {
	username := "testuser"
	password := "password123"
	user := &models.User{
		ID:       1,
		Username: username,
		IsActive: true,
	}
	user.SetPassword(password)

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("",
		fmt.Errorf("token generation failed"))

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to generate token", err.Error())

	suite.userService.AssertExpectations(suite.T())
	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_RefreshTokenGenerationError() {
	username := "testuser"
	password := "password123"
	user := &models.User{
		ID:       1,
		Username: username,
		IsActive: true,
	}
	user.SetPassword(password)

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", mock.Anything).Return("",
		fmt.Errorf("refresh token generation failed"))

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to generate token", err.Error())

	suite.userService.AssertExpectations(suite.T())
	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_WhitelistError() {
	username := "testuser"
	password := "password123"
	user := &models.User{
		ID:       1,
		Username: username,
		IsActive: true,
	}
	user.SetPassword(password)

	suite.userService.On("GetUserByUsername", mock.Anything, username).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", mock.Anything).Return("refresh_token", nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, "access_token", uint(1), mock.Anything).Return(
		fmt.Errorf("whitelist failed"))

	authResponse, err := suite.authService.Login(context.Background(), username, password)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to whitelist token", err.Error())

	suite.userService.AssertExpectations(suite.T())
	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_Success() {
	req := dto.CreateUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	createdUser := &models.User{
		ID:        1,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	suite.userService.On("CreateUser", mock.Anything, req).Return(createdUser, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", mock.Anything).Return("refresh_token", nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	authResponse, err := suite.authService.Register(context.Background(), req)

	suite.NoError(err)
	suite.NotNil(authResponse)
	suite.Equal("access_token", authResponse.AccessToken)
	suite.Equal("refresh_token", authResponse.RefreshToken)
	suite.Equal("Bearer", authResponse.TokenType)
	suite.Equal(createdUser, authResponse.User)

	suite.userService.AssertExpectations(suite.T())
	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_UserAlreadyExists() {
	req := dto.CreateUserRequest{
		Username: "existinguser",
		Email:    "existing@example.com",
	}

	suite.userService.On("CreateUser", mock.Anything, req).Return(nil,
		fmt.Errorf("user already exists"))

	authResponse, err := suite.authService.Register(context.Background(), req)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("user already exists", err.Error())

	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_CreateUserError() {
	req := dto.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
	}

	suite.userService.On("CreateUser", mock.Anything, req).Return(nil,
		fmt.Errorf("database error"))

	authResponse, err := suite.authService.Register(context.Background(), req)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to create user", err.Error())

	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_AccessTokenGenerationError() {
	req := dto.CreateUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	createdUser := &models.User{
		ID:        1,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	suite.userService.On("CreateUser", mock.Anything, req).Return(createdUser, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("access_token", errors.New("failed to generate token"))
	authResponse, err := suite.authService.Register(context.Background(), req)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to generate token", err.Error())
}

func (suite *AuthServiceTestSuite) TestRegister_RefreshTokenGenerationError() {
	req := dto.CreateUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	createdUser := &models.User{
		ID:        1,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	suite.userService.On("CreateUser", mock.Anything, req).Return(createdUser, nil)
	suite.tokenService.On("GenerateAccessToken", mock.Anything).Return("access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", mock.Anything).Return("refresh_token", errors.New("failed to generate token"))
	authResponse, err := suite.authService.Register(context.Background(), req)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to generate token", err.Error())
}

func (suite *AuthServiceTestSuite) TestValidateToken() {
	tokenString := "valid.token.string"
	expectedClaims := &dto.Claims{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	suite.tokenService.On("ValidateToken", tokenString).Return(expectedClaims, nil)

	claims, err := suite.authService.ValidateToken(tokenString)

	suite.NoError(err)
	suite.Equal(expectedClaims, claims)

	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestValidateToken_Invalid() {
	tokenString := "invalid.token.string"

	suite.tokenService.On("ValidateToken", tokenString).Return(nil,
		fmt.Errorf("invalid token"))

	claims, err := suite.authService.ValidateToken(tokenString)

	suite.Error(err)
	suite.Nil(claims)

	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_Success() {
	refreshToken := "valid.refresh.token"
	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}
	user := &models.User{
		ID:       1,
		Username: "testuser",
		IsActive: true,
	}

	suite.tokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.userService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", user).Return("new_refresh_token", nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.tokenService.On("RemoveTokenFromWhitelist", mock.Anything, refreshToken)

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.NoError(err)
	suite.NotNil(authResponse)
	suite.Equal("new_access_token", authResponse.AccessToken)
	suite.Equal("new_refresh_token", authResponse.RefreshToken)

	suite.tokenService.AssertExpectations(suite.T())
	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_InvalidToken() {
	refreshToken := "invalid.refresh.token"

	suite.tokenService.On("ValidateToken", refreshToken).Return(nil,
		fmt.Errorf("invalid token"))

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Contains(err.Error(), "invalid refresh token")

	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_NotRefreshType() {
	refreshToken := "valid.access.token"
	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "access", // Wrong type
	}

	suite.tokenService.On("ValidateToken", refreshToken).Return(claims, nil)

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("invalid refresh token", err.Error())

	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_UserNotFound() {
	refreshToken := "valid.refresh.token"
	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}

	suite.tokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.userService.On("GetUserByID", mock.Anything, uint(1)).Return(nil,
		fmt.Errorf("user not found"))

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("user not found", err.Error())

	suite.tokenService.AssertExpectations(suite.T())
	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_InactiveUser() {
	refreshToken := "valid.refresh.token"
	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}
	user := &models.User{
		ID:       1,
		Username: "testuser",
		IsActive: false, // User deactivated
	}

	suite.tokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.userService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("account is disabled", err.Error())

	suite.tokenService.AssertExpectations(suite.T())
	suite.userService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogout() {
	userID := uint(1)

	suite.tokenService.On("RemoveUserTokensFromWhitelist", mock.Anything, userID)

	err := suite.authService.Logout(context.Background(), userID)

	suite.NoError(err)
	suite.tokenService.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_ErrAddAccessTokenToWhitelist() {
	refreshToken := "valid.refresh.token"
	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}
	user := &models.User{
		ID:       1,
		Username: "testuser",
		IsActive: true,
	}

	suite.tokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.userService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", user).Return("new_refresh_token", nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, "new_access_token", mock.Anything, mock.Anything).Return(errors.New("failed to whitelist token"))

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to whitelist token", err.Error())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_ErrAddRefreshTokenToWhitelist() {
	refreshToken := "valid.refresh.token"
	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "refresh",
	}
	user := &models.User{
		ID:       1,
		Username: "testuser",
		IsActive: true,
	}

	suite.tokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.userService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
	suite.tokenService.On("GenerateAccessToken", user).Return("new_access_token", nil)
	suite.tokenService.On("GenerateRefreshToken", user).Return("new_refresh_token", nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, "new_access_token", mock.Anything, mock.Anything).Return(nil)
	suite.tokenService.On("AddTokenToWhitelist", mock.Anything, "new_refresh_token", mock.Anything, mock.Anything).Return(errors.New("failed to whitelist token"))

	authResponse, err := suite.authService.RefreshToken(context.Background(), refreshToken)

	suite.Error(err)
	suite.Nil(authResponse)
	suite.Equal("failed to whitelist token", err.Error())
}

func (suite *AuthServiceTestSuite) TestLogoutWithToken() {
	token := "some.token.string"

	suite.tokenService.On("RemoveTokenFromWhitelist", mock.Anything, token)

	err := suite.authService.LogoutWithToken(context.Background(), token)

	suite.NoError(err)
	suite.tokenService.AssertExpectations(suite.T())
}
