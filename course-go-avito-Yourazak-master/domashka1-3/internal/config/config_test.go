package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	originalPort := os.Getenv("PORT")
	originalHost := os.Getenv("POSTGRES_HOST")
	originalDB := os.Getenv("POSTGRES_DB")

	os.Setenv("PORT", "9090")
	os.Setenv("POSTGRES_HOST", "test-host")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "test-user")
	os.Setenv("POSTGRES_PASSWORD", "test-pass")
	os.Setenv("POSTGRES_DB", "test-db")

	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("POSTGRES_HOST", originalHost)
		os.Setenv("POSTGRES_DB", originalDB)
	}()

	cfg := LoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "test-host", cfg.DB.Host)
	assert.Equal(t, "5433", cfg.DB.Port)
	assert.Equal(t, "test-user", cfg.DB.User)
	assert.Equal(t, "test-pass", cfg.DB.Password)
	assert.Equal(t, "test-db", cfg.DB.Name)
}
