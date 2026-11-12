package router

import (
	"e-vendas/internal/handler"
	repository "e-vendas/internal/repository/postgres"
	usercase "e-vendas/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

func Router(db *pgxpool.Pool) http.Handler {
	r := http.NewServeMux()

	userRepository := repository.NewUserRepository(db)
	userUsecase := usercase.NewUserUsecase(userRepository)
	userHandler := handler.NewUserHandler(userUsecase)

	r.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ola mundo"))
	})

	r.HandleFunc("POST /api/v1/users", userHandler.Register)

	return r
}
