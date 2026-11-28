package middleware

import (
	"e-vendas/internal/responses"
	"e-vendas/internal/security"
	"log"
	"net/http"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s", r.Method, r.RequestURI, r.Host)
		next(w, r)
	}
}

func Authentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := security.ValidateToken(r); err != nil {
			responses.ResponseError(w, http.StatusUnauthorized, "unauthorized", "token inválido ou ausente")
			return
		}

		next(w, r)
	}
}
