package postgres

import (
	"context"
	"e-vendas/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func (r UserRepository) CreateUser(ctx context.Context)
