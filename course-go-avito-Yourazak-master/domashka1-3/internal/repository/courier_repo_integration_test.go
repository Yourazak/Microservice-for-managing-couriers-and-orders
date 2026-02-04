package repository

import (
	"context"
	"testing"
	"time"

	"avito-courier/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	t.Cleanup(func() {
		postgresContainer.Terminate(ctx)
	})

	host, err := postgresContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	connStr := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	time.Sleep(3 * time.Second)

	pool, err := pgxpool.New(ctx, connStr)
	assert.NoError(t, err)

	err = runTestMigrations(pool)
	assert.NoError(t, err)

	return pool
}

func runTestMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS couriers (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            phone TEXT UNIQUE NOT NULL,
            status TEXT NOT NULL,
            transport_type TEXT NOT NULL, -- Убедись что NOT NULL
            created_at TIMESTAMP DEFAULT NOW(),
            updated_at TIMESTAMP DEFAULT NOW()
        );
        
        CREATE TABLE IF NOT EXISTS delivery (
            id SERIAL PRIMARY KEY,
            courier_id INTEGER REFERENCES couriers(id),
            order_id TEXT UNIQUE NOT NULL,
            assigned_at TIMESTAMP NOT NULL,
            deadline TIMESTAMP NOT NULL
        );
    `)
	return err
}

func TestCourierRepository_Integration_Create_Courier_with_Duplicate_Phone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewCourierRepository(pool)

	courier := &model.Courier{
		Name:          "Duplicate Phone",
		Phone:         "+76666666666",
		Status:        "available",
		TransportType: "car",
	}

	err := repo.Create(context.Background(), courier)
	assert.NoError(t, err)

	duplicateCourier := &model.Courier{
		Name:          "Duplicate",
		Phone:         "+76666666666",
		Status:        "available",
		TransportType: "bike",
	}

	err = repo.Create(context.Background(), duplicateCourier)

	assert.Error(t, err)
	t.Logf("Expected duplicate phone error: %v", err)
}
