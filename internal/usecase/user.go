package usecase

import (
	"context"
	"e-vendas/internal/domain"
	"e-vendas/internal/security"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// nameMaxLength acompanha o padrão de produto (ARQUITETURA.md 5.1) e fica
// abaixo do limite físico da coluna users.name (VARCHAR(255)).
const nameMaxLength = 120

type UserCase struct {
	repository domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserCase {
	return &UserCase{
		repository: repo,
	}
}

func (us *UserCase) CreateUser(ctx context.Context, name, email, phone, password string) (domain.User, error) {

	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	originalPhone := strings.TrimSpace(phone)
	phone = onlyDigits(originalPhone)

	if originalPhone != "" && phone == "" {
		return domain.User{}, &domain.ValidationError{Message: "telefone inválido"}
	}

	err := validateRegister(name, email, phone, password)
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
		Phone:        phone,
		PasswordHash: string(hash),
	}

	userID, err := us.repository.CreateUser(ctx, user)

	if err != nil {
		return domain.User{}, err
	}

	user.ID = userID

	return user, nil
}

func validateRegister(name, email, phone, password string) error {
	if name == "" {
		return &domain.ValidationError{Message: "o nome é obrigatório"}
	}

	if utf8.RuneCountInString(name) > nameMaxLength {
		return &domain.ValidationError{Message: "o nome deve ter no máximo 120 caracteres"}
	}

	if email == "" {
		return &domain.ValidationError{Message: "o e-mail e obrigatorio"}
	}

	if len(email) > 254 {
		return &domain.ValidationError{Message: "e-mail muito longo"}
	}

	if phone != "" && (len(phone) < 10 || len(phone) > 13) {
		return &domain.ValidationError{Message: "telefone inválido"}
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

func onlyDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}
