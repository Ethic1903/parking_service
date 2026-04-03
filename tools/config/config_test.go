package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("CONFIG_FILE", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "")
	t.Setenv("HTTP_WRITE_TIMEOUT_SEC", "")
	t.Setenv("HTTP_IDLE_TIMEOUT_SEC", "")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT_SEC", "")
	t.Setenv("DB_DRIVER", "")
	t.Setenv("DB_POSTGRES_HOST", "")
	t.Setenv("DB_POSTGRES_PORT", "")
	t.Setenv("DB_POSTGRES_DBNAME", "")
	t.Setenv("DB_POSTGRES_USER", "")
	t.Setenv("DB_POSTGRES_PASSWORD", "")
	t.Setenv("DB_POSTGRES_SSLMODE", "")
	t.Setenv("DB_SQLITE_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != "dev" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "dev")
	}
	if cfg.HTTP.Port != "8080" {
		t.Fatalf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "8080")
	}
	if cfg.HTTP.ReadTimeout != 5*time.Second {
		t.Fatalf("ReadTimeout = %s, want %s", cfg.HTTP.ReadTimeout, 5*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 10*time.Second {
		t.Fatalf("WriteTimeout = %s, want %s", cfg.HTTP.WriteTimeout, 10*time.Second)
	}
	if cfg.HTTP.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want %s", cfg.HTTP.IdleTimeout, 60*time.Second)
	}
	if cfg.HTTP.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want %s", cfg.HTTP.ShutdownTimeout, 10*time.Second)
	}
	if cfg.Storage.Driver != "sqlite" {
		t.Fatalf("Storage.Driver = %q, want %q", cfg.Storage.Driver, "sqlite")
	}
	if cfg.Storage.SQLite.Path != "parking.db" {
		t.Fatalf("SQLite.Path = %q, want %q", cfg.Storage.SQLite.Path, "parking.db")
	}
}

func TestLoadEnvOverridesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "custom.yaml")
	cfgContent := []byte("app:\n  env: file-env\nhttp:\n  port: \"9090\"\n  read_timeout_sec: 11\n  write_timeout_sec: 12\n  idle_timeout_sec: 13\n  shutdown_timeout_sec: 14\ndb:\n  driver: sqlite\n  sqlite:\n    path: \"/tmp/parking.sqlite\"\n")
	if err := os.WriteFile(cfgPath, cfgContent, 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	t.Setenv("CONFIG_FILE", cfgPath)
	t.Setenv("APP_ENV", "staging")
	t.Setenv("HTTP_PORT", "7070")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "3")
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_POSTGRES_HOST", "db-service")
	t.Setenv("DB_POSTGRES_PORT", "5433")
	t.Setenv("DB_POSTGRES_DBNAME", "parking_test")
	t.Setenv("DB_POSTGRES_USER", "tester")
	t.Setenv("DB_POSTGRES_PASSWORD", "secret")
	t.Setenv("DB_POSTGRES_SSLMODE", "require")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != "staging" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "staging")
	}
	if cfg.HTTP.Port != "7070" {
		t.Fatalf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "7070")
	}
	if cfg.HTTP.ReadTimeout != 3*time.Second {
		t.Fatalf("ReadTimeout = %s, want %s", cfg.HTTP.ReadTimeout, 3*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 12*time.Second {
		t.Fatalf("WriteTimeout = %s, want %s", cfg.HTTP.WriteTimeout, 12*time.Second)
	}
	if cfg.HTTP.IdleTimeout != 13*time.Second {
		t.Fatalf("IdleTimeout = %s, want %s", cfg.HTTP.IdleTimeout, 13*time.Second)
	}
	if cfg.HTTP.ShutdownTimeout != 14*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want %s", cfg.HTTP.ShutdownTimeout, 14*time.Second)
	}
	if cfg.Storage.Driver != "postgres" {
		t.Fatalf("Storage.Driver = %q, want %q", cfg.Storage.Driver, "postgres")
	}
	if cfg.Storage.Postgres.Host != "db-service" {
		t.Fatalf("Postgres.Host = %q, want %q", cfg.Storage.Postgres.Host, "db-service")
	}
	if cfg.Storage.Postgres.Port != "5433" {
		t.Fatalf("Postgres.Port = %q, want %q", cfg.Storage.Postgres.Port, "5433")
	}
	if cfg.Storage.Postgres.DBName != "parking_test" {
		t.Fatalf("Postgres.DBName = %q, want %q", cfg.Storage.Postgres.DBName, "parking_test")
	}
	if cfg.Storage.Postgres.User != "tester" {
		t.Fatalf("Postgres.User = %q, want %q", cfg.Storage.Postgres.User, "tester")
	}
	if cfg.Storage.Postgres.Password != "secret" {
		t.Fatalf("Postgres.Password = %q, want %q", cfg.Storage.Postgres.Password, "secret")
	}
	if cfg.Storage.Postgres.SSLMode != "require" {
		t.Fatalf("Postgres.SSLMode = %q, want %q", cfg.Storage.Postgres.SSLMode, "require")
	}
}

func TestLoadFailsWhenExplicitConfigMissing(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.yaml")
	t.Setenv("CONFIG_FILE", missingPath)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadFallsBackForInvalidTimeoutValues(t *testing.T) {
	t.Setenv("CONFIG_FILE", "")
	t.Setenv("HTTP_READ_TIMEOUT_SEC", "-1")
	t.Setenv("HTTP_WRITE_TIMEOUT_SEC", "wrong")
	t.Setenv("HTTP_IDLE_TIMEOUT_SEC", "0")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT_SEC", "-100")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTP.ReadTimeout != 5*time.Second {
		t.Fatalf("ReadTimeout = %s, want %s", cfg.HTTP.ReadTimeout, 5*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 10*time.Second {
		t.Fatalf("WriteTimeout = %s, want %s", cfg.HTTP.WriteTimeout, 10*time.Second)
	}
	if cfg.HTTP.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want %s", cfg.HTTP.IdleTimeout, 60*time.Second)
	}
	if cfg.HTTP.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want %s", cfg.HTTP.ShutdownTimeout, 10*time.Second)
	}
}
