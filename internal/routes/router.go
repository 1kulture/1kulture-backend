package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/controllers"
	"github.com/1kulture/1kulture-backend/internal/middleware"
	"github.com/1kulture/1kulture-backend/internal/repositories"
	"github.com/1kulture/1kulture-backend/internal/services"
	"github.com/1kulture/1kulture-backend/internal/utils/email"
	"github.com/1kulture/1kulture-backend/internal/utils/jwt"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, jwtManager *jwt.JWTManager) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.Security.AllowedOrigins))
	router.Use(middleware.RequestIDMiddleware())

	// Set trusted proxies
	router.SetTrustedProxies(cfg.Security.TrustedProxies)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	emailVerifRepo := repositories.NewEmailVerificationRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)
	kycRepo := repositories.NewKYCRepository(db)
	waitlistRepo := repositories.NewWaitlistRepository(db)

	// Initialize services
	emailService := email.NewEmailService(&cfg.Email)
	authService := services.NewAuthService(
		userRepo,
		roleRepo,
		passwordResetRepo,
		refreshTokenRepo,
		emailVerifRepo,
		auditLogRepo,
		jwtManager,
		emailService,
		cfg,
	)
	userService := services.NewUserService(userRepo, roleRepo, auditLogRepo, kycRepo)
	waitlistService := services.NewWaitlistService(waitlistRepo, auditLogRepo)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	waitlistController := controllers.NewWaitlistController(waitlistService)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/signup", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.SignUp)
			authRoutes.POST("/signin", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.SignIn)
			authRoutes.POST("/verify-email", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.VerifyEmail)
			authRoutes.POST("/resend-verification", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.ResendVerification)
			authRoutes.POST("/refresh-token", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.RefreshToken)
			authRoutes.POST("/logout", authController.Logout)
			authRoutes.POST("/forgot-password", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.ForgotPassword)
			authRoutes.POST("/reset-password", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), authController.ResetPassword)
		}

		v1.POST("/waitlist", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), waitlistController.AddToWaitlist)

		// Protected auth routes
		protectedAuthRoutes := v1.Group("/auth")
		protectedAuthRoutes.Use(middleware.AuthMiddleware(jwtManager))
		{
			protectedAuthRoutes.POST("/change-password", authController.ChangePassword)
		}

		// Protected routes
		protectedRoutes := v1.Group("")
		protectedRoutes.Use(middleware.AuthMiddleware(jwtManager))
		{
			// User routes
			userRoutes := protectedRoutes.Group("/users")
			{
				userRoutes.GET("/profile", userController.GetProfile)
				userRoutes.PUT("/profile", userController.UpdateProfile)
				userRoutes.POST("/role", userController.UpdateRole)
				userRoutes.POST("/kyc", userController.SubmitKYC)
				userRoutes.GET("/kyc/status", userController.GetKYCStatus)
			}
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"version":   cfg.App.Version,
		})
	})

	// Readiness check endpoint
	router.GET("/ready", func(c *gin.Context) {
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(503, gin.H{
				"status":    "error",
				"message":   "Database connection failed",
				"timestamp": time.Now().UTC(),
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{
				"status":    "error",
				"message":   "Database ping failed",
				"timestamp": time.Now().UTC(),
			})
			return
		}

		// Check Redis connection if available
		if redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				c.JSON(503, gin.H{
					"status":    "error",
					"message":   "Redis connection failed",
					"timestamp": time.Now().UTC(),
				})
				return
			}
		}

		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	return router
}
