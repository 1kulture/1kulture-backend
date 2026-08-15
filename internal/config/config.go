package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Email       EmailConfig
	Redis       RedisConfig
	RateLimit   RateLimitConfig
	Security    SecurityConfig
	Environment string
	App         AppConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	APIURL      string
	WebURL      string
	Issuer      string
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

type JWTConfig struct {
	Secret             string
	RefreshSecret      string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
	Audience           string
}

type EmailConfig struct {
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	FromName      string
	FromAddress   string
	TemplatesPath string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

type SecurityConfig struct {
	BCryptCost           int
	MaxLoginAttempts     int
	LoginBlockDuration   time.Duration
	PasswordResetTimeout time.Duration
	VerificationTimeout  time.Duration
	MaxVerificationRetry int
	Enable2FA            bool
	AllowedOrigins       []string
	TrustedProxies       []string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read .env file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{
		App: AppConfig{
			Name:        viper.GetString("APP_NAME"),
			Version:     viper.GetString("APP_VERSION"),
			Environment: viper.GetString("ENVIRONMENT"),
			APIURL:      viper.GetString("API_URL"),
			WebURL:      viper.GetString("WEB_URL"),
			Issuer:      viper.GetString("JWT_ISSUER"),
		},
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Host: viper.GetString("SERVER_HOST"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
			Timezone: viper.GetString("DB_TIMEZONE"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			RefreshSecret:      viper.GetString("JWT_REFRESH_SECRET"),
			AccessTokenExpiry:  viper.GetDuration("JWT_ACCESS_TOKEN_EXPIRY"),
			RefreshTokenExpiry: viper.GetDuration("JWT_REFRESH_TOKEN_EXPIRY"),
			Issuer:             viper.GetString("JWT_ISSUER"),
			Audience:           viper.GetString("JWT_AUDIENCE"),
		},
		Email: EmailConfig{
			SMTPHost:      viper.GetString("SMTP_HOST"),
			SMTPPort:      viper.GetInt("SMTP_PORT"),
			SMTPUsername:  viper.GetString("SMTP_USERNAME"),
			SMTPPassword:  viper.GetString("SMTP_PASSWORD"),
			FromName:      viper.GetString("EMAIL_FROM_NAME"),
			FromAddress:   viper.GetString("EMAIL_FROM_ADDRESS"),
			TemplatesPath: viper.GetString("EMAIL_TEMPLATES_PATH"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: viper.GetDuration("RATE_LIMIT_DURATION"),
		},
		Security: SecurityConfig{
			BCryptCost:           viper.GetInt("BCRYPT_COST"),
			MaxLoginAttempts:     viper.GetInt("MAX_LOGIN_ATTEMPTS"),
			LoginBlockDuration:   viper.GetDuration("LOGIN_BLOCK_DURATION"),
			PasswordResetTimeout: viper.GetDuration("PASSWORD_RESET_TIMEOUT"),
			VerificationTimeout:  viper.GetDuration("VERIFICATION_TIMEOUT"),
			MaxVerificationRetry: viper.GetInt("MAX_VERIFICATION_RETRY"),
			Enable2FA:            viper.GetBool("ENABLE_2FA"),
			AllowedOrigins:       viper.GetStringSlice("ALLOWED_ORIGINS"),
			TrustedProxies:       viper.GetStringSlice("TRUSTED_PROXIES"),
		},
		Environment: viper.GetString("ENVIRONMENT"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults() {
	viper.SetDefault("APP_NAME", "1Kulture")
	viper.SetDefault("APP_VERSION", "1.0.0")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_TIMEZONE", "UTC")
	viper.SetDefault("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_TOKEN_EXPIRY", "168h")
	viper.SetDefault("JWT_ISSUER", "1kulture-api")
	viper.SetDefault("JWT_AUDIENCE", "1kulture-web")
	viper.SetDefault("BCRYPT_COST", 12)
	viper.SetDefault("MAX_LOGIN_ATTEMPTS", 5)
	viper.SetDefault("LOGIN_BLOCK_DURATION", "15m")
	viper.SetDefault("PASSWORD_RESET_TIMEOUT", "1h")
	viper.SetDefault("VERIFICATION_TIMEOUT", "30m")
	viper.SetDefault("MAX_VERIFICATION_RETRY", 5)
	viper.SetDefault("ENABLE_2FA", false)
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_DURATION", "1m")
	viper.SetDefault("ALLOWED_ORIGINS", []string{"http://localhost:3000"})
	viper.SetDefault("TRUSTED_PROXIES", []string{"127.0.0.1"})
	viper.SetDefault("SMTP_PORT", 587)
}

func (c *Config) validate() error {
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-key-change-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	if c.JWT.RefreshSecret == "" || c.JWT.RefreshSecret == "your-refresh-secret-key-change-in-production" {
		return fmt.Errorf("JWT_REFRESH_SECRET must be set to a secure value")
	}
	if c.JWT.Issuer == "" {
		return fmt.Errorf("JWT_ISSUER must be set")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD must be set")
	}
	if c.Security.BCryptCost < 10 || c.Security.BCryptCost > 15 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 15")
	}
	if c.App.Environment == "production" {
		if c.Database.SSLMode != "require" && c.Database.SSLMode != "verify-full" {
			return fmt.Errorf("DB_SSLMODE must be 'require' or 'verify-full' in production")
		}
	}
	return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
		c.Database.Timezone,
	)
}
