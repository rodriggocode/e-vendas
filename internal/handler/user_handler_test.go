package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"e-vendas/internal/domain"
	"e-vendas/internal/usecase"
)

const rotaRegister = "/api/v1/auth/register"

type fakeUserRepository struct {
	created domain.User
	id      int
	err     error
	called  bool
}

func (f *fakeUserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	f.called = true
	f.created = user
	return f.id, f.err
}

func novoHandlerDeTeste(t *testing.T, repo *fakeUserRepository) http.Handler {
	t.Helper()
	uc := usecase.NewUserUsecase(repo)
	h := NewUserHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST "+rotaRegister, h.Register)
	return mux
}

func fazerRequest(t *testing.T, h http.Handler, method, corpo string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, rotaRegister, strings.NewReader(corpo))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

// Regressão: dto.go define o contrato de resposta (id, name, email).
// A entidade não pode vazar campos internos.
func TestRegister_RegistroValidoRetornaDTO(t *testing.T) {
	repo := &fakeUserRepository{id: 7}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado 201; corpo: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, esperado application/json", ct)
	}

	var dto RegisterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &dto); err != nil {
		t.Fatalf("corpo não é JSON válido: %v", err)
	}
	if dto.ID != 7 || dto.Name != "João" || dto.Email != "joao@example.com" {
		t.Errorf("DTO = %+v", dto)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("corpo não é JSON válido: %v", err)
	}
	for _, campo := range []string{"active", "is_platform_admin", "email_verified_at", "created_at", "updated_at", "password_hash"} {
		if _, ok := body[campo]; ok {
			t.Errorf("resposta expôs campo interno %q: %s", campo, w.Body.String())
		}
	}
}

func TestRegister_NaoExpoeSenha(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha-super-secreta"}`)

	if strings.Contains(w.Body.String(), "senha-super-secreta") {
		t.Errorf("resposta expõe a senha: %s", w.Body.String())
	}
}

func TestRegister_EnviaDadosParaUsecase(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	_ = fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}`)

	if !repo.called {
		t.Fatal("repositório não foi chamado")
	}
	if repo.created.Name != "João" || repo.created.Email != "joao@example.com" {
		t.Errorf("dados enviados = %+v", repo.created)
	}
	if repo.created.PasswordHash == "senha12345" {
		t.Error("senha chegou ao repositório em texto plano")
	}
}

// Regressão: ARQUITETURA.md 6.3 — JSON malformado ou campo desconhecido → 400.
func TestRegister_JSONInvalidoRetorna400(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{json quebrado`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado 400; corpo: %s", w.Code, w.Body.String())
	}
}

func TestRegister_CorpoVazioRetorna400(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, ``)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado 400; corpo: %s", w.Code, w.Body.String())
	}
}

// Regressão: ARQUITETURA.md 6.3/6.6 — validação de formato na borda → 422.
func TestRegister_CamposObrigatoriosRetornam422(t *testing.T) {
	casos := []struct {
		nome  string
		corpo string
	}{
		{"sem name", `{"email":"joao@example.com","password":"senha12345"}`},
		{"sem email", `{"name":"João","password":"senha12345"}`},
		{"sem password", `{"name":"João","email":"joao@example.com"}`},
		{"name vazio", `{"name":"","email":"joao@example.com","password":"senha12345"}`},
		{"email vazio", `{"name":"João","email":"","password":"senha12345"}`},
		{"password vazio", `{"name":"João","email":"joao@example.com","password":""}`},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			repo := &fakeUserRepository{id: 1}
			h := novoHandlerDeTeste(t, repo)

			w := fazerRequest(t, h, http.MethodPost, c.corpo)

			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, esperado 422; corpo: %s", w.Code, w.Body.String())
			}
			if repo.called {
				t.Error("repositório não deveria ser chamado com dados inválidos")
			}
		})
	}
}

func TestRegister_EmailInvalidoRetorna422(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"não-e-mail","password":"senha12345"}`)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422; corpo: %s", w.Code, w.Body.String())
	}
}

// Regressão: ARQUITETURA.md 6.3 — conflito (ex.: e-mail duplicado) → 409.
func TestRegister_EmailDuplicadoRetorna409(t *testing.T) {
	repo := &fakeUserRepository{err: domain.ErrConflict}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, esperado 409; corpo: %s", w.Code, w.Body.String())
	}
}

// Regressão: ValidationError do domínio deve virar 422, não 500.
func TestRegister_ValidationErrorRetorna422(t *testing.T) {
	repo := &fakeUserRepository{err: &domain.ValidationError{Message: "o nome é obrigatório"}}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}`)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422; corpo: %s", w.Code, w.Body.String())
	}
}

func TestRegister_ErroInternoRetorna500(t *testing.T) {
	repo := &fakeUserRepository{err: errors.New("falha inesperada")}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, esperado 500; corpo: %s", w.Code, w.Body.String())
	}
}

// Regressão: erro interno não pode vazar detalhe técnico para o cliente.
func TestRegister_ErroInternoNaoVazaDetalhe(t *testing.T) {
	repo := &fakeUserRepository{err: errors.New("falha inesperada no banco")}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}`)

	if strings.Contains(w.Body.String(), "falha inesperada") {
		t.Errorf("resposta vazou erro interno: %s", w.Body.String())
	}
}

// Regressão: ARQUITETURA.md 6.6 — rejeitar campos desconhecidos.
func TestRegister_CampoDesconhecidoRetorna400(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345","id":999}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado 400; corpo: %s", w.Code, w.Body.String())
	}
}

func TestRegister_JSONComLixoRetorna400(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodPost, `{"name":"João","email":"joao@example.com","password":"senha12345"}lixo`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado 400; corpo: %s", w.Code, w.Body.String())
	}
}

func TestRegister_SenhaMuitoLongaRetorna422(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	corpo := `{"name":"João","email":"joao@example.com","password":"` + strings.Repeat("a", 100) + `"}`
	w := fazerRequest(t, h, http.MethodPost, corpo)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422; corpo: %s", w.Code, w.Body.String())
	}
}

func TestRegister_MetodoErradoRetorna405(t *testing.T) {
	repo := &fakeUserRepository{id: 1}
	h := novoHandlerDeTeste(t, repo)

	w := fazerRequest(t, h, http.MethodGet, ``)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, esperado 405; corpo: %s", w.Code, w.Body.String())
	}
}
