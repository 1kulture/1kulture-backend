package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/1kulture/1kulture-backend/internal/config"
	appLogger "github.com/1kulture/1kulture-backend/internal/utils/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	var err error

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  getLogLevel(cfg.Environment),
			IgnoreRecordNotFoundError: true,
			Colorful:                  cfg.Environment != "production",
		},
	)

	// Open database connection
	DB, err = gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger:                                   gormLogger,
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   false,
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Enable UUID extension
	if err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return fmt.Errorf("failed to enable uuid extension: %w", err)
	}

	appLogger.Info("Database connection established successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func getLogLevel(environment string) logger.LogLevel {
	switch environment {
	case "production":
		return logger.Warn
	case "staging":
		return logger.Info
	default:
		return logger.Info
	}
}
