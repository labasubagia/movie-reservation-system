package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testServer *echo.Echo

func TestMain(m *testing.M) {

	ctx := context.Background()

	config := NewConfig()

	container, err := postgres.Run(
		ctx,
		"postgres:17",
		postgres.WithDatabase(config.PostgresDB),
		postgres.WithUsername(config.PostgresUser),
		postgres.WithPassword(config.PostgresPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	pool, err := NewDBPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	trxProvider := NewTransactionProvider(config, pool)
	handler := NewHandler(config, trxProvider)
	testServer = setupServer(config, handler)

	m.Run()
}
