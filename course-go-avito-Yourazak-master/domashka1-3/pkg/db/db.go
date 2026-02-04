package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func InitPool(ctx context.Context, cfg *DBConfig, retryAttempts int) (*pgxpool.Pool, error) {

	_ = godotenv.Load()

	host := cfg.Host
	if host == "" {
		host = os.Getenv("POSTGRES_HOST")
	}
	port := cfg.Port
	if port == "" {
		port = os.Getenv("POSTGRES_PORT")
	}
	user := cfg.User
	if user == "" {
		user = os.Getenv("POSTGRES_USER")
	}
	pass := cfg.Password
	if pass == "" {
		pass = os.Getenv("POSTGRES_PASSWORD")
	}
	dbname := cfg.DBName
	if dbname == "" {
		dbname = os.Getenv("POSTGRES_DB")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, dbname)

	cfgPool := &pgxpool.Config{}
	parsed, err := pgxpool.ParseConfig(dsn)
	if err == nil {
		cfgPool = parsed
	} else {
		// fallback: try New
		cfgPool, _ = pgxpool.ParseConfig(dsn)
	}

	if cfg.MaxConns > 0 {
		cfgPool.MaxConns = int32(cfg.MaxConns)
	}
	if cfg.MinConns >= 0 {
		cfgPool.MinConns = int32(cfg.MinConns)
	}
	if cfg.MaxConnLifetime > 0 {
		cfgPool.MaxConnLifetime = cfg.MaxConnLifetime
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfgPool)
	if err != nil {
		return nil, err
	}

	attempts := retryAttempts
	if attempts <= 0 {
		attempts = 5
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		lastErr = pool.Ping(pingCtx)
		cancel()
		if lastErr == nil {
			log.Println("Connected to Postgres")
			return pool, nil
		}
		sleep := time.Duration(500*(i+1)) * time.Millisecond
		log.Printf("DB ping failed, attempt %d/%d: %v — retry in %v\n", i+1, attempts, lastErr, sleep)
		time.Sleep(sleep)
	}

	pool.Close()
	return nil, lastErr
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func NewDBConfigFromEnv() *DBConfig {
	_ = godotenv.Load()
	maxConns := 10
	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxConns = n
		}
	}
	minConns := 1
	if v := os.Getenv("DB_MIN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			minConns = n
		}
	}
	maxLifetime := time.Hour
	if v := os.Getenv("DB_MAX_CONN_LIFETIME"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			maxLifetime = d
		}
	}

	return &DBConfig{
		Host:            os.Getenv("POSTGRES_HOST"),
		Port:            os.Getenv("POSTGRES_PORT"),
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		DBName:          os.Getenv("POSTGRES_DB"),
		MaxConns:        maxConns,
		MinConns:        minConns,
		MaxConnLifetime: maxLifetime,
	}
}
