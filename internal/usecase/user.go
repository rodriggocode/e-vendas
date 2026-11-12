package usecase

import (
	"context"
	"e-vendas/internal/domain"
	"e-vendas/internal/security"
)

type UserCase struct {
	repository domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserCase {
	return &UserCase{
		repository: repo,
	}
}

func (us *UserCase) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {

	hash, err := security.HashPassword(user.PasswordHash)
	if err != nil {
		return domain.User{}, err
	}

	user.PasswordHash = string(hash)

	userID, err := us.repository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	user.ID = userID

	return user, nil
}
