package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv   string
	HTTPPort string
	LogLevel string
	// SwaggerEnabled — default aktif kecuali APP_ENV=production.
	SwaggerEnabled bool
	Database       DatabaseConfig
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// Load membaca konfigurasi dari environment variable, dengan default untuk local dev.
func Load() (*Config, error) {
	maxOpen, err := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	maxIdle, err := getEnvInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	lifetime, err := getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return nil, err
	}

	appEnv := getEnv("APP_ENV", "development")
	swaggerEnabled, err := getEnvBool("SWAGGER_ENABLED", appEnv != "production")
	if err != nil {
		return nil, err
	}

	return &Config{
		AppEnv:         appEnv,
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		SwaggerEnabled: swaggerEnabled,
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "pokemon"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    maxOpen,
			MaxIdleConns:    maxIdle,
			ConnMaxLifetime: lifetime,
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("env %s harus integer: %w", key, err)
	}
	return n, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("env %s harus durasi (contoh: 30m): %w", key, err)
	}
	return d, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("env %s harus boolean: %w", key, err)
	}
	return b, nil
}
