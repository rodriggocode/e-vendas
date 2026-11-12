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

	entrada := domain.User{Name: "João", Email: "joao@example.com", PasswordHash: "senha-plana"}
	criado, err := uc.CreateUser(context.Background(), entrada)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if !repo.called {
		t.Fatal("repositório não foi chamado")
	}
	if criado.ID != 42 {
		t.Errorf("ID = %d, esperado 42", criado.ID)
	}
	if repo.created.PasswordHash == "senha-plana" {
		t.Fatal("senha foi gravada em texto plano")
	}
	if err := security.CheckPassword(repo.created.PasswordHash, "senha-plana"); err != nil {
		t.Errorf("hash gravada não corresponde à senha informada: %v", err)
	}
}

func TestCreateUser_PropagaErroDoRepositorio(t *testing.T) {
	repo := &fakeUserRepository{err: errors.New("falha no banco")}
	uc := NewUserUsecase(repo)

	if _, err := uc.CreateUser(context.Background(), domain.User{PasswordHash: "x"}); err == nil {
		t.Fatal("CreateUser deveria propagar o erro do repositório")
	}
}

func TestCreateUser_ErroNoHashNaoChamaRepositorio(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	uc := NewUserUsecase(repo)

	longa := strings.Repeat("a", 100)
	if _, err := uc.CreateUser(context.Background(), domain.User{PasswordHash: longa}); err == nil {
		t.Fatal("CreateUser deveria falhar com senha longa demais")
	}
	if repo.called {
		t.Fatal("repositório não deveria ser chamado quando o hash falha")
	}
}

func TestCreateUser_PassaContextoParaRepositorio(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "valor")

	repo := &fakeUserRepository{id: 1}
	uc := NewUserUsecase(repo)

	if _, err := uc.CreateUser(ctx, domain.User{PasswordHash: "x"}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if repo.receivedCtx.Value(ctxKey{}) != "valor" {
		t.Error("contexto não foi repassado ao repositório")
	}
}
