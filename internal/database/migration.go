package database

import (
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	appLogger "github.com/1kulture/1kulture-backend/internal/utils/logger"
)

// AutoMigrate dynamically migrates all models and seeds baseline data.
func AutoMigrate(db *gorm.DB) error {
	appLogger.Info("Starting database migration...")

	// Create extension for UUID if not exists
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Create enum types if needed
	if err := createEnumTypes(db); err != nil {
		return fmt.Errorf("failed to create enum types: %w", err)
	}

	// List of all models
	modelList := []interface{}{
		// Phase 0 — Auth & Users
		&models.User{},
		&models.Role{},
		&models.RefreshToken{},
		&models.KYCVerification{},
		&models.AuditLog{},
		&models.EmailVerification{},
		&models.PasswordReset{},
		&models.WaitlistEntry{},

		// Phase 0 — Platform Config
		&models.Setting{},
		&models.Currency{},
		&models.CurrencyRequest{},
		&models.CommissionTier{},
		&models.OrganizerCommissionOverride{},

		// Phase 1 — Events
		&models.Event{},
		&models.EventOccurrence{},
		&models.EventCoOrganizer{},
		&models.EventFollower{},
		&models.EventOrganizerFollower{},
		&models.EventShare{},
		&models.EventStaff{},

		// Phase 2 — Ticketing
		&models.TicketType{},
		&models.Order{},
		&models.OrderItem{},
		&models.Ticket{},
		&models.PromoCode{},
		&models.PromoCodeRedemption{},
		&models.CheckIn{},
		&models.Refund{},
		&models.LedgerEntry{},
		&models.PaymentTransaction{},
		&models.TicketTransfer{},

		// Phase 2A — Brand Partnerships
		&models.BrandProfile{},
		&models.EventPartnershipConfig{},
		&models.PartnershipOpportunity{},
		&models.PartnershipRequest{},
		&models.PartnershipMetric{},
		&models.Notification{},
		&models.AffiliateCode{},

		// Phase 2B — Interests & Categories
		&models.EventCategory{},
		&models.UserInterest{},
	}

	// Perform migration
	for _, model := range modelList {
		if err := db.AutoMigrate(model); err != nil {
			appLogger.WithError(err).Error("Failed to migrate model: ", reflect.TypeOf(model).String())
			return fmt.Errorf("failed to migrate model %s: %w", reflect.TypeOf(model).String(), err)
		}
		appLogger.Info("Successfully migrated model: ", reflect.TypeOf(model).String())
	}

	// Post-migration fixes for existing columns
	// AutoMigrate won't widen an existing varchar to text, so we do it manually.
	db.Exec(`ALTER TABLE refresh_tokens ALTER COLUMN token TYPE TEXT;`)
	db.Exec(`ALTER TABLE ticket_transfers ALTER COLUMN token TYPE TEXT;`)

	// ==========================================================
	// Seeding
	// ==========================================================
	if err := seedDefaultData(db); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}
	if err := seedPhaseZeroData(db); err != nil {
		return fmt.Errorf("failed to seed phase zero data: %w", err)
	}
	if err := seedCategories(db); err != nil {
		return fmt.Errorf("failed to seed categories: %w", err)
	}
	if err := seedSampleEvents(db); err != nil {
		return fmt.Errorf("failed to seed sample events: %w", err)
	}

	appLogger.Info("Database migration completed successfully")
	return nil
}

// ==========================================================
// Enum types
// ==========================================================

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

// ==========================================================
// Default data — Roles
// ==========================================================

func seedDefaultData(db *gorm.DB) error {
	roles := []models.Role{
		{Name: string(models.RoleGuest), Description: "Guest user - can browse and purchase tickets"},
		{Name: string(models.RoleVendor), Description: "Vendor - can provide services for events"},
		{Name: string(models.RoleEventManager), Description: "Event Manager - can create and manage events"},
		{Name: string(models.RoleBrand), Description: "Brand - can partner with events"},
		{Name: string(models.RoleAdmin), Description: "System Administrator"},
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

// ==========================================================
// Phase 0 data — Settings, Currencies, Commission Tiers
// ==========================================================

func seedPhaseZeroData(db *gorm.DB) error {
	// ---------- Settings ----------
	defaults := []models.Setting{
		{Key: models.SettingEscrowReleaseDays, Value: "3", ValueType: "int", Category: "escrow", Description: "Days after event end before escrow is released.", IsPublic: false},
		{Key: models.SettingDefaultCommissionRate, Value: "500", ValueType: "int", Category: "commission", Description: "Default platform commission in basis points (500 = 5%).", IsPublic: false},
		{Key: models.SettingDefaultCurrency, Value: "NGN", ValueType: "string", Category: "currency", Description: "Platform default currency.", IsPublic: true},
		{Key: models.SettingPlatformFeeRate, Value: "0", ValueType: "int", Category: "commission", Description: "Platform processing fee in basis points.", IsPublic: false},
		{Key: models.SettingRefundPolicyDefaultDays, Value: "7", ValueType: "int", Category: "refund", Description: "Default refund window in days.", IsPublic: true},
		{Key: models.SettingRefundPolicyDefaultText, Value: "Refunds available up to 7 days before event start.", ValueType: "string", Category: "refund", Description: "Default refund policy text.", IsPublic: true},
		{Key: models.SettingEventMaxGalleryImages, Value: "10", ValueType: "int", Category: "event", Description: "Maximum gallery images per event.", IsPublic: true},
		{Key: models.SettingTicketTransferEnabled, Value: "true", ValueType: "bool", Category: "ticket", Description: "Enable ticket transfers.", IsPublic: true},
	}

	for _, s := range defaults {
		var count int64
		db.Model(&models.Setting{}).Where("key = ?", s.Key).Count(&count)
		if count == 0 {
			if err := db.Create(&s).Error; err != nil {
				return fmt.Errorf("failed to seed setting %s: %w", s.Key, err)
			}
		}
	}

	// ---------- Currencies ----------
	currencies := []models.Currency{
		{Code: "NGN", Name: "Nigerian Naira", Symbol: "₦", IsEnabled: true, IsDefault: true, MinAmountMinor: 10000},
		{Code: "USD", Name: "US Dollar", Symbol: "$", IsEnabled: false, IsDefault: false, MinAmountMinor: 100},
		{Code: "GBP", Name: "Pound Sterling", Symbol: "£", IsEnabled: false, IsDefault: false, MinAmountMinor: 100},
		{Code: "EUR", Name: "Euro", Symbol: "€", IsEnabled: false, IsDefault: false, MinAmountMinor: 100},
		{Code: "GHS", Name: "Ghanaian Cedi", Symbol: "₵", IsEnabled: false, IsDefault: false, MinAmountMinor: 100},
		{Code: "KES", Name: "Kenyan Shilling", Symbol: "KSh", IsEnabled: false, IsDefault: false, MinAmountMinor: 1000},
		{Code: "ZAR", Name: "South African Rand", Symbol: "R", IsEnabled: false, IsDefault: false, MinAmountMinor: 1000},
	}

	for _, c := range currencies {
		var count int64
		db.Model(&models.Currency{}).Where("code = ?", c.Code).Count(&count)
		if count == 0 {
			if err := db.Create(&c).Error; err != nil {
				return fmt.Errorf("failed to seed currency %s: %w", c.Code, err)
			}
		}
	}

	// ---------- Commission Tiers ----------
	tiers := []models.CommissionTier{
		{Name: "Starter", MinRevenueMinor: 0, MaxRevenueMinor: int64Ptr(50_000_00), RateBps: 700, Priority: 1, IsActive: true},
		{Name: "Growing", MinRevenueMinor: 50_000_00, MaxRevenueMinor: int64Ptr(500_000_00), RateBps: 600, Priority: 2, IsActive: true},
		{Name: "Established", MinRevenueMinor: 500_000_00, MaxRevenueMinor: int64Ptr(5_000_000_00), RateBps: 500, Priority: 3, IsActive: true},
		{Name: "Enterprise", MinRevenueMinor: 5_000_000_00, MaxRevenueMinor: nil, RateBps: 400, Priority: 4, IsActive: true},
	}

	for _, t := range tiers {
		var count int64
		db.Model(&models.CommissionTier{}).Where("name = ?", t.Name).Count(&count)
		if count == 0 {
			if err := db.Create(&t).Error; err != nil {
				return fmt.Errorf("failed to seed commission tier %s: %w", t.Name, err)
			}
		}
	}
	return nil
}

// ==========================================================
// Event Categories
// ==========================================================

func seedCategories(db *gorm.DB) error {
	categories := []models.EventCategory{
		{Slug: "concerts", Name: "Concert", Icon: "music", SortOrder: 1, IsActive: true},
		{Slug: "corporate", Name: "Corporate", Icon: "briefcase", SortOrder: 2, IsActive: true},
		{Slug: "birthdays", Name: "Birthdays", Icon: "cake", SortOrder: 3, IsActive: true},
		{Slug: "conferences", Name: "Conferences", Icon: "presentation", SortOrder: 4, IsActive: true},
		{Slug: "picnics", Name: "Picnics", Icon: "sun", SortOrder: 5, IsActive: true},
		{Slug: "anniversary", Name: "Anniversary", Icon: "heart", SortOrder: 6, IsActive: true},
		{Slug: "tech-events", Name: "Tech Events", Icon: "cpu", SortOrder: 7, IsActive: true},
		{Slug: "religious-events", Name: "Religious Events", Icon: "church", SortOrder: 8, IsActive: true},
		{Slug: "nightlife", Name: "Nightlife", Icon: "moon", SortOrder: 9, IsActive: true},
		{Slug: "sports", Name: "Sports", Icon: "activity", SortOrder: 10, IsActive: true},
		{Slug: "arts-theatre", Name: "Arts & Theatre", Icon: "theater", SortOrder: 11, IsActive: true},
		{Slug: "food-drinks", Name: "Food & Drinks", Icon: "utensils", SortOrder: 12, IsActive: true},
	}

	for _, c := range categories {
		var count int64
		db.Model(&models.EventCategory{}).Where("slug = ?", c.Slug).Count(&count)
		if count == 0 {
			if err := db.Create(&c).Error; err != nil {
				return fmt.Errorf("failed to seed category %s: %w", c.Slug, err)
			}
			appLogger.Info("Seeded category: ", c.Slug)
		}
	}
	return nil
}

// ==========================================================
// Sample Events
// ==========================================================

func seedSampleEvents(db *gorm.DB) error {
	// Only seed if no events exist yet (safety so it doesn't run on prod that has real data)
	var count int64
	db.Model(&models.Event{}).Count(&count)
	if count > 0 {
		appLogger.Info("Skipping event seed — events already exist")
		return nil
	}

	// Find an existing event manager to be the organizer.
	// If none exists, skip — real users will create events.
	var organizer models.User
	if err := db.Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", "event_manager").
		First(&organizer).Error; err != nil {
		appLogger.Info("Skipping event seed — no event manager exists yet")
		return nil
	}

	now := time.Now().UTC()

	events := []models.Event{
		{
			OrganizerID:    organizer.ID,
			Slug:           "lagos-music-fest-2026",
			Title:          "Lagos Music Fest 2026",
			Summary:        "The biggest Afrobeats gathering in West Africa.",
			Description:    "Three nights of non-stop Afrobeats, Amapiano, and Afro-fusion. Featuring top Nigerian and international artists.",
			Category:       "concerts",
			EventType:      models.EventTypeInPerson,
			VenueName:      "Eko Convention Centre",
			VenueCity:      "Lagos",
			VenueCountry:   "Nigeria",
			Timezone:       "Africa/Lagos",
			StartAt:        now.AddDate(0, 1, 10),
			EndAt:          now.AddDate(0, 1, 12),
			Status:         models.EventStatusPublished,
			Visibility:     models.EventVisibilityPublic,
			Currency:       "NGN",
			IsFeatured:     true,
			AgeRestriction: models.Age18Plus,
		},
		{
			OrganizerID:  organizer.ID,
			Slug:         "techcrunch-lagos-2026",
			Title:        "TechCrunch Lagos 2026",
			Summary:      "Where African founders meet global investors.",
			Description:  "A one-day summit bringing together 500+ founders, VCs, and operators across Africa's tech ecosystem.",
			Category:     "tech-events",
			EventType:    models.EventTypeHybrid,
			VenueName:    "Landmark Centre",
			VenueCity:    "Lagos",
			VenueCountry: "Nigeria",
			Timezone:     "Africa/Lagos",
			StartAt:      now.AddDate(0, 2, 5),
			EndAt:        now.AddDate(0, 2, 5).Add(9 * time.Hour),
			VirtualURL:   "https://meet.example.com/techcrunch-lagos",
			Status:       models.EventStatusPublished,
			Visibility:   models.EventVisibilityPublic,
			Currency:     "NGN",
			IsFeatured:   true,
		},
		{
			OrganizerID:  organizer.ID,
			Slug:         "abuja-business-conference",
			Title:        "Abuja Business Conference",
			Summary:      "Nigeria's flagship corporate leadership summit.",
			Description:  "Two days of keynotes, panels, and networking for C-suite executives and entrepreneurs.",
			Category:     "conferences",
			EventType:    models.EventTypeInPerson,
			VenueName:    "Transcorp Hilton",
			VenueCity:    "Abuja",
			VenueCountry: "Nigeria",
			Timezone:     "Africa/Lagos",
			StartAt:      now.AddDate(0, 1, 20),
			EndAt:        now.AddDate(0, 1, 21),
			Status:       models.EventStatusPublished,
			Visibility:   models.EventVisibilityPublic,
			Currency:     "NGN",
		},
		{
			OrganizerID:    organizer.ID,
			Slug:           "detty-december-beach-party",
			Title:          "Detty December Beach Party",
			Summary:        "Lagos' wildest beach party of the season.",
			Description:    "Sun, sand, and Afro-house all day long. Featuring top DJs and live performances.",
			Category:       "nightlife",
			EventType:      models.EventTypeInPerson,
			VenueName:      "Elegushi Beach",
			VenueCity:      "Lagos",
			VenueCountry:   "Nigeria",
			Timezone:       "Africa/Lagos",
			StartAt:        now.AddDate(0, 2, 15),
			EndAt:          now.AddDate(0, 2, 15).Add(12 * time.Hour),
			Status:         models.EventStatusPublished,
			Visibility:     models.EventVisibilityPublic,
			Currency:       "NGN",
			AgeRestriction: models.Age18Plus,
		},
		{
			OrganizerID:    organizer.ID,
			Slug:           "family-picnic-ikoyi",
			Title:          "Family Picnic at Ikoyi",
			Summary:        "A relaxed family day out with games, food, and music.",
			Description:    "Bring the whole family for a full-day picnic with bouncy castles, food stalls, and live acoustic music.",
			Category:       "picnics",
			EventType:      models.EventTypeInPerson,
			VenueName:      "Ikoyi Park",
			VenueCity:      "Lagos",
			VenueCountry:   "Nigeria",
			Timezone:       "Africa/Lagos",
			StartAt:        now.AddDate(0, 0, 14),
			EndAt:          now.AddDate(0, 0, 14).Add(8 * time.Hour),
			Status:         models.EventStatusPublished,
			Visibility:     models.EventVisibilityPublic,
			Currency:       "NGN",
			AgeRestriction: models.AgeAllAges,
		},
		{
			OrganizerID:  organizer.ID,
			Slug:         "gospel-praise-night",
			Title:        "Gospel Praise Night",
			Summary:      "An evening of worship and praise with top gospel artists.",
			Description:  "A spirit-filled night of worship featuring Nigeria's leading gospel ministers.",
			Category:     "religious-events",
			EventType:    models.EventTypeInPerson,
			VenueName:    "The Rock Cathedral",
			VenueCity:    "Lagos",
			VenueCountry: "Nigeria",
			Timezone:     "Africa/Lagos",
			StartAt:      now.AddDate(0, 1, 5),
			EndAt:        now.AddDate(0, 1, 5).Add(5 * time.Hour),
			Status:       models.EventStatusPublished,
			Visibility:   models.EventVisibilityPublic,
			Currency:     "NGN",
		},
		{
			OrganizerID:  organizer.ID,
			Slug:         "lagos-fashion-week",
			Title:        "Lagos Fashion Week",
			Summary:      "Africa's premier fashion showcase.",
			Description:  "Five days of runway shows, showrooms, and industry conversations featuring Africa's top designers.",
			Category:     "arts-theatre",
			EventType:    models.EventTypeInPerson,
			VenueName:    "Federal Palace Hotel",
			VenueCity:    "Lagos",
			VenueCountry: "Nigeria",
			Timezone:     "Africa/Lagos",
			StartAt:      now.AddDate(0, 3, 1),
			EndAt:        now.AddDate(0, 3, 5),
			Status:       models.EventStatusPublished,
			Visibility:   models.EventVisibilityPublic,
			Currency:     "NGN",
		},
		{
			OrganizerID:    organizer.ID,
			Slug:           "street-food-festival",
			Title:          "Street Food Festival",
			Summary:        "The best street food in Lagos under one roof.",
			Description:    "50+ vendors serving everything from suya to shawarma. Live music all day.",
			Category:       "food-drinks",
			EventType:      models.EventTypeInPerson,
			VenueName:      "Muritala Muhammed Square",
			VenueCity:      "Lagos",
			VenueCountry:   "Nigeria",
			Timezone:       "Africa/Lagos",
			StartAt:        now.AddDate(0, 0, 25),
			EndAt:          now.AddDate(0, 0, 26),
			Status:         models.EventStatusPublished,
			Visibility:     models.EventVisibilityPublic,
			Currency:       "NGN",
			AgeRestriction: models.AgeAllAges,
		},
		{
			OrganizerID:  organizer.ID,
			Slug:         "lagos-marathon-2027",
			Title:        "Lagos City Marathon 2027",
			Summary:      "42km through the heart of Lagos.",
			Description:  "West Africa's biggest marathon. Join 30,000 runners on a scenic route across Lagos.",
			Category:     "sports",
			EventType:    models.EventTypeInPerson,
			VenueName:    "Lagos Island",
			VenueCity:    "Lagos",
			VenueCountry: "Nigeria",
			Timezone:     "Africa/Lagos",
			StartAt:      now.AddDate(0, 4, 1),
			EndAt:        now.AddDate(0, 4, 1).Add(8 * time.Hour),
			Status:       models.EventStatusPublished,
			Visibility:   models.EventVisibilityPublic,
			Currency:     "NGN",
		},
		{
			OrganizerID:    organizer.ID,
			Slug:           "milestone-40th-birthday",
			Title:          "Milestone 40th Birthday Celebration",
			Summary:        "An intimate evening celebrating 40 years.",
			Description:    "Dinner, dancing, and toasts. Black-tie optional.",
			Category:       "birthdays",
			EventType:      models.EventTypeInPerson,
			VenueName:      "The Wheatbaker Hotel",
			VenueCity:      "Lagos",
			VenueCountry:   "Nigeria",
			Timezone:       "Africa/Lagos",
			StartAt:        now.AddDate(0, 1, 15),
			EndAt:          now.AddDate(0, 1, 15).Add(6 * time.Hour),
			Status:         models.EventStatusPublished,
			Visibility:     models.EventVisibilityPublic,
			Currency:       "NGN",
			AgeRestriction: models.Age21Plus,
		},
	}

	for _, e := range events {
		if err := db.Create(&e).Error; err != nil {
			return fmt.Errorf("failed to seed event %s: %w", e.Slug, err)
		}
		appLogger.Info("Seeded event: ", e.Slug)
	}
	return nil
}

// ==========================================================
// Helpers
// ==========================================================

func int64Ptr(v int64) *int64 { return &v }