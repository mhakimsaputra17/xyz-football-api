package config

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration values.
type Config struct {
	App    AppConfig
	DB     DBConfig
	JWT    JWTConfig
	Server ServerConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name string
	Env  string // development, staging, production
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

// JWTConfig holds JWT token settings.
type JWTConfig struct {
	Secret            string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Load reads configuration from .env file and environment variables.
// Environment variables take precedence over .env file values.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Read .env file; not an error if it doesn't exist (env vars may be set directly)
	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("failed to read .env file, relying on environment variables", "error", err)
	}

	// Set defaults
	viper.SetDefault("APP_NAME", "xyz-football-api")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_TIMEZONE", "UTC")
	viper.SetDefault("JWT_ACCESS_EXPIRATION_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_EXPIRATION_DAYS", 7)
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_READ_TIMEOUT_SECONDS", 10)
	viper.SetDefault("SERVER_WRITE_TIMEOUT_SECONDS", 10)

	cfg := &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Env:  viper.GetString("APP_ENV"),
		},
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
			TimeZone: viper.GetString("DB_TIMEZONE"),
		},
		JWT: JWTConfig{
			Secret:            viper.GetString("JWT_SECRET"),
			AccessExpiration:  time.Duration(viper.GetInt("JWT_ACCESS_EXPIRATION_MINUTES")) * time.Minute,
			RefreshExpiration: time.Duration(viper.GetInt("JWT_REFRESH_EXPIRATION_DAYS")) * 24 * time.Hour,
		},
		Server: ServerConfig{
			Port:         viper.GetString("SERVER_PORT"),
			ReadTimeout:  time.Duration(viper.GetInt("SERVER_READ_TIMEOUT_SECONDS")) * time.Second,
			WriteTimeout: time.Duration(viper.GetInt("SERVER_WRITE_TIMEOUT_SECONDS")) * time.Second,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *DBConfig) DSN() string {
	return "host=" + c.Host +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Name +
		" port=" + c.Port +
		" sslmode=" + c.SSLMode +
		" TimeZone=" + c.TimeZone
}

// validate checks that all required configuration values are present.
func (c *Config) validate() error {
	required := map[string]string{
		"DB_USER":     c.DB.User,
		"DB_PASSWORD": c.DB.Password,
		"DB_NAME":     c.DB.Name,
		"JWT_SECRET":  c.JWT.Secret,
	}

	for key, val := range required {
		if val == "" {
			return &ConfigError{Field: key, Message: "is required but not set"}
		}
	}

	return nil
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config: " + e.Field + " " + e.Message
}
