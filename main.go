package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	config := NewConfig()

	pool, err := NewDBPool(ctx, config.PostgresDSN())
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
