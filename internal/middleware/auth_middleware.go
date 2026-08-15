package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/1kulture/1kulture-backend/internal/utils/jwt"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
)

func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required", nil)
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			logger.WithRequest(c).Error("Token validation failed: ", err)
			response.Unauthorized(c, "Invalid or expired token", nil)
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("session_id", claims.SessionID)
		c.Set("token_type", claims.TokenType)

		c.Next()
	}
}

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "Access denied")
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			response.Forbidden(c, "Invalid roles")
			c.Abort()
			return
		}

		for _, allowedRole := range allowedRoles {
			for _, userRole := range userRoles {
				if userRole == allowedRole {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

func OptionalAuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}
