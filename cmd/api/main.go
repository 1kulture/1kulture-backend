package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/database"
	"github.com/1kulture/1kulture-backend/internal/routes"
	"github.com/1kulture/1kulture-backend/internal/utils/jwt"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	// Import your own docs package
	_ "github.com/1kulture/1kulture-backend/docs"
)

// @title 1Kulture API
// @version 1.0
// @description Enterprise Event Management System API with multi-tenant SaaS architecture
// @termsOfService https://1kulture.com/terms

// @contact.name API Support
// @contact.url https://1kulture.com/support
// @contact.email support@1kulture.com

// @license.name Proprietary
// @license.url https://1kulture.com/license

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	CommitSHA = "unknown"
)

type Application struct {
	config     *config.Config
	router     *gin.Engine
	db         *gorm.DB
	redis      *redis.Client
	jwtManager *jwt.JWTManager
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.App.Environment)
	logger.Info("Starting 1Kulture API server...")
	logger.Info(fmt.Sprintf("Environment: %s", cfg.App.Environment))
	logger.Info(fmt.Sprintf("Version: %s", Version))
	logger.Info(fmt.Sprintf("Build Time: %s", BuildTime))
	logger.Info(fmt.Sprintf("Commit SHA: %s", CommitSHA))

	// Initialize validator
	validator.Init()

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create application instance
	app := &Application{
		config: cfg,
	}

	// Initialize database
	if err := app.initDatabase(); err != nil {
		logger.Fatal("Failed to initialize database:", err)
	}
	defer app.closeDatabase()

	// Initialize Redis (optional in development)
	if err := app.initRedis(); err != nil {
		if cfg.App.Environment == "production" {
			logger.Fatal("Failed to initialize Redis:", err)
		} else {
			logger.Warning("Redis not available, continuing without it: ", err)
			app.redis = nil
		}
	}
	defer app.closeRedis()

	// Initialize JWT manager
	app.jwtManager = jwt.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Initialize router
	app.router = routes.SetupRouter(cfg, app.db, app.redis, app.jwtManager)

	// Setup Swagger (conditionally)
	app.setupSwagger()

	// Start server
	app.startServer()
}

func (app *Application) initDatabase() error {
	if err := database.Init(app.config); err != nil {
		return err
	}
	app.db = database.GetDB()

	// Run migrations
	if err := database.AutoMigrate(app.db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database initialized successfully")
	return nil
}

func (app *Application) closeDatabase() {
	if err := database.Close(); err != nil {
		logger.Error("Failed to close database connection:", err)
	}
}

func (app *Application) initRedis() error {
	app.redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", app.config.Redis.Host, app.config.Redis.Port),
		Password: app.config.Redis.Password,
		DB:       app.config.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully")
	return nil
}

func (app *Application) closeRedis() {
	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			logger.Error("Failed to close Redis connection:", err)
		}
	}
}

func (app *Application) setupSwagger() {
	if os.Getenv("ENABLE_SWAGGER") == "true" || app.config.App.Environment != "production" {
		app.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		logger.Info("Swagger documentation available at /swagger/index.html")
	} else {
		logger.Info("Swagger is disabled in production")
	}
}

func (app *Application) startServer() {
	serverAddr := fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port)

	srv := &http.Server{
		Addr:           serverAddr,
		Handler:        app.router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server is running on http://%s", serverAddr))
		logger.Info(fmt.Sprintf("Health check: http://%s/health", serverAddr))
		if app.config.App.Environment != "production" {
			logger.Info(fmt.Sprintf("Swagger: http://%s/swagger/index.html", serverAddr))
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited properly")
}
