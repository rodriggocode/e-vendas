package postgres

import (
	"context"
	"e-vendas/internal/domain"

	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (repo *UserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	sql := `INSERT INTO users(name, email,password_hash)
		VALUES ($1,$2, $3) RETURNING id
	`

	var id int

	err := repo.db.QueryRow(ctx, sql, user.Name, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("criando usuário: %w", err)
	}

	return id, nil

}
