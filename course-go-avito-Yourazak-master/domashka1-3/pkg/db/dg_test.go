package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDBConfigFromEnv(t *testing.T) {
	originalVars := map[string]string{
		"POSTGRES_HOST":        os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":        os.Getenv("POSTGRES_PORT"),
		"POSTGRES_USER":        os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD":    os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":          os.Getenv("POSTGRES_DB"),
		"DB_MAX_CONNS":         os.Getenv("DB_MAX_CONNS"),
		"DB_MIN_CONNS":         os.Getenv("DB_MIN_CONNS"),
		"DB_MAX_CONN_LIFETIME": os.Getenv("DB_MAX_CONN_LIFETIME"),
	}

	os.Setenv("POSTGRES_HOST", "test-host")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "test-user")
	os.Setenv("POSTGRES_PASSWORD", "test-pass")
	os.Setenv("POSTGRES_DB", "test-db")
	os.Setenv("DB_MAX_CONNS", "20")
	os.Setenv("DB_MIN_CONNS", "5")
	os.Setenv("DB_MAX_CONN_LIFETIME", "2h")

	defer func() {
		for key, value := range originalVars {
			os.Setenv(key, value)
		}
	}()

	cfg := NewDBConfigFromEnv()

	assert.NotNil(t, cfg)
	assert.Equal(t, "test-host", cfg.Host)
	assert.Equal(t, "5433", cfg.Port)
	assert.Equal(t, "test-user", cfg.User)
	assert.Equal(t, "test-pass", cfg.Password)
	assert.Equal(t, "test-db", cfg.DBName)
	assert.Equal(t, 20, cfg.MaxConns)
	assert.Equal(t, 5, cfg.MinConns)
	assert.Equal(t, 2*time.Hour, cfg.MaxConnLifetime)
}

func TestNewDBConfigFromEnv_DefaultValues(t *testing.T) {
	originalVars := map[string]string{
		"DB_MAX_CONNS":         os.Getenv("DB_MAX_CONNS"),
		"DB_MIN_CONNS":         os.Getenv("DB_MIN_CONNS"),
		"DB_MAX_CONN_LIFETIME": os.Getenv("DB_MAX_CONN_LIFETIME"),
	}

	os.Unsetenv("DB_MAX_CONNS")
	os.Unsetenv("DB_MIN_CONNS")
	os.Unsetenv("DB_MAX_CONN_LIFETIME")

	defer func() {
		for key, value := range originalVars {
			os.Setenv(key, value)
		}
	}()

	cfg := NewDBConfigFromEnv()

	assert.NotNil(t, cfg)
	assert.Equal(t, 10, cfg.MaxConns)
	assert.Equal(t, 1, cfg.MinConns)
	assert.Equal(t, time.Hour, cfg.MaxConnLifetime)
}

func TestNewDBConfigFromEnv_InvalidValues(t *testing.T) {
	originalVars := map[string]string{
		"DB_MAX_CONNS":         os.Getenv("DB_MAX_CONNS"),
		"DB_MIN_CONNS":         os.Getenv("DB_MIN_CONNS"),
		"DB_MAX_CONN_LIFETIME": os.Getenv("DB_MAX_CONN_LIFETIME"),
	}

	os.Setenv("DB_MAX_CONNS", "invalid")
	os.Setenv("DB_MIN_CONNS", "invalid")
	os.Setenv("DB_MAX_CONN_LIFETIME", "invalid")

	defer func() {
		for key, value := range originalVars {
			os.Setenv(key, value)
		}
	}()

	cfg := NewDBConfigFromEnv()

	assert.Equal(t, 10, cfg.MaxConns)
	assert.Equal(t, 1, cfg.MinConns)
	assert.Equal(t, time.Hour, cfg.MaxConnLifetime)
}

func TestInitPool_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	invalidConfig := &DBConfig{
		Host:     "invalid-host",
		Port:     "9999",
		User:     "invalid-user",
		Password: "invalid-pass",
		DBName:   "invalid-db",
	}

	pool, err := InitPool(ctx, invalidConfig, 1)

	assert.Error(t, err)
	assert.Nil(t, pool)
}
