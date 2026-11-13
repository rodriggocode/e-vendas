package usecase

import (
	"context"
	"e-vendas/internal/domain"
	"e-vendas/internal/security"
	"net/mail"
	"strings"
)

type UserCase struct {
	repository domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserCase {
	return &UserCase{
		repository: repo,
	}
}

func (us *UserCase) CreateUser(ctx context.Context, name, email, password string) (domain.User, error) {

	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	err := validateRegister(name, email, password)
	if err != nil {
		return domain.User{}, err
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	userID, err := us.repository.CreateUser(ctx, user)

	if err != nil {
		return domain.User{}, err
	}

	user.ID = userID

	return user, nil
}

func validateRegister(name, email, password string) error {
	if name == "" {
		return &domain.ValidationError{Message: "o nome é obrigatório"}
	}

	if email == "" {
		return &domain.ValidationError{Message: "o e-mail e obrigatorio"}
	}

	if len(email) > 254 {
		return &domain.ValidationError{Message: "e-mail muito longo"}
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return &domain.ValidationError{Message: "e-mail inválido"}
	}

	if len(password) < 8 {
		return &domain.ValidationError{Message: "a senha deve ter no mínimo 8 caracteres"}
	}

	if len(password) > 72 {
		return &domain.ValidationError{Message: "a senha deve ter no máximo 72 caracteres"}
	}

	return nil
}
