package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    -1,
					"message": "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
