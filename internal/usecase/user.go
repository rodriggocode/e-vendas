package usecase

import (
	"context"
	"e-vendas/internal/domain"
	"e-vendas/internal/security"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const nameMaxLength = 120

type UserCase struct {
	repository domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserCase {
	return &UserCase{
		repository: repo,
	}
}

type LoginUsecase struct{
  login domain.LoginRepository
}

func NewLogin(login domain.LoginRepository) *LoginUsecase{
  return &LoginUsecase{login: login}
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

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

func (us *UserCase) GetUsers(ctx context.Context) ([]domain.User, error) {
	return us.repository.GetUsers(ctx)
}


func(us *LoginUsecase) Login(ctx context.Context, email, password string)(domain.User, error){
  user, err := us.login.Login(ctx, email)
  if err != nil{
    return  domain.User{}, domain.ErrNotFound
  }
  if err := security.CheckPassword(user.PasswordHash, password); err != nil{
    return  domain.User{}, domain.ErrNotFound
  }

  user.PasswordHash = ""

  return user, nil
}

func onlyDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

