package router

import (
	"e-vendas/internal/handler"
	"e-vendas/internal/repository/postgres"
	"e-vendas/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

func Router(db *pgxpool.Pool) http.Handler {
	r := http.NewServeMux()

	userRepository := postgres.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handler.NewUserHandler(userUsecase)

	r.HandleFunc("POST /api/v1/auth/register", userHandler.Register)

	return r
}
