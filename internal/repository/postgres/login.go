package postgres

import (
	"context"
	"e-vendas/internal/domain"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LoginRepository struct {
	db *pgxpool.Pool
}

func NewLoginRepository(db *pgxpool.Pool) *LoginRepository {
	return &LoginRepository{db: db}
}

func (repo *LoginRepository) Login(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT id, name, email, password_hash, active, is_platform_admin
	   FROM users WHERE lower(email) = lower($1)
	`

	var user domain.User

	err := repo.db.QueryRow(ctx, query, email).Scan(
	    &user.ID,
	    &user.Name,
	    &user.Email,
	    &user.PasswordHash,
	    &user.Active,
	    &user.IsPlatformAdmin,

	 )
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("buscando usuário para login: %w", err)
	}

	return user, nil
}
