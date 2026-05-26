package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Application
	AppEnv   string
	AppName  string
	AppDebug bool

	// Domains
	Domain        string
	APIDomain     string
	WSDomain      string
	ConsoleDomain string
	DNSDomainName string
	SMTPDomainName string

	// Ports
	FrontendPort int
	APIPort      int
	WSPort       int
	ConsolePort  int
	DNSPort      int
	SMTPPort     int
	WorkerPort   int

	// Database
	Database DatabaseConfig

	// Redis
	Redis RedisConfig

	// JWT
	JWT JWTConfig

	// Security
	CSRFSecret   string
	CookieDomain string
	CookieSecure bool
	CORSOrigins  string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   int

	// Payload
	PayloadMaxFree    int64
	PayloadMaxPremium int64

	// Retention
	RetentionFree       int
	RetentionPro        int
	RetentionTeam       int
	RetentionEnterprise int

	// Monitoring
	MetricsEnabled   bool
	MetricsPort      int
	MetricsPath      string
	MetricsAuthToken string

	// Logging
	LogLevel  string
	LogFormat string

	// Billing
	PayPal  PayPalConfig
	PayHere PayHereConfig

	// TLS
	TLSCertPath string
	TLSKeyPath  string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

// DSN returns the PostgreSQL connection string
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	MaxRetries int
}

// Addr returns Redis address string
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// PayPalConfig holds PayPal configuration
type PayPalConfig struct {
	ClientID     string
	ClientSecret string
	WebhookID    string
	Mode         string
}

// PayHereConfig holds PayHere configuration
type PayHereConfig struct {
	MerchantID     string
	MerchantSecret string
	Mode           string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error)
	_ = godotenv.Load()

	cfg := &Config{
		// Application
		AppEnv:   getEnv("APP_ENV", "development"),
		AppName:  getEnv("APP_NAME", "webhook.inst.lk"),
		AppDebug: getEnvBool("APP_DEBUG", false),

		// Domains
		Domain:         getEnv("DOMAIN", "webhook.inst.lk"),
		APIDomain:      getEnv("API_DOMAIN", "api.webhook.inst.lk"),
		WSDomain:       getEnv("WS_DOMAIN", "ws.webhook.inst.lk"),
		ConsoleDomain:  getEnv("CONSOLE_DOMAIN", "console.webhook.inst.lk"),
		DNSDomainName:  getEnv("DNS_DOMAIN", "dns.webhook.inst.lk"),
		SMTPDomainName: getEnv("SMTP_DOMAIN", "smtp.webhook.inst.lk"),

		// Ports
		FrontendPort: getEnvInt("FRONTEND_PORT", 2100),
		APIPort:      getEnvInt("API_PORT", 2200),
		WSPort:       getEnvInt("WS_PORT", 2300),
		ConsolePort:  getEnvInt("CONSOLE_PORT", 2400),
		DNSPort:      getEnvInt("DNS_PORT", 2500),
		SMTPPort:     getEnvInt("SMTP_PORT", 2600),
		WorkerPort:   getEnvInt("WORKER_PORT", 2700),

		// Database
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "webhook"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", "webhook"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: int32(getEnvInt("POSTGRES_MAX_CONNS", 50)),
			MinConns: int32(getEnvInt("POSTGRES_MIN_CONNS", 5)),
		},

		// Redis
		Redis: RedisConfig{
			Host:       getEnv("REDIS_HOST", "localhost"),
			Port:       getEnvInt("REDIS_PORT", 6379),
			Password:   getEnv("REDIS_PASSWORD", ""),
			DB:         getEnvInt("REDIS_DB", 0),
			MaxRetries: getEnvInt("REDIS_MAX_RETRIES", 3),
		},

		// JWT
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			AccessExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "webhook.inst.lk"),
		},

		// Security
		CSRFSecret:   getEnv("CSRF_SECRET", ""),
		CookieDomain: getEnv("COOKIE_DOMAIN", ".webhook.inst.lk"),
		CookieSecure: getEnvBool("COOKIE_SECURE", true),
		CORSOrigins:  getEnv("CORS_ORIGINS", "https://webhook.inst.lk"),

		// Rate Limiting
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60),

		// Payload
		PayloadMaxFree:    int64(getEnvInt("PAYLOAD_MAX_FREE", 524288)),
		PayloadMaxPremium: int64(getEnvInt("PAYLOAD_MAX_PREMIUM", 5242880)),

		// Retention (hours)
		RetentionFree:       getEnvInt("RETENTION_FREE", 72),
		RetentionPro:        getEnvInt("RETENTION_PRO", 720),
		RetentionTeam:       getEnvInt("RETENTION_TEAM", 2160),
		RetentionEnterprise: getEnvInt("RETENTION_ENTERPRISE", 8760),

		// Monitoring
		MetricsEnabled:   getEnvBool("METRICS_ENABLED", true),
		MetricsPort:      getEnvInt("METRICS_PORT", 2800),
		MetricsPath:      getEnv("METRICS_PATH", "/metrics"),
		MetricsAuthToken: getEnv("METRICS_AUTH_TOKEN", ""),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// Billing
		PayPal: PayPalConfig{
			ClientID:     getEnv("PAYPAL_CLIENT_ID", ""),
			ClientSecret: getEnv("PAYPAL_CLIENT_SECRET", ""),
			WebhookID:    getEnv("PAYPAL_WEBHOOK_ID", ""),
			Mode:         getEnv("PAYPAL_MODE", "sandbox"),
		},
		PayHere: PayHereConfig{
			MerchantID:     getEnv("PAYHERE_MERCHANT_ID", ""),
			MerchantSecret: getEnv("PAYHERE_MERCHANT_SECRET", ""),
			Mode:           getEnv("PAYHERE_MODE", "sandbox"),
		},

		// TLS
		TLSCertPath: getEnv("TLS_CERT_PATH", ""),
		TLSKeyPath:  getEnv("TLS_KEY_PATH", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks required configuration values
func (c *Config) Validate() error {
	if c.JWT.Secret == "" && c.AppEnv == "production" {
		return fmt.Errorf("JWT_SECRET is required in production")
	}
	if c.Database.Password == "" && c.AppEnv == "production" {
		return fmt.Errorf("POSTGRES_PASSWORD is required in production")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
