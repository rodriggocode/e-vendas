package main

import (
	"context"
	"e-vendas/internal/config"
	"e-vendas/internal/router"
	"fmt"
	"log"
	"net/http"
	"time"

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

	srv := &http.Server{
		Handler:           r,
		Addr:              ":" + cfg.Port,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("Server run: http://localhost:%s/\n", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
