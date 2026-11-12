package main

import (
	"context"
	"e-vendas/internal/config"
	"e-vendas/internal/router"
	"fmt"
	"log"
	"net/http"

	"e-vendas/internal/repository/postgres"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	r := router.Router(db)

	fmt.Printf("Server run: http://localhost:%s/\n", cfg.Port)
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(":%s", cfg.Port),
		r,
	))
}
