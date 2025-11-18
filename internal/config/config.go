package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment string
	LogLevel    string
	MaxRequests int
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("config validation error [%s]: %s", e.Field, e.Message)
}

// Load reads and validates configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnvWithDefault("SERVER_HOST", "localhost"),
			Port:         getEnvAsIntWithDefault("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDurationWithDefault("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDurationWithDefault("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvAsDurationWithDefault("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Path:            getEnvWithDefault("DB_PATH", "./useragent.db"),
			MaxOpenConns:    getEnvAsIntWithDefault("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsIntWithDefault("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDurationWithDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		App: AppConfig{
			Environment: getEnvWithDefault("APP_ENV", "development"),
			LogLevel:    getEnvWithDefault("LOG_LEVEL", "info"),
			MaxRequests: getEnvAsIntWithDefault("MAX_REQUESTS_PER_MINUTE", 100),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return ValidationError{
			Field:   "SERVER_PORT",
			Message: fmt.Sprintf("port must be between 1 and 65535, got %d", c.Server.Port),
		}
	}

	if c.Server.ReadTimeout <= 0 {
		return ValidationError{
			Field:   "SERVER_READ_TIMEOUT",
			Message: "read timeout must be positive",
		}
	}

	if c.Server.WriteTimeout <= 0 {
		return ValidationError{
			Field:   "SERVER_WRITE_TIMEOUT",
			Message: "write timeout must be positive",
		}
	}

	// Validate database config
	if c.Database.Path == "" {
		return ValidationError{
			Field:   "DB_PATH",
			Message: "database path cannot be empty",
		}
	}

	if c.Database.MaxOpenConns < 1 {
		return ValidationError{
			Field:   "DB_MAX_OPEN_CONNS",
			Message: "max open connections must be at least 1",
		}
	}

	if c.Database.MaxIdleConns < 0 {
		return ValidationError{
			Field:   "DB_MAX_IDLE_CONNS",
			Message: "max idle connections cannot be negative",
		}
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return ValidationError{
			Field:   "DB_MAX_IDLE_CONNS",
			Message: "max idle connections cannot exceed max open connections",
		}
	}

	// Validate app config
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvs[c.App.Environment] {
		return ValidationError{
			Field:   "APP_ENV",
			Message: fmt.Sprintf("environment must be one of: development, staging, production; got %s", c.App.Environment),
		}
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[strings.ToLower(c.App.LogLevel)] {
		return ValidationError{
			Field:   "LOG_LEVEL",
			Message: fmt.Sprintf("log level must be one of: debug, info, warn, error; got %s", c.App.LogLevel),
		}
	}

	if c.App.MaxRequests < 1 {
		return ValidationError{
			Field:   "MAX_REQUESTS_PER_MINUTE",
			Message: "max requests per minute must be at least 1",
		}
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// Helper functions for environment variables

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntWithDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// Return default on parse error
		return defaultValue
	}

	return value
}

func getEnvAsDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		// Return default on parse error
		return defaultValue
	}

	return value
}
