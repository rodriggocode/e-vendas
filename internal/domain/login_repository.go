package domain

import "context"

type LoginRepository interface{
    Login(ctx context.Context, email string) (User, error)
}
