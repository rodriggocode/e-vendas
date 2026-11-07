package main

import (
	"e-vendas/internal/config"
	"e-vendas/internal/router"
	"fmt"
	"log"
	"net/http"
)

func main() {

	config.Load()

	r := router.Router()

	fmt.Println("Server run: http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", r))
}
