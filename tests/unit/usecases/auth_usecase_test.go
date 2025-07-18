package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/usecases"
	"github.com/th1enq/server_management_system/tests/mock"
	"go.uber.org/zap"
)

type AuthUseCaseSuiteTest struct {
	suite.Suite
	authUseCase          usecases.AuthUseCase
	mockUserUseCase      *mock.UserUseCaseMock
	mockPasswordServices *mock.PasswordServicesMock
	mockTokenServices    *mock.TokenServicesMock
	mockTokenRepository  *mock.TokenRepositoryMock
}

func (suite *AuthUseCaseSuiteTest) SetupTest() {
	suite.mockUserUseCase = new(mock.UserUseCaseMock)
	suite.mockPasswordServices = new(mock.PasswordServicesMock)
	suite.mockTokenServices = new(mock.TokenServicesMock)
	suite.mockTokenRepository = new(mock.TokenRepositoryMock)

	suite.authUseCase = usecases.NewAuthUseCase(
		suite.mockUserUseCase,
		suite.mockTokenRepository,
		suite.mockTokenServices,
		suite.mockPasswordServices,
		zap.NewNop(),
	)
}

func (suite *AuthUseCaseSuiteTest) TearDownTest() {
	suite.mockUserUseCase.AssertExpectations(suite.T())
	suite.mockPasswordServices.AssertExpectations(suite.T())
	suite.mockTokenServices.AssertExpectations(suite.T())
	suite.mockTokenRepository.AssertExpectations(suite.T())
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthUseCaseSuiteTest))
}

func (suite *AuthUseCaseSuiteTest) TestLogin_Success() {
	ctx := context.Background()
	username := "testuser"
	password := "testpass123"
	hashedPassword := "hashedpass123"

	user := &entity.User{
		ID:       1,
		Username: username,
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}

	req := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	accessToken := "access_token_123"
	refreshToken := "refresh_token_123"

	// Mock expectations
	suite.mockUserUseCase.On("GetUserByUsername", ctx, username).Return(user, nil)
	suite.mockPasswordServices.On("Verify", hashedPassword, password).Return(true, nil)
	suite.mockTokenServices.On("GenerateAccessToken", user).Return(accessToken, nil)
	suite.mockTokenServices.On("GenerateRefreshToken", user).Return(refreshToken, nil)
	suite.mockTokenRepository.On("AddTokenToWhitelist", ctx, accessToken, user.ID, time.Hour*24).Return(nil)
	suite.mockTokenRepository.On("AddTokenToWhitelist", ctx, refreshToken, user.ID, time.Hour*24*7).Return(nil)

	// Execute
	result, err := suite.authUseCase.Login(ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), accessToken, result.AccessToken)
	assert.Equal(suite.T(), refreshToken, result.RefreshToken)
	assert.Equal(suite.T(), "Bearer", result.TokenType)
}

func (suite *AuthUseCaseSuiteTest) TestLogin_UserNotFound() {
	ctx := context.Background()
	username := "nonexistent"
	password := "testpass123"

	req := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	// Mock expectations
	suite.mockUserUseCase.On("GetUserByUsername", ctx, username).Return((*entity.User)(nil), errors.New("user not found"))

	// Execute
	result, err := suite.authUseCase.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid credentials")
}

func (suite *AuthUseCaseSuiteTest) TestLogin_InactiveUser() {
	ctx := context.Background()
	username := "testuser"
	password := "testpass123"

	user := &entity.User{
		ID:       1,
		Username: username,
		Email:    "test@example.com",
		Password: "hashedpass123",
		IsActive: false,
	}

	req := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	// Mock expectations
	suite.mockUserUseCase.On("GetUserByUsername", ctx, username).Return(user, nil)

	// Execute
	result, err := suite.authUseCase.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "account is disabled")
}

func (suite *AuthUseCaseSuiteTest) TestLogin_InvalidPassword() {
	ctx := context.Background()
	username := "testuser"
	password := "wrongpass"
	hashedPassword := "hashedpass123"

	user := &entity.User{
		ID:       1,
		Username: username,
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}

	req := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	// Mock expectations
	suite.mockUserUseCase.On("GetUserByUsername", ctx, username).Return(user, nil)
	suite.mockPasswordServices.On("Verify", hashedPassword, password).Return(false, nil)

	// Execute
	result, err := suite.authUseCase.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid credentials")
}

func (suite *AuthUseCaseSuiteTest) TestLogin_PasswordVerificationError() {
	ctx := context.Background()
	username := "testuser"
	password := "testpass123"
	hashedPassword := "hashedpass123"

	user := &entity.User{
		ID:       1,
		Username: username,
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}

	req := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	// Mock expectations
	suite.mockUserUseCase.On("GetUserByUsername", ctx, username).Return(user, nil)
	suite.mockPasswordServices.On("Verify", hashedPassword, password).Return(false, errors.New("verification error"))

	// Execute
	result, err := suite.authUseCase.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid credentials")
}

func (suite *AuthUseCaseSuiteTest) TestLogin_TokenGenerationError() {
	ctx := context.Background()
	username := "testuser"
	password := "testpass123"
	hashedPassword := "hashedpass123"

	user := &entity.User{
		ID:       1,
		Username: username,
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}

	req := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	// Mock expectations
	suite.mockUserUseCase.On("GetUserByUsername", ctx, username).Return(user, nil)
	suite.mockPasswordServices.On("Verify", hashedPassword, password).Return(true, nil)
	suite.mockTokenServices.On("GenerateAccessToken", user).Return("", errors.New("token generation failed"))

	// Execute
	result, err := suite.authUseCase.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate token")
}

func (suite *AuthUseCaseSuiteTest) TestRegister_Success() {
	ctx := context.Background()
	req := dto.RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "newpass123",
		FirstName: "John",
		LastName:  "Doe",
	}

	createdUser := &entity.User{
		ID:        1,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	createUserRequest := dto.CreateUserRequest{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	accessToken := "access_token_123"
	refreshToken := "refresh_token_123"

	// Mock expectations
	suite.mockUserUseCase.On("CreateUser", ctx, createUserRequest).Return(createdUser, nil)
	suite.mockTokenServices.On("GenerateAccessToken", createdUser).Return(accessToken, nil)
	suite.mockTokenServices.On("GenerateRefreshToken", createdUser).Return(refreshToken, nil)
	suite.mockTokenRepository.On("AddTokenToWhitelist", ctx, accessToken, createdUser.ID, time.Hour*24).Return(nil)
	suite.mockTokenRepository.On("AddTokenToWhitelist", ctx, refreshToken, createdUser.ID, time.Hour*24*7).Return(nil)

	// Execute
	result, err := suite.authUseCase.Register(ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), accessToken, result.AccessToken)
	assert.Equal(suite.T(), refreshToken, result.RefreshToken)
	assert.Equal(suite.T(), "Bearer", result.TokenType)
}

func (suite *AuthUseCaseSuiteTest) TestRegister_UserAlreadyExists() {
	ctx := context.Background()
	req := dto.RegisterRequest{
		Username:  "existinguser",
		Email:     "existing@example.com",
		Password:  "newpass123",
		FirstName: "John",
		LastName:  "Doe",
	}

	createUserRequest := dto.CreateUserRequest{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Mock expectations
	suite.mockUserUseCase.On("CreateUser", ctx, createUserRequest).Return((*entity.User)(nil), errors.New("user already exists"))

	// Execute
	result, err := suite.authUseCase.Register(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "user already exists")
}

func (suite *AuthUseCaseSuiteTest) TestValidateToken_Success() {
	ctx := context.Background()
	tokenString := "valid_token_123"

	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		Email:     "test@example.com",
		TokenType: "access",
	}

	// Mock expectations
	suite.mockTokenRepository.On("IsTokenWhitelisted", ctx, tokenString).Return(true)
	suite.mockTokenServices.On("ValidateToken", tokenString).Return(claims, nil)

	// Execute
	result, err := suite.authUseCase.ValidateToken(ctx, tokenString)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), claims.UserID, result.UserID)
	assert.Equal(suite.T(), claims.Username, result.Username)
	assert.Equal(suite.T(), claims.TokenType, result.TokenType)
}

func (suite *AuthUseCaseSuiteTest) TestValidateToken_NotWhitelisted() {
	ctx := context.Background()
	tokenString := "invalid_token_123"

	// Mock expectations
	suite.mockTokenRepository.On("IsTokenWhitelisted", ctx, tokenString).Return(false)

	// Execute
	result, err := suite.authUseCase.ValidateToken(ctx, tokenString)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token is not whitelisted")
}

func (suite *AuthUseCaseSuiteTest) TestRefreshToken_Success() {
	ctx := context.Background()
	refreshToken := "refresh_token_123"

	req := dto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		Email:     "test@example.com",
		TokenType: "refresh",
	}

	user := &entity.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}

	newAccessToken := "new_access_token_123"
	newRefreshToken := "new_refresh_token_123"

	// Mock expectations
	suite.mockTokenRepository.On("IsTokenWhitelisted", ctx, refreshToken).Return(true)
	suite.mockTokenServices.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.mockUserUseCase.On("GetUserByID", ctx, claims.UserID).Return(user, nil)
	suite.mockTokenServices.On("GenerateAccessToken", user).Return(newAccessToken, nil)
	suite.mockTokenServices.On("GenerateRefreshToken", user).Return(newRefreshToken, nil)
	suite.mockTokenRepository.On("AddTokenToWhitelist", ctx, newAccessToken, user.ID, time.Hour*24).Return(nil)
	suite.mockTokenRepository.On("AddTokenToWhitelist", ctx, newRefreshToken, user.ID, time.Hour*24*7).Return(nil)
	suite.mockTokenRepository.On("RemoveTokenFromWhitelist", ctx, refreshToken).Return(nil)

	// Execute
	result, err := suite.authUseCase.RefreshToken(ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), newAccessToken, result.AccessToken)
	assert.Equal(suite.T(), newRefreshToken, result.RefreshToken)
	assert.Equal(suite.T(), "Bearer", result.TokenType)
}

func (suite *AuthUseCaseSuiteTest) TestRefreshToken_InvalidToken() {
	ctx := context.Background()
	refreshToken := "invalid_refresh_token"

	req := dto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	// Mock expectations
	suite.mockTokenRepository.On("IsTokenWhitelisted", ctx, refreshToken).Return(false)

	// Execute
	result, err := suite.authUseCase.RefreshToken(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
}

func (suite *AuthUseCaseSuiteTest) TestRefreshToken_NotRefreshToken() {
	ctx := context.Background()
	refreshToken := "access_token_123"

	req := dto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		Email:     "test@example.com",
		TokenType: "access", // Not a refresh token
	}

	// Mock expectations
	suite.mockTokenRepository.On("IsTokenWhitelisted", ctx, refreshToken).Return(true)
	suite.mockTokenServices.On("ValidateToken", refreshToken).Return(claims, nil)

	// Execute
	result, err := suite.authUseCase.RefreshToken(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
}

func (suite *AuthUseCaseSuiteTest) TestRefreshToken_InactiveUser() {
	ctx := context.Background()
	refreshToken := "refresh_token_123"

	req := dto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	claims := &dto.Claims{
		UserID:    1,
		Username:  "testuser",
		Email:     "test@example.com",
		TokenType: "refresh",
	}

	user := &entity.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: false, // Inactive user
	}

	// Mock expectations
	suite.mockTokenRepository.On("IsTokenWhitelisted", ctx, refreshToken).Return(true)
	suite.mockTokenServices.On("ValidateToken", refreshToken).Return(claims, nil)
	suite.mockUserUseCase.On("GetUserByID", ctx, claims.UserID).Return(user, nil)

	// Execute
	result, err := suite.authUseCase.RefreshToken(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "account is disabled")
}

func (suite *AuthUseCaseSuiteTest) TestLogout_Success() {
	ctx := context.Background()
	userID := uint(1)

	// Mock expectations
	suite.mockTokenRepository.On("RemoveUserTokensFromWhitelist", ctx, userID).Return(nil)

	// Execute
	err := suite.authUseCase.Logout(ctx, userID)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *AuthUseCaseSuiteTest) TestLogout_Error() {
	ctx := context.Background()
	userID := uint(1)

	// Mock expectations
	suite.mockTokenRepository.On("RemoveUserTokensFromWhitelist", ctx, userID).Return(errors.New("failed to remove tokens"))

	// Execute
	err := suite.authUseCase.Logout(ctx, userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to logout user")
}
