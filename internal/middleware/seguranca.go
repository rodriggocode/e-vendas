package middleware

import (
	"log"
	"net/http"

	"e-vendas/internal/responses"
)

// SecurityHeaders adiciona os cabeçalhos de segurança em todas as respostas.
// A API não serve HTML; o CSP é restritivo justamente para não haver superfície
// de script no mesmo origin.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")

		next.ServeHTTP(w, r)
	})
}

// CORS libera apenas as origens configuradas. Lista vazia (default) nega tudo:
// a API é same-origin e não há motivo para aceitar navegador de outro origin.
func CORS(origins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false

		if origin != "" {
			for _, o := range origins {
				if o == origin {
					allowed = true
					w.Header().Set("Access-Control-Allow-Origin", o)
					w.Header().Set("Vary", "Origin")
					break
				}
			}
		}

		if r.Method == http.MethodOptions {
			// Para origem não autorizada responde 204 sem CORS nenhum (fail-closed).
			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Recoverer converte panic em 500 com o envelope padrão, sem derrubar o processo.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recuperado := recover(); recuperado != nil {
				log.Printf("panic recuperado: %v", recuperado)
				responses.ResponseError(w, http.StatusInternalServerError, "internal_error", "erro interno")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
