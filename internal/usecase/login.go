package usecase

import(
    "context"
    "fmt"
    "errors"
    "e-vendas/internal/domain"
    "e-vendas/internal/security"
)

var(
    ErrInvalidCredentials = errors.New("email ou senha invalidos")
    ErrInactiveUser = errors.New("usuario invativo")
)

type LoginUseCase struct{
    repository domain.LoginRepository
}

func NewLoginUseCase(repository domain.LoginRepository) *LoginUseCase{
    return &LoginUseCase{repository: repository}
}

func(us *LoginUseCase)Login(ctx context.Context, email, password string)(domain.User, string,error){
   user, err := us.repository.Login(ctx, email)
   if err != nil{
       return domain.User{}, "", ErrInvalidCredentials
   }

   err = security.CheckPassword(user.PasswordHash, password)
   if err != nil{
       return domain.User{}, "", ErrInvalidCredentials
   }
   if !user.Active{
       return domain.User{}, "", ErrInactiveUser
   }

   token, err := security.GenerationToken(user.ID)
   if err != nil{
       return domain.User{}, "", fmt.Errorf("gerar token: %w", err)
   }

   return user, token, nil
}
