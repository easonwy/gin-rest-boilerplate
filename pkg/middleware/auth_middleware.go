package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tapas/go-user-service/internal/config"
	"github.com/tapas/go-user-service/pkg/response"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
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
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			// TODO: Handle specific JWT errors (e.g., expired, invalid signature)
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 提取 claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 将用户 ID 存储到 Gin 上下文
			userID, ok := claims["user_id"].(string)
			if !ok {
				response.InternalServerError(c, "User ID not found in token claims")
				c.Abort()
				return
			}
			c.Set("userID", userID)

			// 继续处理请求
			c.Next()
		} else {
			response.Unauthorized(c, "Invalid token claims")
			c.Abort()
		}
	}
}
