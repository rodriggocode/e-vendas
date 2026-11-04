package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("Ola Mundo"))
	})

	fmt.Println("Server run: http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", r))
}
