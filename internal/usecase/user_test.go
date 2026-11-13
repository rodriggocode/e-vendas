package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"e-vendas/internal/domain"
	"e-vendas/internal/security"
)

type fakeUserRepository struct {
	created     domain.User
	id          int
	err         error
	called      bool
	receivedCtx context.Context
}

func (f *fakeUserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	f.called = true
	f.receivedCtx = ctx
	f.created = user
	return f.id, f.err
}

func TestCreateUser_HashesSenhaAntesDeSalvar(t *testing.T) {
	repo := &fakeUserRepository{id: 42}
	uc := NewUserUsecase(repo)

	criado, err := uc.CreateUser(context.Background(), "João", "joao@example.com", "senha-plana")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if !repo.called {
		t.Fatal("repositório não foi chamado")
	}
	if criado.ID != 42 {
		t.Errorf("ID = %d, esperado 42", criado.ID)
	}
	if repo.created.Name != "João" || repo.created.Email != "joao@example.com" {
		t.Errorf("usuário enviado = %+v", repo.created)
	}
	if repo.created.PasswordHash == "senha-plana" {
		t.Fatal("senha foi gravada em texto plano")
	}
	if err := security.CheckPassword(repo.created.PasswordHash, "senha-plana"); err != nil {
		t.Errorf("hash gravada não corresponde à senha informada: %v", err)
	}
}

func TestCreateUser_PropagaErroDoRepositorio(t *testing.T) {
	repo := &fakeUserRepository{err: domain.ErrConflict}
	uc := NewUserUsecase(repo)

	_, err := uc.CreateUser(context.Background(), "João", "joao@example.com", "senha12345")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("erro = %v, esperado domain.ErrConflict", err)
	}
}

func TestCreateUser_PassaContextoParaRepositorio(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "valor")

	repo := &fakeUserRepository{id: 1}
	uc := NewUserUsecase(repo)

	if _, err := uc.CreateUser(ctx, "João", "joao@example.com", "senha12345"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if repo.receivedCtx.Value(ctxKey{}) != "valor" {
		t.Error("contexto não foi repassado ao repositório")
	}
}

// Regressão: a validação de cadastro deve ser aplicada antes de gravar.
func TestCreateUser_RejeitaEntradaInvalida(t *testing.T) {
	casos := []struct {
		nome            string
		nomeIn, emailIn string
		senhaIn         string
	}{
		{"nome vazio", "", "joao@example.com", "senha12345"},
		{"email vazio", "João", "", "senha12345"},
		{"email inválido", "João", "não-e-mail", "senha12345"},
		{"senha curta", "João", "joao@example.com", "1234567"},
		{"senha longa", "João", "joao@example.com", strings.Repeat("a", 73)},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			repo := &fakeUserRepository{id: 1}
			uc := NewUserUsecase(repo)

			_, err := uc.CreateUser(context.Background(), c.nomeIn, c.emailIn, c.senhaIn)
			if err == nil {
				t.Fatalf("CreateUser aceitou entrada inválida: nome=%q email=%q senha=%q", c.nomeIn, c.emailIn, c.senhaIn)
			}
			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Errorf("erro = %v, esperado *domain.ValidationError", err)
			}
			if repo.called {
				t.Error("repositório não deveria ser chamado com dados inválidos")
			}
		})
	}
}

// A função existe e está correta; o problema é que ninguém a chama.
func TestValidateRegister_RejeitaEntradaInvalida(t *testing.T) {
	casos := []struct {
		nome            string
		nomeIn, emailIn string
		senhaIn         string
	}{
		{"nome vazio", "", "joao@example.com", "senha12345"},
		{"email vazio", "João", "", "senha12345"},
		{"email inválido", "João", "não-e-mail", "senha12345"},
		{"senha curta", "João", "joao@example.com", "1234567"},
		{"senha longa", "João", "joao@example.com", strings.Repeat("a", 73)},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			var validation *domain.ValidationError
			if err := validateRegister(c.nomeIn, c.emailIn, c.senhaIn); !errors.As(err, &validation) {
				t.Errorf("validateRegister(%q, %q, %q) = %v, esperado *domain.ValidationError", c.nomeIn, c.emailIn, c.senhaIn, err)
			}
		})
	}
}

func TestValidateRegister_AceitaEntradaValida(t *testing.T) {
	if err := validateRegister("João", "joao@example.com", "senha12345"); err != nil {
		t.Errorf("validateRegister rejeitou entrada válida: %v", err)
	}
}
