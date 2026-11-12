package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"e-vendas/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func novaPoolDeTeste(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL não definida; pulando teste de integração")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("abrindo pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func emailUnico(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("qa.%d@example.com", time.Now().UnixNano())
}

func limparUsuario(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email); err != nil {
		t.Logf("limpeza do usuário %s: %v", email, err)
	}
}

func TestCreateUser_InsereERetornaID(t *testing.T) {
	pool := novaPoolDeTeste(t)
	repo := NewUserRepository(pool)

	email := emailUnico(t)
	t.Cleanup(func() { limparUsuario(t, pool, email) })

	user := domain.User{Name: "QA Integração", Email: email, PasswordHash: "hash-fake"}
	id, err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if id <= 0 {
		t.Fatalf("ID = %d, esperado > 0", id)
	}

	var name, emailDB, hashDB string
	err = pool.QueryRow(context.Background(),
		"SELECT name, email, password_hash FROM users WHERE id = $1", id).
		Scan(&name, &emailDB, &hashDB)
	if err != nil {
		t.Fatalf("lendo usuário inserido: %v", err)
	}
	if name != user.Name || emailDB != email {
		t.Errorf("registro = (%q, %q), esperado (%q, %q)", name, emailDB, user.Name, email)
	}
	if hashDB != "hash-fake" {
		t.Errorf("password_hash = %q; o repositório deve gravar o hash que recebe", hashDB)
	}
}

func TestCreateUser_EmailDuplicadoRetornaErro(t *testing.T) {
	pool := novaPoolDeTeste(t)
	repo := NewUserRepository(pool)

	email := emailUnico(t)
	t.Cleanup(func() { limparUsuario(t, pool, email) })

	if _, err := repo.CreateUser(context.Background(), domain.User{Name: "A", Email: email, PasswordHash: "x"}); err != nil {
		t.Fatalf("primeira inserção: %v", err)
	}
	if _, err := repo.CreateUser(context.Background(), domain.User{Name: "B", Email: email, PasswordHash: "x"}); err == nil {
		t.Fatal("segunda inserção com o mesmo email deveria falhar")
	}
}

func TestCreateUser_GravaPayloadDeSQLInjectionLiteralmente(t *testing.T) {
	pool := novaPoolDeTeste(t)
	repo := NewUserRepository(pool)

	email := emailUnico(t)
	t.Cleanup(func() { limparUsuario(t, pool, email) })

	nome := "Robert'); DROP TABLE users; --"
	if _, err := repo.CreateUser(context.Background(), domain.User{Name: nome, Email: email, PasswordHash: "x"}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	var nomeDB string
	if err := pool.QueryRow(context.Background(), "SELECT name FROM users WHERE email = $1", email).Scan(&nomeDB); err != nil {
		t.Fatalf("lendo usuário: %v", err)
	}
	if nomeDB != nome {
		t.Errorf("name gravado = %q, esperado %q", nomeDB, nome)
	}
}

func TestCreateUser_ContextoCanceladoRetornaErro(t *testing.T) {
	pool := novaPoolDeTeste(t)
	repo := NewUserRepository(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	email := emailUnico(t)
	if _, err := repo.CreateUser(ctx, domain.User{Name: "QA", Email: email, PasswordHash: "x"}); err == nil {
		t.Fatal("CreateUser deveria falhar com o contexto cancelado")
	}
}

func TestSchema_UsersTemColunasEsperadas(t *testing.T) {
	pool := novaPoolDeTeste(t)

	colunas := []string{"id", "name", "email", "password_hash", "active", "is_platform_admin", "email_verified_at", "created_at", "updated_at"}
	for _, coluna := range colunas {
		var existe bool
		err := pool.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = $1
			)`, coluna).Scan(&existe)
		if err != nil {
			t.Fatalf("verificando coluna %s: %v", coluna, err)
		}
		if !existe {
			t.Errorf("coluna users.%s não existe", coluna)
		}
	}
}
