package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware 认证中间件（简化版，实际应使用 JWT）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取用户ID
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			// 生成临时用户ID
			userID = uuid.New().String()
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
