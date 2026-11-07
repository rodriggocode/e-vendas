package router

import (
	"net/http"
)

func Router() http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ola mundo"))
	})

	return r
}
