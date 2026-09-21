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

	// Trusted proxies
	router.SetTrustedProxies(cfg.Security.TrustedProxies)

	// ==========================================================
	// Repositories
	// ==========================================================
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	emailVerifRepo := repositories.NewEmailVerificationRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)
	kycRepo := repositories.NewKYCRepository(db)
	waitlistRepo := repositories.NewWaitlistRepository(db)

	// Phase 0
	settingRepo := repositories.NewSettingRepository(db)
	currencyRepo := repositories.NewCurrencyRepository(db)
	currencyReqRepo := repositories.NewCurrencyRequestRepository(db)
	commissionTierRepo := repositories.NewCommissionTierRepository(db)
	commissionOverrideRepo := repositories.NewOrganizerCommissionOverrideRepository(db)

	// Phase 1
	eventRepo := repositories.NewEventRepository(db)
	occRepo := repositories.NewEventOccurrenceRepository(db)
	coOrgRepo := repositories.NewEventCoOrganizerRepository(db)
	staffRepo := repositories.NewEventStaffRepository(db)
	followerRepo := repositories.NewEventFollowerRepository(db)
	orgFollowerRepo := repositories.NewOrganizerFollowerRepository(db)
	shareRepo := repositories.NewEventShareRepository(db)

	// ==========================================================
	// Services
	// ==========================================================
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

	// Phase 0
	settingService := services.NewSettingService(settingRepo, auditLogRepo)
	currencyService := services.NewCurrencyService(currencyRepo, currencyReqRepo, auditLogRepo)
	commissionService := services.NewCommissionService(
		commissionTierRepo,
		commissionOverrideRepo,
		settingRepo,
		auditLogRepo,
	)

	// Phase 1
	eventService := services.NewEventService(
		eventRepo,
		occRepo,
		coOrgRepo,
		staffRepo,
		followerRepo,
		orgFollowerRepo,
		shareRepo,
		currencyRepo,
		settingService,
		auditLogRepo,
	)

	// ==========================================================
	// Controllers
	// ==========================================================
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	waitlistController := controllers.NewWaitlistController(waitlistService)

	// Phase 0
	settingController := controllers.NewSettingController(settingService)
	currencyController := controllers.NewCurrencyController(currencyService)
	commissionController := controllers.NewCommissionController(commissionService)

	// Phase 1
	eventController := controllers.NewEventController(eventService)

	// ==========================================================
	// Routes
	// ==========================================================
	v1 := router.Group("/api/v1")
	{
		// ------------------------------------------------------
		// AUTH (public)
		// ------------------------------------------------------
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

		// Authenticated auth
		protectedAuthRoutes := v1.Group("/auth")
		protectedAuthRoutes.Use(middleware.AuthMiddleware(jwtManager))
		{
			protectedAuthRoutes.POST("/change-password", authController.ChangePassword)
		}

		// ------------------------------------------------------
		// Waitlist (public)
		// ------------------------------------------------------
		v1.POST("/waitlist", middleware.RateLimitMiddleware(redisClient, cfg.RateLimit), waitlistController.AddToWaitlist)

		// ------------------------------------------------------
		// Public settings & currencies
		// ------------------------------------------------------
		v1.GET("/settings/public", settingController.PublicSettings)
		v1.GET("/currencies", currencyController.ListEnabled)

		// ------------------------------------------------------
		// Public events & optional-auth events
		// ------------------------------------------------------
		v1.GET("/events", middleware.OptionalAuthMiddleware(jwtManager), eventController.List)
		v1.GET("/events/slug/:slug", middleware.OptionalAuthMiddleware(jwtManager), eventController.GetBySlug)
		v1.GET("/events/:id", middleware.OptionalAuthMiddleware(jwtManager), eventController.GetByID)
		v1.POST("/events/:id/share", middleware.OptionalAuthMiddleware(jwtManager), eventController.RecordShare)
		v1.GET("/events/:id/occurrences", eventController.ListOccurrences)
		v1.GET("/events/:id/co-organizers", eventController.ListCoOrganizers)
		v1.GET("/events/:id/staff", eventController.ListStaff)

		// ------------------------------------------------------
		// Authenticated general group
		// ------------------------------------------------------
		protectedRoutes := v1.Group("")
		protectedRoutes.Use(middleware.AuthMiddleware(jwtManager))
		{
			// Users
			userRoutes := protectedRoutes.Group("/users")
			{
				userRoutes.GET("/profile", userController.GetProfile)
				userRoutes.PUT("/profile", userController.UpdateProfile)
				userRoutes.POST("/role", userController.UpdateRole)
				userRoutes.POST("/kyc", userController.SubmitKYC)
				userRoutes.GET("/kyc/status", userController.GetKYCStatus)
			}

			// Currency requests (EventManager / Vendor)
			protectedRoutes.POST("/currency-requests", currencyController.RequestCurrency)
			protectedRoutes.GET("/currency-requests/me", currencyController.ListMyRequests)

			// Organizer follow/unfollow
			protectedRoutes.POST("/organizers/:organizer_id/follow", eventController.FollowOrganizer)
			protectedRoutes.DELETE("/organizers/:organizer_id/follow", eventController.UnfollowOrganizer)
		}

		// ------------------------------------------------------
		// Authenticated events (create/manage)
		// ------------------------------------------------------
		eventAuth := v1.Group("/events")
		eventAuth.Use(middleware.AuthMiddleware(jwtManager))
		{
			eventAuth.POST("", eventController.Create)
			eventAuth.PUT("/:id", eventController.Update)
			eventAuth.DELETE("/:id", eventController.Delete)
			eventAuth.POST("/:id/publish", eventController.Publish)
			eventAuth.POST("/:id/unpublish", eventController.Unpublish)
			eventAuth.POST("/:id/cancel", eventController.Cancel)
			eventAuth.POST("/:id/postpone", eventController.Postpone)

			// Occurrences
			eventAuth.POST("/:id/occurrences", eventController.AddOccurrence)
			eventAuth.PUT("/:id/occurrences/:occ_id", eventController.UpdateOccurrence)
			eventAuth.DELETE("/:id/occurrences/:occ_id", eventController.DeleteOccurrence)

			// Co-organizers
			eventAuth.POST("/:id/co-organizers", eventController.AddCoOrganizer)
			eventAuth.DELETE("/:id/co-organizers/:user_id", eventController.RemoveCoOrganizer)

			// Staff
			eventAuth.POST("/:id/staff", eventController.AddStaff)
			eventAuth.DELETE("/:id/staff/:user_id", eventController.RemoveStaff)

			// Follow / unfollow / followers
			eventAuth.POST("/:id/follow", eventController.FollowEvent)
			eventAuth.DELETE("/:id/follow", eventController.UnfollowEvent)
			eventAuth.GET("/:id/followers", eventController.ListEventFollowers)
		}

		// ------------------------------------------------------
		// Admin group
		// ------------------------------------------------------
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(jwtManager))
		adminGroup.Use(middleware.AdminMiddleware())
		{
			// Settings
			adminGroup.POST("/settings", settingController.Create)
			adminGroup.GET("/settings", settingController.List)
			adminGroup.PUT("/settings", settingController.Update)
			adminGroup.PUT("/settings/bulk", settingController.BulkUpdate)
			adminGroup.GET("/settings/:key", settingController.Get)
			adminGroup.DELETE("/settings/:key", settingController.Delete)

			// Currencies
			adminGroup.POST("/currencies", currencyController.Create)
			adminGroup.GET("/currencies", currencyController.ListAdmin)
			adminGroup.PUT("/currencies/:code", currencyController.Update)
			adminGroup.POST("/currencies/:code/default", currencyController.SetDefault)

			// Currency requests
			adminGroup.GET("/currency-requests", currencyController.ListAllRequests)
			adminGroup.POST("/currency-requests/:id/review", currencyController.ReviewRequest)

			// Commission tiers
			adminGroup.POST("/commission/tiers", commissionController.CreateTier)
			adminGroup.GET("/commission/tiers", commissionController.ListTiers)
			adminGroup.PUT("/commission/tiers/:id", commissionController.UpdateTier)
			adminGroup.DELETE("/commission/tiers/:id", commissionController.DeleteTier)

			// Commission overrides
			adminGroup.POST("/commission/overrides", commissionController.SetOverride)
			adminGroup.GET("/commission/overrides/:organizer_id", commissionController.GetOverride)
			adminGroup.DELETE("/commission/overrides/:organizer_id", commissionController.DeleteOverride)
		}
	}

	// ==========================================================
	// Health & readiness
	// ==========================================================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"version":   cfg.App.Version,
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(503, gin.H{"status": "error", "message": "Database connection failed", "timestamp": time.Now().UTC()})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "error", "message": "Database ping failed", "timestamp": time.Now().UTC()})
			return
		}
		if redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				c.JSON(503, gin.H{"status": "error", "message": "Redis connection failed", "timestamp": time.Now().UTC()})
				return
			}
		}
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
	})

	return router
}
