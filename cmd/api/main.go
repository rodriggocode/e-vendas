package main

import (
	"context"
	"e-vendas/internal/config"
	"e-vendas/internal/repository/postgres"
	"e-vendas/internal/router"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	r := router.Router(db, cfg)

	srv := &http.Server{
		Handler:           r,
		Addr:              ":" + cfg.Port,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("Server run: http://localhost:%s/\n", cfg.Port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	sigCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-sigCtx.Done()
	log.Println("shutting down gracefully...")

	defer stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
