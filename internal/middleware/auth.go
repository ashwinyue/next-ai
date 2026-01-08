package middleware

import (
	"strings"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware 认证中间件
// 如果提供了有效的 JWT token，则使用该用户；否则生成临时用户ID
func AuthMiddleware(svc *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 Authorization 获取 Bearer Token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			user, err := svc.Auth.ValidateToken(c.Request.Context(), token)
			if err == nil {
				// Token 有效，设置用户到上下文
				c.Set("user", user)
				c.Set("user_id", user.ID)
				// 设置租户 ID 到上下文
				if user.TenantID != "" {
					c.Set("tenant_id", user.TenantID)
				}
				c.Next()
				return
			}
			// Token 无效，继续尝试其他方式
		}

		// 从 Header 获取用户ID（兼容旧版）
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			// 生成临时用户ID
			userID = uuid.New().String()
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// RequireAuth 要求有效认证的中间件
// 必须提供有效的 JWT token，否则返回 401
func RequireAuth(svc *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"code":    -1,
				"message": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{
				"code":    -1,
				"message": "Invalid Authorization header format",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := svc.Auth.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    -1,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Token 有效，设置用户到上下文
		c.Set("user", user)
		c.Set("user_id", user.ID)
		// 设置租户 ID 到上下文
		if user.TenantID != "" {
			c.Set("tenant_id", user.TenantID)
		}
		c.Next()
	}
}

// GetCurrentUser 从上下文获取当前用户
func GetCurrentUser(c *gin.Context) (*model.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	u, ok := user.(*model.User)
	return u, ok
}

// GetUserID 从上下文获取当前用户ID
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// GetTenantID 从上下文获取当前租户ID
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return ""
}
