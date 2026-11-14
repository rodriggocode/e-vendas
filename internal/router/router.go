package router

import (
	"e-vendas/internal/handler"
	"e-vendas/internal/repository/postgres"
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Router(db *pgxpool.Pool) http.Handler {
	r := http.NewServeMux()

	userRepository := postgres.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handler.NewUserHandler(userUsecase)

	r.HandleFunc("POST /api/v1/auth/register", userHandler.Register)
	r.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		responses.JSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		})
	})

	return r
}
