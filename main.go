package main

import (
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	config := NewConfig()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDB,
	)
	pool, err := NewDBPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	trxProvider := NewTransactionProvider(config, pool)
	handler := NewHandler(config, trxProvider)

	err = RunServer(config, handler)
	if err != nil {
		log.Fatal(err)
	}
}
