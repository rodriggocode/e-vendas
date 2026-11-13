package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Router(nil) é seguro aqui: estas rotas não encostam no banco.
func TestRouter_RaizNaoExisteRetorna404(t *testing.T) {
	w := httptest.NewRecorder()
	Router(nil).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado 404; corpo: %s", w.Code, w.Body.String())
	}
}

// Regressão: ARQUITETURA.md P0 — /healthz deve existir.
func TestRouter_HealthzRetorna200JSON(t *testing.T) {
	w := httptest.NewRecorder()
	Router(nil).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200; corpo: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, esperado application/json; corpo: %s", ct, w.Body.String())
	}
}

func TestRouter_RotaDesconhecidaRetorna404(t *testing.T) {
	w := httptest.NewRecorder()
	Router(nil).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/rota/qualquer", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado 404; corpo: %s", w.Code, w.Body.String())
	}
}

func TestRouter_MetodoNaoPermitidoRetorna405(t *testing.T) {
	w := httptest.NewRecorder()
	Router(nil).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, esperado 405; corpo: %s", w.Code, w.Body.String())
	}
}
