package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/1kulture/1kulture-backend/internal/utils/response"
)

// AdminMiddleware ensures the authenticated user has the "admin" role.
// Must be used after AuthMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "Access denied")
			c.Abort()
			return
		}
		roleList, ok := roles.([]string)
		if !ok {
			response.Forbidden(c, "Access denied")
			c.Abort()
			return
		}
		for _, r := range roleList {
			if r == "admin" {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "Super admin privileges required")
		c.Abort()
	}
}
