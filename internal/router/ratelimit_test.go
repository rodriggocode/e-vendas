package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"e-vendas/internal/config"
	"e-vendas/internal/repository/postgres"
)

// Testes do rate limit de POST /api/v1/auth/register (S1, bloqueador do primeiro
// deploy). Ficam vermelhos enquanto o middleware nao existe: 429 esperado chega
// como 2xx. Quando o @devops implementar seguindo docs/design-rate-limit.md,
// os mesmos asserts passam sem mudar o teste.
//
// As cotas sao baixas de proposito (bursts curtos) para estourar sem esperar a
// janela de reposicao.

func handlerComCota(t *testing.T, ipPorMin, ipBurst, emailPorMin, emailBurst string) http.Handler {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL não definida; pulando teste de rate limit")
	}

	t.Setenv("RATE_LIMIT_IP_PER_MIN", ipPorMin)
	t.Setenv("RATE_LIMIT_IP_BURST", ipBurst)
	t.Setenv("RATE_LIMIT_EMAIL_PER_MIN", emailPorMin)
	t.Setenv("RATE_LIMIT_EMAIL_BURST", emailBurst)
	t.Setenv("TRUSTED_PROXIES", "")

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

	return Router(pool, cfg)
}

func postRegistro(t *testing.T, h http.Handler, remoteAddr, xff, email string) *httptest.ResponseRecorder {
	t.Helper()
	corpo := []byte(`{"name":"QA Rate","email":"` + email + `","password":"senha12345"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(string(corpo)))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func assert429(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, esperado 429; corpo: %s", w.Code, w.Body.String())
	}
	retry := w.Header().Get("Retry-After")
	if retry == "" {
		t.Error("429 sem header Retry-After")
	}
	var corpo map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("corpo não é JSON: %s", w.Body.String())
	}
	errObj, ok := corpo["error"].(map[string]any)
	if !ok || errObj["code"] != "rate_limited" {
		t.Errorf("envelope = %s; esperado error.code = rate_limited", w.Body.String())
	}
}

// TestRateLimit_EstouraCotaDoIP cobre o 429, o envelope e o Retry-After.
func TestRateLimit_EstouraCotaDoIP(t *testing.T) {
	h := handlerComCota(t, "60", "1", "60", "1")

	var bloqueadas, passaram int
	for i := 0; i < 5; i++ {
		w := postRegistro(t, h, "203.0.113.7:4321", "", emailDeFiacao(t, "rl.ip"))
		switch w.Code {
		case http.StatusTooManyRequests:
			bloqueadas++
			assert429(t, w)
		default:
			passaram++
		}
	}
	if bloqueadas == 0 {
		t.Fatal("nenhum 429 após 5 registros do mesmo IP; o rate limit não está aplicado")
	}
	if passaram == 0 {
		t.Error("todas as requisições barraram; o burst deveria admitir ao menos uma")
	}
}

// TestRateLimit_HealthzNuncaLeva429 garante que o healthcheck fica fora da cota.
func TestRateLimit_HealthzNuncaLeva429(t *testing.T) {
	h := handlerComCota(t, "60", "1", "60", "1")

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("healthz recebeu 429 na requisição %d; deveria ficar fora da cota", i+1)
		}
		if w.Code != http.StatusOK {
			t.Fatalf("healthz status = %d, esperado 200", w.Code)
		}
	}
}

// TestRateLimit_DoisIPsNaoColapsam confirma que o balde é por IP.
func TestRateLimit_DoisIPsNaoColapsam(t *testing.T) {
	h := handlerComCota(t, "60", "1", "600", "100")

	for i, ip := range []string{"203.0.113.7:4321", "198.51.100.9:4321"} {
		w := postRegistro(t, h, ip, "", emailDeFiacao(t, "rl.dois"))
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("IP %d recebeu 429 sendo o primeiro do seu balde; baldes colapsando", i+1)
		}
	}
}

// TestRateLimit_XFFForjadoComPeerPublicoEIgnorado e o teste do ataque de burla:
// com TRUSTED_PROXIES vazio, o XFF do cliente não pode trocar de balde.
func TestRateLimit_XFFForjadoComPeerPublicoEIgnorado(t *testing.T) {
	h := handlerComCota(t, "60", "1", "600", "100")

	var bloqueadas int
	for i := 0; i < 5; i++ {
		w := postRegistro(t, h, "203.0.113.7:4321", "198.51.100."+string(rune('1'+i)), emailDeFiacao(t, "rl.xff"))
		if w.Code == http.StatusTooManyRequests {
			bloqueadas++
		}
	}
	if bloqueadas == 0 {
		t.Fatal("XFF rotativo contornou o limite; o peer público deve ignorar XFF")
	}
}

// TestRateLimit_CotaPorEmailIndependeDoIP fecha o buraco do oráculo (S2): o
// mesmo e-mail de IP diferente continua com cota própria.
func TestRateLimit_CotaPorEmailIndependeDoIP(t *testing.T) {
	h := handlerComCota(t, "600", "100", "60", "1")

	email := emailDeFiacao(t, "rl.email")
	w1 := postRegistro(t, h, "203.0.113.7:4321", "", email)
	if w1.Code == http.StatusTooManyRequests {
		t.Fatal("primeira requisição do e-mail já levou 429")
	}
	w2 := postRegistro(t, h, "198.51.100.9:4321", "", email)
	assert429(t, w2)
}
