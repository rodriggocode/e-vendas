package router

import (
	"e-vendas/internal/config"
	"e-vendas/internal/handler"
	"e-vendas/internal/repository/postgres"
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Router monta a API. Ordem dos middlewares: recover → headers → CORS → mux;
// o rate limit fica só na rota de registro, então /healthz nunca leva 429.
func Router(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := http.NewServeMux()

	userRepository := postgres.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handler.NewUserHandler(userUsecase)
	loginRepsitory := postgres.NewLoginRepository(db)
	loginUsecase := usecase.NewLoginUseCase(loginRepsitory)
	loginHandler := handler.NewLoginHandler(loginUsecase)

	r.HandleFunc("POST /api/v1/auth/register", userHandler.Register)
	r.HandleFunc("GET /api/v1/users", userHandler.GetUsers)
	r.HandleFunc("POST /api/v1/login", loginHandler.Login)
	r.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		responses.JSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		})
	})

	return r
}
