package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("Lendo .env: %w", err)
		}

		log.Println("Sem .env, usando variaveis de ambiente")
	}

	dns := os.Getenv("DATABASE_URL")
	if dns == "" {
		return nil, fmt.Errorf("DATABASE_URL nao definida")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("PORT invalide: %q", port)
	}

	return &Config{
		DatabaseURL: dns,
		Port:        port,
	}, nil
}
