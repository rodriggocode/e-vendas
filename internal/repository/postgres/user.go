package postgres

import (
	"context"
	"e-vendas/internal/domain"
	"errors"

	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrConflict
		}
		return 0, fmt.Errorf("criando usuário: %w", err)
	}

	return id, nil

}
