package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func NewDBPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
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
