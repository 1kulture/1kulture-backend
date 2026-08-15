package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log request details
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		fields := map[string]interface{}{
			"request_id": c.GetString("request_id"),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"status":     c.Writer.Status(),
			"latency":    latency.String(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"errors":     c.Errors.String(),
		}

		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.WithFields(fields).Error("Server error")
		case status >= 400:
			logger.WithFields(fields).Warning("Client error")
		case status >= 300:
			logger.WithFields(fields).Info("Redirect")
		default:
			logger.WithFields(fields).Info("Request processed")
		}
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithRequest(c).Error("Panic recovered: ", err)

				// Return 500 error
				c.JSON(500, gin.H{
					"success": false,
					"message": "Internal server error",
					"error": gin.H{
						"code":    "SERVER_INTERNAL",
						"message": "An unexpected error occurred",
					},
					"timestamp":  time.Now().UTC(),
					"request_id": c.GetString("request_id"),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
