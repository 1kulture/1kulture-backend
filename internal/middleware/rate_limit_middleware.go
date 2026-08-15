package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
)

func RateLimitMiddleware(redisClient *redis.Client, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Redis is not available, skip rate limiting
		if redisClient == nil {
			c.Next()
			return
		}

		ctx := context.Background()

		// Use IP as the key
		key := fmt.Sprintf("rate_limit:%s:%s", c.ClientIP(), c.Request.URL.Path)

		// Get current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// Redis error, allow request but log it
			logger.Warning("Rate limit check failed: ", err)
			c.Next()
			return
		}

		if count >= cfg.Requests {
			response.TooManyRequests(c, "Too many requests, please try again later")
			c.Abort()
			return
		}

		// Increment count
		pipe := redisClient.TxPipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, cfg.Duration)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// Redis error, allow request but log it
			logger.Warning("Rate limit increment failed: ", err)
			c.Next()
			return
		}

		c.Next()
	}
}
