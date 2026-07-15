package config

import "testing"

var configEnvironmentKeys = []string{
	"APP_ENV",
	"HTTP_ADDR",
	"SHUTDOWN_TIMEOUT",
	"DATABASE_URL",
	"DATABASE_MAX_CONNECTIONS",
	"DATABASE_MIN_CONNECTIONS",
	"DATABASE_CONNECT_TIMEOUT",
	"DATABASE_HEALTH_CHECK_PERIOD",
}

func TestLoadRejectsMinConnectionsAboveMax(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("DATABASE_MIN_CONNECTIONS", "10")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "5")

	if _, err := Load(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadRejectsNonPositiveConnectTimeout(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "0s")

	if _, err := Load(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadRejectsNonPositiveShutdownTimeout(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("SHUTDOWN_TIMEOUT", "-1s")

	if _, err := Load(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadUsesDatabaseDefaults(t *testing.T) {
	clearConfigEnvironment(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Database.URL == "" {
		t.Fatal("database URL must not be empty")
	}
	if cfg.Database.MaxConnections <= 0 {
		t.Fatal("maximum connection count must be positive")
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()

	for _, key := range configEnvironmentKeys {
		t.Setenv(key, "")
	}
}
