package postgres

import (
	"context"
	"e-vendas/internal/domain"
	"errors"

	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (repo *UserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	sql := `INSERT INTO users(name, email,phone,password_hash)
		VALUES ($1,$2, $3, $4) RETURNING id
	`

	var phone pgtype.Text
	if user.Phone != "" {
		phone.Valid = true
		phone.String = user.Phone
	}

	var id int

	err := repo.db.QueryRow(ctx, sql, user.Name, user.Email, phone, user.PasswordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return 0, domain.ErrConflict
			case "22001":
				// valor excede o tamanho da coluna: erro de entrada, não interno
				return 0, &domain.ValidationError{Message: "um campo excede o tamanho máximo permitido"}
			}
		}
		return 0, fmt.Errorf("criando usuário: %w", err)
	}

	return id, nil

}
