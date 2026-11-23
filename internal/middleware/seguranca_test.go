package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func rede(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, rede, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR(%q): %v", cidr, err)
	}
	return rede
}

// XFF só é honrado quando o peer está nos proxies confiáveis.
func TestClientIP_ConfiavelHonraXFF(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.4:5555"
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.4")

	if got := ClientIP(r, []*net.IPNet{rede(t, "10.0.0.0/8")}); got != "203.0.113.9" {
		t.Fatalf("ClientIP = %q, esperado 203.0.113.9 (XFF de proxy confiável)", got)
	}
}

func TestClientIP_PeerForaDosConfiaveisIgnoraXFF(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.7:5555"
	r.Header.Set("X-Forwarded-For", "198.51.100.1")

	if got := ClientIP(r, []*net.IPNet{rede(t, "10.0.0.0/8")}); got != "203.0.113.7" {
		t.Fatalf("ClientIP = %q, esperado 203.0.113.7 (peer público não confia em XFF)", got)
	}
}

// CORS fail-closed: só a origem da lista recebe ACAO.
func TestCORS_OrigemPermitidaRecebeACAO(t *testing.T) {
	h := CORS([]string{"https://a.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Origin", "https://a.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://a.com" {
		t.Errorf("ACAO = %q, esperado https://a.com", got)
	}
}

func TestCORS_OrigemNaoPermitidaNaoRecebeACAO(t *testing.T) {
	h := CORS([]string{"https://a.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Origin", "https://maliciosa.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO = %q para origem não listada; esperado vazio (fail-closed)", got)
	}
}

func TestCORS_ListaVaziaNegaTudo(t *testing.T) {
	h := CORS(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Origin", "https://a.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO = %q com CORS_ORIGINS vazio; esperado vazio", got)
	}
}

func TestSecurityHeaders_Presentes(t *testing.T) {
	h := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, header := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"Referrer-Policy",
	} {
		if w.Header().Get(header) == "" {
			t.Errorf("header %s ausente", header)
		}
	}
}

func TestRecoverer_ConvertePanicEm500(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("falha de teste")
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, esperado 500; corpo: %s", w.Code, w.Body.String())
	}
}
