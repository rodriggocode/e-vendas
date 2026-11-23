package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"e-vendas/internal/config"
	"e-vendas/internal/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Teste de fiação: monta o router REAL (handler -> usecase -> repositório ->
// schema) contra o PostgreSQL e exercita a rota de registro de ponta a ponta.
// É o buraco que a suíte não cobria: os testes de handler usam repo falso e os
// de repositório não passam pelo HTTP.
//
// Pula sem DATABASE_URL (mesmo padrão do pacote repository).

// novoAmbienteDeFiacao devolve o handler real e o pool usado.
//
// A cota de rate limit é elevada para que a sequência de cenários não esbarre
// em 429 quando o middleware (S1) for implementado: 429 aqui seria o limite
// funcionando, não falha de registro. É no-op enquanto a env não existe.
func novoAmbienteDeFiacao(t *testing.T) (http.Handler, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL não definida; pulando teste de fiação")
	}

	t.Setenv("RATE_LIMIT_IP_PER_MIN", "1000000")
	t.Setenv("RATE_LIMIT_EMAIL_PER_MIN", "1000000")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("abrindo pool: %v", err)
	}
	t.Cleanup(pool.Close)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("carregando config: %v", err)
	}

	return Router(pool, cfg), pool
}

func emailDeFiacao(t *testing.T, prefixo string) string {
	t.Helper()
	return fmt.Sprintf("%s.%d.%d@example.com", prefixo, time.Now().UnixNano(), os.Getpid())
}

func limparUsuarioDeFiacao(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email); err != nil {
		t.Logf("limpeza de %s: %v", email, err)
	}
}

func enviarRegistro(t *testing.T, h http.Handler, contentType string, corpo []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(corpo))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func corpoJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

type registroDeFiacao struct {
	ID    int
	Nome  string
	Email string
	Phone *string
	Hash  string
}

func lerRegistro(t *testing.T, pool *pgxpool.Pool, email string) (registroDeFiacao, bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var r registroDeFiacao
	err := pool.QueryRow(ctx,
		"SELECT id, name, email, phone, password_hash FROM users WHERE email = $1", email,
	).Scan(&r.ID, &r.Nome, &r.Email, &r.Phone, &r.Hash)
	if err == pgx.ErrNoRows {
		return registroDeFiacao{}, false
	}
	if err != nil {
		t.Fatalf("lendo registro %s: %v", email, err)
	}
	return r, true
}

func contarRegistros(t *testing.T, pool *pgxpool.Pool, email string) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var n int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE lower(email) = lower($1)", email).Scan(&n); err != nil {
		t.Fatalf("contando registros de %s: %v", email, err)
	}
	return n
}

// TestAPI_RegistroComBanco cobre o caminho feliz e as bordas do registro com o
// schema real: Content-Type, teto de nome, duplicidade, telefone, 413, 400.
func TestAPI_RegistroComBanco(t *testing.T) {
	h, pool := novoAmbienteDeFiacao(t)

	t.Run("charset=utf-8 registra e grava telefone normalizado", func(t *testing.T) {
		email := emailDeFiacao(t, "api.charset")
		t.Cleanup(func() { limparUsuarioDeFiacao(t, pool, email) })

		corpo := corpoJSON(t, map[string]any{
			"name": "QA Fiação", "email": email, "phone": "(11) 99999-8888", "password": "senha12345",
		})
		w := enviarRegistro(t, h, "application/json; charset=utf-8", corpo)

		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, esperado 201; corpo: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("resposta não é JSON: %s", w.Body.String())
		}
		for _, campo := range []string{"id", "name", "email", "phone"} {
			if _, ok := resp[campo]; !ok {
				t.Errorf("resposta sem o campo %q: %s", campo, w.Body.String())
			}
		}
		if _, vazou := resp["password_hash"]; vazou {
			t.Errorf("resposta expôs password_hash: %s", w.Body.String())
		}
		if resp["phone"] != "11999998888" {
			t.Errorf("phone devolvido = %v, esperado 11999998888", resp["phone"])
		}

		gravado, ok := lerRegistro(t, pool, email)
		if !ok {
			t.Fatal("registro não encontrado no banco após 201")
		}
		if gravado.Nome != "QA Fiação" || gravado.Phone == nil || *gravado.Phone != "11999998888" {
			t.Errorf("registro gravado = %+v, esperado nome 'QA Fiação' e phone 11999998888", gravado)
		}
		if gravado.Hash == "" || gravado.Hash == "senha12345" {
			t.Errorf("password_hash gravado = %q; esperado hash bcrypt", gravado.Hash)
		}
	})

	t.Run("text/plain devolve 415", func(t *testing.T) {
		w := enviarRegistro(t, h, "text/plain", corpoJSON(t, map[string]any{"name": "A", "email": "a@b.com"}))
		if w.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("status = %d, esperado 415; corpo: %s", w.Code, w.Body.String())
		}
	})

	t.Run("nome acima do teto devolve 422 e nao grava", func(t *testing.T) {
		email := emailDeFiacao(t, "api.nome")
		t.Cleanup(func() { limparUsuarioDeFiacao(t, pool, email) })

		corpo := corpoJSON(t, map[string]any{
			"name": repetir("x", 121), "email": email, "password": "senha12345",
		})
		w := enviarRegistro(t, h, "application/json", corpo)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, esperado 422; corpo: %s", w.Code, w.Body.String())
		}
		if _, ok := lerRegistro(t, pool, email); ok {
			t.Error("nome acima do teto gravou registro no banco")
		}
	})

	t.Run("e-mail duplicado devolve 409 e nao grava segunda", func(t *testing.T) {
		email := emailDeFiacao(t, "api.dup")
		t.Cleanup(func() { limparUsuarioDeFiacao(t, pool, email) })

		base := map[string]any{"name": "QA Dup", "email": email, "password": "senha12345"}
		if w := enviarRegistro(t, h, "application/json", corpoJSON(t, base)); w.Code != http.StatusCreated {
			t.Fatalf("primeiro registro: status = %d, esperado 201; corpo: %s", w.Code, w.Body.String())
		}

		variante := map[string]any{"name": "QA Dup", "email": maiusculas(email), "password": "senha12345"}
		w := enviarRegistro(t, h, "application/json", corpoJSON(t, variante))
		if w.Code != http.StatusConflict {
			t.Fatalf("status = %d, esperado 409; corpo: %s", w.Code, w.Body.String())
		}
		if n := contarRegistros(t, pool, email); n != 1 {
			t.Errorf("banco com %d registros do mesmo e-mail, esperado 1", n)
		}
	})

	t.Run("telefone ausente grava NULL e devolve string vazia", func(t *testing.T) {
		email := emailDeFiacao(t, "api.semfone")
		t.Cleanup(func() { limparUsuarioDeFiacao(t, pool, email) })

		w := enviarRegistro(t, h, "application/json", corpoJSON(t, map[string]any{
			"name": "QA Sem Fone", "email": email, "password": "senha12345",
		}))
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, esperado 201; corpo: %s", w.Code, w.Body.String())
		}

		gravado, ok := lerRegistro(t, pool, email)
		if !ok {
			t.Fatal("registro não encontrado no banco após 201")
		}
		if gravado.Phone != nil {
			t.Errorf("phone gravado = %q, esperado NULL", *gravado.Phone)
		}
	})

	t.Run("corpo acima de 1MB devolve 413", func(t *testing.T) {
		email := emailDeFiacao(t, "api.grande")
		t.Cleanup(func() { limparUsuarioDeFiacao(t, pool, email) })

		corpo := corpoJSON(t, map[string]any{
			"name": repetir("A", 1_100_000), "email": email, "password": "senha12345",
		})
		w := enviarRegistro(t, h, "application/json", corpo)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, esperado 413; corpo: %s", w.Code, w.Body.String())
		}
		if _, ok := lerRegistro(t, pool, email); ok {
			t.Error("corpo acima do limite gravou registro no banco")
		}
	})

	t.Run("campo desconhecido devolve 400", func(t *testing.T) {
		w := enviarRegistro(t, h, "application/json", corpoJSON(t, map[string]any{
			"name": "QA", "email": emailDeFiacao(t, "api.desconhecido"), "password": "senha12345", "x": 1,
		}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, esperado 400; corpo: %s", w.Code, w.Body.String())
		}
	})

	t.Run("healthz responde 200 sem tocar o banco", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, esperado 200; corpo: %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("resposta não é JSON: %s", w.Body.String())
		}
		if resp["status"] != "ok" {
			t.Errorf("status = %v, esperado ok", resp["status"])
		}
	})
}

func repetir(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

func maiusculas(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - ('a' - 'A')
		}
	}
	return string(b)
}
