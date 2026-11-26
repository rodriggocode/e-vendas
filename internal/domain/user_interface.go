package domain

import "context"

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (int, error)
  GetUsers(ctx context.Context) ([]User, error)
}

type LoginRepository interface {
   Login(ctx context.Context, email string) (User, error)
}