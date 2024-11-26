package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTransactionProvider(config *Config, db *pgxpool.Pool) *TransactionProvider {
	return &TransactionProvider{
		config: config,
		db:     db,
	}
}

type TransactionProvider struct {
	config *Config
	db     *pgxpool.Pool
}

func (t *TransactionProvider) Transact(ctx context.Context, fn func(service *ServiceRegistry) error) error {
	return runInTx(ctx, t.db, func(tx pgx.Tx) error {
		repository := NewRepositoryRegistry(tx)
		service := NewService(t.config, repository)
		return fn(service)
	})
}

func runInTx(ctx context.Context, db *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		return tx.Commit(ctx)
	}

	rollbackErr := tx.Rollback(ctx)
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
