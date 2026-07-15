package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultEnvironment               = "local"
	defaultHTTPAddress               = ":8080"
	defaultShutdownTimeout           = 10 * time.Second
	defaultDatabaseURL               = "postgres://employee_requests:employee_requests@localhost:5432/employee_requests?sslmode=disable"
	defaultDatabaseMaxConnections    = int32(20)
	defaultDatabaseMinConnections    = int32(2)
	defaultDatabaseConnectTimeout    = 5 * time.Second
	defaultDatabaseHealthCheckPeriod = 30 * time.Second
)

type Config struct {
	Environment     string
	HTTPAddress     string
	ShutdownTimeout time.Duration
	Database        DatabaseConfig
}

type DatabaseConfig struct {
	URL               string
	MaxConnections    int32
	MinConnections    int32
	ConnectTimeout    time.Duration
	HealthCheckPeriod time.Duration
}

func Load() (Config, error) {
	shutdownTimeout, err := positiveDurationFromEnv("SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	maxConnections, err := positiveInt32FromEnv("DATABASE_MAX_CONNECTIONS", defaultDatabaseMaxConnections)
	if err != nil {
		return Config{}, err
	}

	minConnections, err := nonNegativeInt32FromEnv("DATABASE_MIN_CONNECTIONS", defaultDatabaseMinConnections)
	if err != nil {
		return Config{}, err
	}

	if minConnections > maxConnections {
		return Config{}, fmt.Errorf("DATABASE_MIN_CONNECTIONS must not exceed DATABASE_MAX_CONNECTIONS")
	}

	connectTimeout, err := positiveDurationFromEnv("DATABASE_CONNECT_TIMEOUT", defaultDatabaseConnectTimeout)
	if err != nil {
		return Config{}, err
	}

	healthCheckPeriod, err := positiveDurationFromEnv(
		"DATABASE_HEALTH_CHECK_PERIOD",
		defaultDatabaseHealthCheckPeriod,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:     stringFromEnv("APP_ENV", defaultEnvironment),
		HTTPAddress:     stringFromEnv("HTTP_ADDR", defaultHTTPAddress),
		ShutdownTimeout: shutdownTimeout,
		Database: DatabaseConfig{
			URL:               stringFromEnv("DATABASE_URL", defaultDatabaseURL),
			MaxConnections:    maxConnections,
			MinConnections:    minConnections,
			ConnectTimeout:    connectTimeout,
			HealthCheckPeriod: healthCheckPeriod,
		},
	}, nil
}

func stringFromEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return duration, nil
}

func positiveDurationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	duration, err := durationFromEnv(key, fallback)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return duration, nil
}

func positiveInt32FromEnv(key string, fallback int32) (int32, error) {
	value, err := int32FromEnv(key, fallback)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return value, nil
}

func nonNegativeInt32FromEnv(key string, fallback int32) (int32, error) {
	value, err := int32FromEnv(key, fallback)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must not be negative", key)
	}

	return value, nil
}

func int32FromEnv(key string, fallback int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return int32(parsed), nil
}
