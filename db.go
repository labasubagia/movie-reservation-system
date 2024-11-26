package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func NewDBPool(ctx context.Context, c *Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)
	err = migrate(db)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
