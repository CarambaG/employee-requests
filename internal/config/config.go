package config

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultEnvironment     = "local"
	defaultHTTPAddress     = ":8080"
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	Environment     string
	HTTPAddress     string
	ShutdownTimeout time.Duration
	DatabaseURL     string
}

func Load() (Config, error) {
	shutdownTimeout, err := durationFromEnv("SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:     stringFromEnv("APP_ENV", defaultEnvironment),
		HTTPAddress:     stringFromEnv("HTTP_ADDR", defaultHTTPAddress),
		ShutdownTimeout: shutdownTimeout,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
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
