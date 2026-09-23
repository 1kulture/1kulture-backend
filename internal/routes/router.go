package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/controllers"
	"github.com/1kulture/1kulture-backend/internal/middleware"
	"github.com/1kulture/1kulture-backend/internal/payments"
	"github.com/1kulture/1kulture-backend/internal/repositories"
	"github.com/1kulture/1kulture-backend/internal/services"
	"github.com/1kulture/1kulture-backend/internal/utils/email"
	"github.com/1kulture/1kulture-backend/internal/utils/jwt"
)

func SetupRouter(
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	jwtManager *jwt.JWTManager,
	paymentProvider payments.PaymentProvider,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.Security.AllowedOrigins))
	router.Use(middleware.RequestIDMiddleware())

	router.SetTrustedProxies(cfg.Security.TrustedProxies)

	// ==========================================================
	// REPOSITORIES
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

	// Phase 2
	ticketTypeRepo := repositories.NewTicketTypeRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	orderItemRepo := repositories.NewOrderItemRepository(db)
	ticketRepo := repositories.NewTicketRepository(db)
	promoCodeRepo := repositories.NewPromoCodeRepository(db)
	promoRedemptionRepo := repositories.NewPromoCodeRedemptionRepository(db)
	checkInRepo := repositories.NewCheckInRepository(db)
	refundRepo := repositories.NewRefundRepository(db)
	ledgerRepo := repositories.NewLedgerRepository(db)
	txnRepo := repositories.NewPaymentTransactionRepository(db)
	ticketTransferRepo := repositories.NewTicketTransferRepository(db)

	// Phase 2A — Brand Partnerships
	brandProfileRepo := repositories.NewBrandProfileRepository(db)
	eventPartnershipConfigRepo := repositories.NewEventPartnershipConfigRepository(db)
	partnershipOpportunityRepo := repositories.NewPartnershipOpportunityRepository(db)
	partnershipRequestRepo := repositories.NewPartnershipRequestRepository(db)
	partnershipMetricRepo := repositories.NewPartnershipMetricRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	affiliateCodeRepo := repositories.NewAffiliateCodeRepository(db)

	// ==========================================================
	// SERVICES
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

	// Phase 2
	ledgerService := services.NewLedgerService(ledgerRepo, orderRepo, settingRepo, auditLogRepo)
	ticketTypeService := services.NewTicketTypeService(ticketTypeRepo, eventRepo, coOrgRepo, auditLogRepo)
	promoCodeService := services.NewPromoCodeService(promoCodeRepo, promoRedemptionRepo, eventRepo, coOrgRepo, auditLogRepo)
	orderService := services.NewOrderService(
		orderRepo,
		orderItemRepo,
		ticketRepo,
		ticketTypeRepo,
		eventRepo,
		coOrgRepo,
		promoCodeRepo,
		promoRedemptionRepo,
		txnRepo,
		commissionService,
		ledgerService,
		settingService,
		paymentProvider,
		auditLogRepo,
		cfg,
	)
	ticketService := services.NewTicketService(ticketRepo, eventRepo, coOrgRepo)
	refundService := services.NewRefundService(refundRepo, orderRepo, ticketRepo, eventRepo, coOrgRepo, ledgerService, paymentProvider, auditLogRepo)
	ticketTransferService := services.NewTicketTransferService(ticketTransferRepo, ticketRepo, userRepo, settingService, auditLogRepo)
	checkInService := services.NewCheckInService(checkInRepo, ticketRepo, eventRepo, coOrgRepo, staffRepo, auditLogRepo)

	// Phase 2A — Brand Partnerships
	notificationService := services.NewNotificationService(notificationRepo)
	brandProfileService := services.NewBrandProfileService(
		brandProfileRepo,
		roleRepo,
		userRepo,
		partnershipRequestRepo,
		auditLogRepo,
	)
	partnershipConfigService := services.NewPartnershipConfigService(
		eventPartnershipConfigRepo,
		partnershipOpportunityRepo,
		eventRepo,
		coOrgRepo,
		auditLogRepo,
	)
	partnershipService := services.NewPartnershipService(
		partnershipRequestRepo,
		eventPartnershipConfigRepo,
		partnershipOpportunityRepo,
		brandProfileRepo,
		eventRepo,
		coOrgRepo,
		notificationService,
		auditLogRepo,
	)
	partnershipMetricService := services.NewPartnershipMetricService(
		partnershipMetricRepo,
		partnershipRequestRepo,
		coOrgRepo,
		notificationService,
		auditLogRepo,
	)
	affiliateCodeService := services.NewAffiliateCodeService(
		affiliateCodeRepo,
		promoCodeRepo,
		partnershipRequestRepo,
		eventRepo,
		coOrgRepo,
		auditLogRepo,
	)
	brandDiscoveryService := services.NewBrandDiscoveryService(
		eventRepo,
		eventPartnershipConfigRepo,
		partnershipOpportunityRepo,
		brandProfileRepo,
		partnershipRequestRepo,
		coOrgRepo,
	)

	// ==========================================================
	// CONTROLLERS
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

	// Phase 2
	ticketTypeController := controllers.NewTicketTypeController(ticketTypeService)
	orderController := controllers.NewOrderController(orderService)
	ticketController := controllers.NewTicketController(ticketService)
	promoCodeController := controllers.NewPromoCodeController(promoCodeService)
	refundController := controllers.NewRefundController(refundService)
	ticketTransferController := controllers.NewTicketTransferController(ticketTransferService)
	checkInController := controllers.NewCheckInController(checkInService)
	ledgerController := controllers.NewLedgerController(ledgerService)

	// Phase 2A — Brand Partnerships
	brandController := controllers.NewBrandController(brandProfileService)
	partnershipConfigController := controllers.NewPartnershipConfigController(partnershipConfigService)
	partnershipController := controllers.NewPartnershipController(partnershipService)
	partnershipMetricController := controllers.NewPartnershipMetricController(partnershipMetricService)
	affiliateCodeController := controllers.NewAffiliateCodeController(affiliateCodeService)
	notificationController := controllers.NewNotificationController(notificationService)
	brandDiscoveryController := controllers.NewBrandDiscoveryController(brandDiscoveryService)

	// ==========================================================
	// ROUTES
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
		// Webhooks (public, no auth)
		// ------------------------------------------------------
		v1.POST("/webhooks/paystack", orderController.PaystackWebhook)

		// ------------------------------------------------------
		// Public ticket-types listing & transfer lookup
		// ------------------------------------------------------
		v1.GET("/events/:id/ticket-types", ticketTypeController.List)
		v1.GET("/ticket-transfers/token/:token", ticketTransferController.GetByToken)

		// ------------------------------------------------------
		// Brand partnerships - Public
		// ------------------------------------------------------
		v1.GET("/events/:id/partnership-config", partnershipConfigController.Get)
		v1.GET("/events/:id/partnership-opportunities", partnershipConfigController.ListOpportunities)

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

			// ---------- PHASE 2: Ticketing ----------
			// Ticket types (manage)
			protectedRoutes.POST("/events/:id/ticket-types", ticketTypeController.Create)
			protectedRoutes.PUT("/events/:id/ticket-types/:tt_id", ticketTypeController.Update)
			protectedRoutes.DELETE("/events/:id/ticket-types/:tt_id", ticketTypeController.Delete)

			// Orders
			protectedRoutes.POST("/orders", orderController.Create)
			protectedRoutes.POST("/orders/:id/pay", orderController.InitializePayment)
			protectedRoutes.GET("/orders/me", orderController.ListMine)
			protectedRoutes.GET("/orders/:id", orderController.Get)
			protectedRoutes.GET("/events/:id/orders", orderController.ListEventOrders)

			// Tickets
			protectedRoutes.GET("/tickets/me", ticketController.ListMine)
			protectedRoutes.GET("/tickets/:id", ticketController.Get)
			protectedRoutes.GET("/tickets/code/:code", ticketController.GetByCode)
			protectedRoutes.GET("/events/:id/tickets", ticketController.ListEventTickets)

			// Promo codes
			protectedRoutes.POST("/events/:id/promo-codes", promoCodeController.Create)
			protectedRoutes.GET("/events/:id/promo-codes", promoCodeController.List)
			protectedRoutes.PUT("/events/:id/promo-codes/:promo_id", promoCodeController.Update)
			protectedRoutes.DELETE("/events/:id/promo-codes/:promo_id", promoCodeController.Delete)
			protectedRoutes.POST("/promo-codes/validate", promoCodeController.Validate)

			// Refunds
			protectedRoutes.POST("/orders/:id/refunds", refundController.Request)
			protectedRoutes.GET("/orders/:id/refunds", refundController.ListByOrder)
			protectedRoutes.GET("/refunds/:id", refundController.Get)
			protectedRoutes.POST("/refunds/:id/decide", refundController.Decide)

			// Ticket transfers
			protectedRoutes.POST("/tickets/:id/transfer", ticketTransferController.Initiate)
			protectedRoutes.POST("/ticket-transfers/accept", ticketTransferController.Accept)
			protectedRoutes.POST("/ticket-transfers/decline", ticketTransferController.Decline)
			protectedRoutes.POST("/ticket-transfers/:id/cancel", ticketTransferController.Cancel)
			protectedRoutes.GET("/ticket-transfers/me", ticketTransferController.ListMine)
			protectedRoutes.GET("/ticket-transfers/:id", ticketTransferController.Get)

			// Check-in
			protectedRoutes.POST("/events/:id/checkin/scan", checkInController.ScanQR)
			protectedRoutes.POST("/events/:id/checkin/manual", checkInController.Manual)
			protectedRoutes.GET("/events/:id/checkin", checkInController.List)
			protectedRoutes.GET("/events/:id/checkin/stats", checkInController.Stats)

			// Ledger
			protectedRoutes.GET("/ledger/balance", ledgerController.MyBalance)
			protectedRoutes.GET("/ledger/entries", ledgerController.ListMine)

			// ---------- PHASE 2A: Brand Partnerships ----------
			// Brand profile
			protectedRoutes.POST("/brands/profile", brandController.Create)
			protectedRoutes.GET("/brands/profile", brandController.GetMine)
			protectedRoutes.PUT("/brands/profile", brandController.UpdateMine)
			protectedRoutes.GET("/brands", brandController.List)
			protectedRoutes.GET("/brands/:id", brandController.GetByID)

			// Brand discovery
			protectedRoutes.GET("/brands/discover", brandDiscoveryController.DiscoverEvents)
			protectedRoutes.GET("/brands/recommended", brandDiscoveryController.Recommended)
			protectedRoutes.GET("/brands/dashboard", brandDiscoveryController.Dashboard)
			protectedRoutes.GET("/brands/events/:id", brandDiscoveryController.EventDetail)

			// Brand partnership requests
			protectedRoutes.POST("/events/:id/partnership-requests", partnershipController.Create)
			protectedRoutes.GET("/brands/partnerships", partnershipController.ListMine)
			protectedRoutes.POST("/partnerships/:id/cancel", partnershipController.Cancel)

			// Organizer partnership management
			protectedRoutes.PUT("/events/:id/partnership-config", partnershipConfigController.Upsert)
			protectedRoutes.POST("/events/:id/partnership-opportunities", partnershipConfigController.AddOpportunity)
			protectedRoutes.PUT("/events/:id/partnership-opportunities/:opp_id", partnershipConfigController.UpdateOpportunity)
			protectedRoutes.DELETE("/events/:id/partnership-opportunities/:opp_id", partnershipConfigController.DeleteOpportunity)
			protectedRoutes.GET("/events/:id/partnership-requests", partnershipController.ListEventRequests)
			protectedRoutes.POST("/partnerships/:id/decide", partnershipController.Decide)
			protectedRoutes.POST("/partnerships/:id/activate", partnershipController.Activate)
			protectedRoutes.POST("/partnerships/:id/complete", partnershipController.Complete)
			protectedRoutes.GET("/partnerships/dashboard", partnershipController.Dashboard)
			protectedRoutes.GET("/partnerships/:id", partnershipController.Get)

			// Partnership metrics
			protectedRoutes.POST("/partnerships/:id/metrics", partnershipMetricController.Add)
			protectedRoutes.GET("/partnerships/:id/metrics", partnershipMetricController.List)
			protectedRoutes.DELETE("/partnership-metrics/:id", partnershipMetricController.Delete)

			// Affiliate codes
			protectedRoutes.POST("/partnerships/:id/affiliate-codes", affiliateCodeController.Create)
			protectedRoutes.GET("/partnerships/:id/affiliate-codes", affiliateCodeController.List)

			// Notifications
			protectedRoutes.GET("/notifications", notificationController.List)
			protectedRoutes.GET("/notifications/unread-count", notificationController.CountUnread)
			protectedRoutes.PUT("/notifications/read-all", notificationController.MarkAllRead)
			protectedRoutes.PUT("/notifications/:id/read", notificationController.MarkRead)
			protectedRoutes.DELETE("/notifications/:id", notificationController.Delete)
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

			eventAuth.POST("/:id/occurrences", eventController.AddOccurrence)
			eventAuth.PUT("/:id/occurrences/:occ_id", eventController.UpdateOccurrence)
			eventAuth.DELETE("/:id/occurrences/:occ_id", eventController.DeleteOccurrence)

			eventAuth.POST("/:id/co-organizers", eventController.AddCoOrganizer)
			eventAuth.DELETE("/:id/co-organizers/:user_id", eventController.RemoveCoOrganizer)

			eventAuth.POST("/:id/staff", eventController.AddStaff)
			eventAuth.DELETE("/:id/staff/:user_id", eventController.RemoveStaff)

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

			// Commission
			adminGroup.POST("/commission/tiers", commissionController.CreateTier)
			adminGroup.GET("/commission/tiers", commissionController.ListTiers)
			adminGroup.PUT("/commission/tiers/:id", commissionController.UpdateTier)
			adminGroup.DELETE("/commission/tiers/:id", commissionController.DeleteTier)
			adminGroup.POST("/commission/overrides", commissionController.SetOverride)
			adminGroup.GET("/commission/overrides/:organizer_id", commissionController.GetOverride)
			adminGroup.DELETE("/commission/overrides/:organizer_id", commissionController.DeleteOverride)

			// Refunds
			adminGroup.GET("/refunds", refundController.ListAdmin)

			// Ledger
			adminGroup.GET("/ledger/:organizer_id/balance", ledgerController.OrganizerBalance)

			// Partnerships (admin view)
			adminGroup.GET("/partnerships", partnershipController.AdminList)
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
