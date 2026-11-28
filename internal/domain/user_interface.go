package domain

import "context"

type UserRepository interface {
  CreateUser(ctx context.Context, user User) (int, error)
  GetUsers(ctx context.Context) ([]User, error)
}


