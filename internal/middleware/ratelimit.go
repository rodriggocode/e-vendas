package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"e-vendas/internal/responses"
)

// maxCorpoParaEmail é o mesmo teto do handler (1 MB). Corpo maior não é lido
// aqui, para não consumir o body antes do MaxBytesReader e perder o 413.
const maxCorpoParaEmail = 1 << 20

type RateLimitConfig struct {
	IPPerMin    int
	IPBurst     int
	EmailPerMin int
	EmailBurst  int
	Trusted     []*net.IPNet
}

type entry struct {
	tokens float64
	last   time.Time
}

// memStore é um token bucket em memória, por instância. GC preguiçoso a cada
// acessosLimpeza chamadas evita crescimento sem fim (sem goroutine vazando).
type memStore struct {
	mu      sync.Mutex
	entries map[string]*entry
	acessos int
}

func newMemStore() *memStore {
	return &memStore{entries: make(map[string]*entry)}
}

func (s *memStore) allow(key string, porMin, burst int) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agora := time.Now()
	taxaPorSegundo := float64(porMin) / 60.0

	e, ok := s.entries[key]
	if !ok {
		e = &entry{tokens: float64(burst), last: agora}
		s.entries[key] = e
	}

	e.tokens += agora.Sub(e.last).Seconds() * taxaPorSegundo
	if e.tokens > float64(burst) {
		e.tokens = float64(burst)
	}
	e.last = agora

	if e.tokens >= 1 {
		e.tokens--
		s.acessos++
		if s.acessos%1000 == 0 {
			s.limparOciosos(agora)
		}
		return true, 0
	}

	faltam := (1 - e.tokens) / taxaPorSegundo
	return false, time.Duration(math.Ceil(faltam)) * time.Second
}

func (s *memStore) limparOciosos(agora time.Time) {
	for k, e := range s.entries {
		if agora.Sub(e.last) > 10*time.Minute {
			delete(s.entries, k)
		}
	}
}

// ClientIP devolve o IP do cliente. X-Forwarded-For só é honrado quando o peer
// está nos proxies confiáveis; fora deles, o attacker escolheria o próprio
// balde e contornaria o limite.
func ClientIP(r *http.Request, trusted []*net.IPNet) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "unknown"
	}

	for _, rede := range trusted {
		if rede.Contains(ip) {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				if primeiro := strings.TrimSpace(strings.Split(xff, ",")[0]); primeiro != "" {
					if real := net.ParseIP(primeiro); real != nil {
						return real.String()
					}
				}
			}
			break
		}
	}

	return ip.String()
}

// RateLimit aplica cota por IP e por e-mail no registro. A chave entra no
// store como hash SHA-256 para não reter PII em memória/log.
func RateLimit(cfg RateLimitConfig, next http.Handler) http.Handler {
	store := newMemStore()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r, cfg.Trusted)

		if permitido, retry := store.allow("ip:"+hashKey(ip), cfg.IPPerMin, cfg.IPBurst); !permitido {
			responderRateLimit(w, retry)
			return
		}

		email := emailDoPedido(r)
		if email != "" {
			if permitido, retry := store.allow("email:"+hashKey(email), cfg.EmailPerMin, cfg.EmailBurst); !permitido {
				responderRateLimit(w, retry)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// emailDoPedido extrai o e-mail do corpo sem consumi-lo. Corpos acima do teto
// não são lidos (o handler precisa vê-los inteiros para responder 413).
func emailDoPedido(r *http.Request) string {
	if r.ContentLength < 0 || r.ContentLength > maxCorpoParaEmail {
		return ""
	}

	corpo, err := io.ReadAll(io.LimitReader(r.Body, maxCorpoParaEmail+1))
	if err != nil {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewReader(corpo))

	var req struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(corpo, &req); err != nil {
		return ""
	}

	return normalizarEmail(req.Email)
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashKey(valor string) string {
	sum := sha256.Sum256([]byte(valor))
	return hex.EncodeToString(sum[:])
}

func responderRateLimit(w http.ResponseWriter, retry time.Duration) {
	if retry < time.Second {
		retry = time.Second
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
	responses.ResponseError(w, http.StatusTooManyRequests, "rate_limited", "muitas requisições, tente novamente em instantes")
}
