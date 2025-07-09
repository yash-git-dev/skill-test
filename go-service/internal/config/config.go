package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server  ServerConfig
	NodeJS  NodeJSConfig
	Report  ReportConfig
	Logging LoggingConfig
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NodeJSConfig contains configuration for Node.js API client
type NodeJSConfig struct {
	BaseURL       string        `env:"NODEJS_API_URL" default:"http://localhost:5007/api/v1"`
	Timeout       time.Duration `env:"NODEJS_TIMEOUT" default:"30s"`
	RetryAttempts int           `env:"NODEJS_RETRY_ATTEMPTS" default:"3"`
	RetryDelay    time.Duration `env:"NODEJS_RETRY_DELAY" default:"1s"`

	// Authentication for service-to-service communication
	ServiceUsername string `env:"NODEJS_SERVICE_USERNAME" default:"admin@school-admin.com"`
	ServicePassword string `env:"NODEJS_SERVICE_PASSWORD" default:"3OU4zn3q6Zh9"`
}

// ReportConfig contains PDF report generation configuration
type ReportConfig struct {
	OutputDir     string
	MaxFileSize   int64
	Cleanup       bool
	CleanupAfter  time.Duration
	WatermarkText string
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("GO_SERVICE_PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		},
		NodeJS: NodeJSConfig{
			BaseURL:         getEnv("NODEJS_API_URL", "http://localhost:5007/api/v1"),
			Timeout:         getDurationEnv("NODEJS_TIMEOUT", 30*time.Second),
			RetryAttempts:   getIntEnv("NODEJS_RETRY_ATTEMPTS", 3),
			RetryDelay:      getDurationEnv("NODEJS_RETRY_DELAY", 1*time.Second),
			ServiceUsername: getEnv("NODEJS_SERVICE_USERNAME", "admin@school-admin.com"),
			ServicePassword: getEnv("NODEJS_SERVICE_PASSWORD", "3OU4zn3q6Zh9"),
		},
		Report: ReportConfig{
			OutputDir:     getEnv("REPORT_OUTPUT_DIR", "./reports"),
			MaxFileSize:   getInt64Env("REPORT_MAX_FILE_SIZE", 10*1024*1024), // 10MB
			Cleanup:       getBoolEnv("REPORT_CLEANUP", true),
			CleanupAfter:  getDurationEnv("REPORT_CLEANUP_AFTER", 24*time.Hour),
			WatermarkText: getEnv("REPORT_WATERMARK", "Student Management System - Confidential"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here if needed
	return nil
}
