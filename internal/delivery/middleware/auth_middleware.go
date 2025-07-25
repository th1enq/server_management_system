package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/domain/scope"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authUseCase usecases.AuthUseCase
	logger      *zap.Logger
}

func NewAuthMiddleware(authUseCase usecases.AuthUseCase, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		claims, err := m.authUseCase.ValidateToken(c, token)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Invalid token",
				nil,
			))
			c.Abort()
			return
		}
		if claims.TokenType != "access" {
			m.logger.Warn("Invalid token type", zap.String("token_type", claims.TokenType))
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Invalid token type",
				nil,
			))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", string(claims.Role))
		c.Set("scopes", claims.Scopes)
		c.Set("claims", claims)

		c.Next()
	}
}

func (m *AuthMiddleware) ServerRequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		claims, err := m.authUseCase.ValidateToken(c, token)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Invalid token",
				nil,
			))
			c.Abort()
			return
		}
		if claims.TokenType != "access" {
			m.logger.Warn("Invalid token type", zap.String("token_type", claims.TokenType))
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Invalid token type",
				nil,
			))
			c.Abort()
			return
		}

		c.Set("token", token)

		c.Next()
	}
}

// RequireRole is a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.CodeInternalServerError,
				"Invalid role data",
				nil))
			c.Abort()
			return
		}

		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(
				domain.CodeForbidden,
				"Insufficient permissions",
				nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// RequireScope is a middleware that requires a specific API scope
func (m *AuthMiddleware) RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		scopes, exists := c.Get("scopes")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Authentication required",
				nil))
			c.Abort()
			return
		}

		userScopes, ok := scopes.([]scope.APIScope)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.CodeInternalServerError,
				"Invalid scope data",
				nil))
			c.Abort()
			return
		}

		// Check if user has the required scope
		hasScope := false
		for _, scope := range userScopes {
			if string(scope) == requiredScope {
				hasScope = true
				break
			}
		}

		if !hasScope {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(
				domain.CodeForbidden,
				"Insufficient scope permissions",
				nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyScope is a middleware that requires any of the specified scopes
func (m *AuthMiddleware) RequireAnyScope(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		scopes, exists := c.Get("scopes")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
				domain.CodeUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		userScopes, ok := scopes.([]scope.APIScope)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.CodeInternalServerError,
				"Invalid scope data",
				nil,
			))
			c.Abort()
			return
		}

		// Check if user has any of the required scopes
		hasScope := false
		for _, userScope := range userScopes {
			for _, requiredScope := range requiredScopes {
				if string(userScope) == requiredScope {
					hasScope = true
					break
				}
			}
			if hasScope {
				break
			}
		}

		if !hasScope {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(
				domain.CodeForbidden,
				"Insufficient scope permissions",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractTokenFromHeader extracts the Bearer token from the Authorization header
func (m *AuthMiddleware) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if the header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	// Extract the token part
	return strings.TrimPrefix(authHeader, "Bearer ")
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}
