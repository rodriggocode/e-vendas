package router

import (
	"net"
	"net/http"

	"e-vendas/internal/config"
	"e-vendas/internal/handler"
	"e-vendas/internal/middleware"
	"e-vendas/internal/repository/postgres"
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Router monta a API. Ordem dos middlewares: recover → headers → CORS → mux;
// o rate limit fica só na rota de registro, então /healthz nunca leva 429.
func Router(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := http.NewServeMux()

	userRepository := postgres.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handler.NewUserHandler(userUsecase)

	registro := middleware.RateLimit(middleware.RateLimitConfig{
		IPPerMin:    cfg.RateLimitIPPerMin,
		IPBurst:     cfg.RateLimitIPBurst,
		EmailPerMin: cfg.RateLimitEmailPerMin,
		EmailBurst:  cfg.RateLimitEmailBurst,
		Trusted:     redesConfiaveis(cfg.TrustedProxies),
	}, http.HandlerFunc(userHandler.Register))

	r.Handle("POST /api/v1/auth/register", registro)
	r.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		responses.JSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		})
	})

	var h http.Handler = r
	h = middleware.CORS(cfg.CORSOrigins, h)
	h = middleware.SecurityHeaders(h)
	h = middleware.Recoverer(h)

	return h
}

// redesConfiaveis converte os CIDRs validados em config.Load.
func redesConfiaveis(cidrs []string) []*net.IPNet {
	if len(cidrs) == 0 {
		return nil
	}

	redes := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, rede, err := net.ParseCIDR(cidr)
		if err != nil {
			continue // config.Load já barrou CIDR inválido no boot
		}
		redes = append(redes, rede)
	}

	return redes
}
