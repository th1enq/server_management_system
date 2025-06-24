package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/models"
)

func JWTAuth(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.CodeUnauthorized,
				"Authorization header is required",
				nil,
			))
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, "")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.CodeUnauthorized,
				"Invalid Authorization header format",
				nil,
			))
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			code := models.CodeInvalidToken
			message := "Invalid token"

			if err == jwt.ErrTokenExpired {
				code = models.CodeTokenExpired
				message = "Token has expired"
			}

			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				code,
				message,
				nil,
			))
			c.Abort()
			return
		}
	}
}
