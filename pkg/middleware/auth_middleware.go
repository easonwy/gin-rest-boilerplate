package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tapas/go-user-service/internal/auth/repository"
	"github.com/tapas/go-user-service/internal/config"
	"github.com/tapas/go-user-service/pkg/response"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(cfg *config.Config, authRepo repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// 检查 Authorization 格式是否为 "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid Authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析和验证 JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})

		// Handle parsing errors
		if err != nil {
			response.Unauthorized(c, "Invalid token: "+err.Error())
			c.Abort()
			return
		}

		// 提取 claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user ID from claims
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				response.InternalServerError(c, "User ID not found in token claims or invalid type")
				c.Abort()
				return
			}
			userID := uint(userIDFloat)

			// Check if the refresh token associated with this user ID exists in Redis
			// This acts as a simple token revocation mechanism (if refresh token is deleted, access token is also invalid)
			_, err := authRepo.GetUserRefreshToken(context.Background(), userID)
			if err != nil {
				if errors.Is(err, redis.Nil) {
					response.Unauthorized(c, "Token revoked or expired")
					c.Abort()
					return
				}
				response.InternalServerError(c, "Failed to check token validity")
				c.Abort()
				return
			}

			// Store user ID in Gin context
			c.Set("userID", userID)

			// Continue to next handler
			c.Next()
		} else {
			response.Unauthorized(c, "Invalid token claims")
			c.Abort()
		}
	}
}
