package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var SecretKey []byte

type Config struct {
	DatabaseURL string
	Port        string

	// Rate limit de POST /api/v1/auth/register (S1).
	RateLimitIPPerMin    int
	RateLimitIPBurst     int
	RateLimitEmailPerMin int
	RateLimitEmailBurst  int

	// CORS fail-closed e proxies confiáveis para X-Forwarded-For.
	CORSOrigins    []string
	TrustedProxies []string
}

// Defaults do rate limit: burst menor que a taxa, para negar rajada.
const (
	defaultIPPerMin    = 10
	defaultIPBurst     = 5
	defaultEmailPerMin = 5
	defaultEmailBurst  = 3
)

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("lendo .env: %w", err)
		}

		log.Println("Sem .env, usando variaveis de ambiente")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL nao definida")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	SecretKey = []byte(os.Getenv("SECRET_KEY"))

	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("PORT invalida: %q", port)
	}

	ipPerMin, err := inteiroPositivo("RATE_LIMIT_IP_PER_MIN", defaultIPPerMin)
	if err != nil {
		return nil, err
	}

	ipBurst, err := inteiroPositivo("RATE_LIMIT_IP_BURST", defaultIPBurst)
	if err != nil {
		return nil, err
	}

	emailPerMin, err := inteiroPositivo("RATE_LIMIT_EMAIL_PER_MIN", defaultEmailPerMin)
	if err != nil {
		return nil, err
	}

	emailBurst, err := inteiroPositivo("RATE_LIMIT_EMAIL_BURST", defaultEmailBurst)
	if err != nil {
		return nil, err
	}

	trusted := listaDeEnv("TRUSTED_PROXIES")
	for _, cidr := range trusted {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return nil, fmt.Errorf("TRUSTED_PROXIES invalida: %q", cidr)
		}
	}

	return &Config{
		DatabaseURL:          dsn,
		Port:                 port,
		RateLimitIPPerMin:    ipPerMin,
		RateLimitIPBurst:     ipBurst,
		RateLimitEmailPerMin: emailPerMin,
		RateLimitEmailBurst:  emailBurst,
		CORSOrigins:          listaDeEnv("CORS_ORIGINS"),
		TrustedProxies:       trusted,
	}, nil
}

// inteiroPositivo falha no boot com valor ausente/zero/negativo: limite
// "desligado silenciosamente" é pior que subir quebrado.
func inteiroPositivo(nome string, padrao int) (int, error) {
	valor := os.Getenv(nome)
	if valor == "" {
		return padrao, nil
	}

	n, err := strconv.Atoi(valor)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s invalida: %q", nome, valor)
	}

	return n, nil
}

// listaDeEnv separa por vírgula, remove espaços e ignora itens vazios.
func listaDeEnv(nome string) []string {
	var itens []string

	for _, item := range strings.Split(os.Getenv(nome), ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			itens = append(itens, item)
		}
	}

	return itens
}
