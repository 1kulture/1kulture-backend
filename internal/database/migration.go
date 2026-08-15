package database

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	appLogger "github.com/1kulture/1kulture-backend/internal/utils/logger"
)

// AutoMigrate dynamically migrates all models
func AutoMigrate(db *gorm.DB) error {
	appLogger.Info("Starting database migration...")

	// List of all models
	modelList := []interface{}{
		&models.User{},
		&models.Role{},
		&models.RefreshToken{},
		&models.KYCVerification{},
		&models.AuditLog{},
		&models.EmailVerification{},
		&models.PasswordReset{},
		// Add more models here as they are created
	}

	// Create extension for UUID if not exists
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Create enum types if needed
	if err := createEnumTypes(db); err != nil {
		return fmt.Errorf("failed to create enum types: %w", err)
	}

	// Perform migration
	for _, model := range modelList {
		if err := db.AutoMigrate(model); err != nil {
			appLogger.WithError(err).Error("Failed to migrate model: ", reflect.TypeOf(model).String())
			return fmt.Errorf("failed to migrate model %s: %w", reflect.TypeOf(model).String(), err)
		}
		appLogger.Info("Successfully migrated model: ", reflect.TypeOf(model).String())
	}

	// Seed default data
	if err := seedDefaultData(db); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}

	appLogger.Info("Database migration completed successfully")
	return nil
}

func createEnumTypes(db *gorm.DB) error {
	enums := []string{
		`DO $$ BEGIN
			CREATE TYPE user_status AS ENUM ('pending', 'active', 'suspended', 'deleted');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
		`DO $$ BEGIN
			CREATE TYPE kyc_status AS ENUM ('pending', 'approved', 'rejected');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
	}

	for _, enum := range enums {
		if err := db.Exec(enum).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedDefaultData(db *gorm.DB) error {
	// Seed default roles
	roles := []models.Role{
		{
			Name:        string(models.RoleGuest),
			Description: "Guest user - can browse and purchase tickets",
		},
		{
			Name:        string(models.RoleVendor),
			Description: "Vendor - can provide services for events",
		},
		{
			Name:        string(models.RoleEventManager),
			Description: "Event Manager - can create and manage events",
		},
		{
			Name:        string(models.RoleAdmin),
			Description: "System Administrator",
		},
	}

	for _, role := range roles {
		var count int64
		db.Model(&models.Role{}).Where("name = ?", role.Name).Count(&count)
		if count == 0 {
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to seed role %s: %w", role.Name, err)
			}
			appLogger.Info("Seeded role: ", role.Name)
		}
	}

	return nil
}
