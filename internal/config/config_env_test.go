package config

import (
	"os"
	"strings"
	"testing"
)

func escreverDotenv(t *testing.T, conteudo string) {
	t.Helper()
	if err := os.WriteFile(".env", []byte(conteudo), 0o600); err != nil {
		t.Fatalf("escrevendo .env: %v", err)
	}
}

// semVariavel remove a variável de ambiente e restaura o valor original ao fim
// do teste. Necessário porque o godotenv não sobrescreve variável já presente,
// mesmo que vazia.
func semVariavel(t *testing.T, nome string) {
	t.Helper()
	antigo, existia := os.LookupEnv(nome)
	if err := os.Unsetenv(nome); err != nil {
		t.Fatalf("removendo %s: %v", nome, err)
	}
	t.Cleanup(func() {
		if existia {
			_ = os.Setenv(nome, antigo)
			return
		}
		_ = os.Unsetenv(nome)
	})
}

// Regressão: sem arquivo .env a aplicação deve subir com as variáveis de ambiente.
func TestLoad_SemDotenv_UsaVariaveisDeAmbiente(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/loja")
	semVariavel(t, "PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load falhou sem .env mesmo com DATABASE_URL definida: %v", err)
	}
	if cfg.DatabaseURL != "postgres://localhost:5432/loja" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.Port != "3000" {
		t.Errorf("Port = %q, esperado 3000 (default)", cfg.Port)
	}
}

func TestLoad_DotenvDefineValores(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\nPORT=4000\n")
	semVariavel(t, "DATABASE_URL")
	semVariavel(t, "PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DatabaseURL != "postgres://localhost:5432/loja" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.Port != "4000" {
		t.Errorf("Port = %q, esperado 4000", cfg.Port)
	}
}

func TestLoad_DatabaseURLAusenteRetornaErro(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "PORT=4000\n")
	semVariavel(t, "DATABASE_URL")
	semVariavel(t, "PORT")

	_, err := Load()
	if err == nil {
		t.Fatal("Load deveria falhar sem DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("erro = %q, deveria mencionar DATABASE_URL", err)
	}
}

func TestLoad_PortInvalidaRetornaErro(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\nPORT=abc\n")
	semVariavel(t, "DATABASE_URL")
	semVariavel(t, "PORT")

	_, err := Load()
	if err == nil {
		t.Fatal("Load deveria falhar com PORT inválida")
	}
	if !strings.Contains(err.Error(), "PORT") {
		t.Errorf("erro = %q, deveria mencionar PORT", err)
	}
}

func TestLoad_PortForaDoIntervaloRetornaErro(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\nPORT=70000\n")
	semVariavel(t, "DATABASE_URL")
	semVariavel(t, "PORT")

	if _, err := Load(); err == nil {
		t.Fatal("Load deveria falhar com PORT fora do intervalo 1-65535")
	}
}

// As cotas têm defaults e aceitam valor explícito.
func TestLoad_CotasDeRateLimite(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\n"+
		"RATE_LIMIT_IP_PER_MIN=30\nRATE_LIMIT_IP_BURST=9\n"+
		"RATE_LIMIT_EMAIL_PER_MIN=7\nRATE_LIMIT_EMAIL_BURST=2\n")
	for _, nome := range []string{"DATABASE_URL", "RATE_LIMIT_IP_PER_MIN", "RATE_LIMIT_IP_BURST", "RATE_LIMIT_EMAIL_PER_MIN", "RATE_LIMIT_EMAIL_BURST"} {
		semVariavel(t, nome)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RateLimitIPPerMin != 30 || cfg.RateLimitIPBurst != 9 ||
		cfg.RateLimitEmailPerMin != 7 || cfg.RateLimitEmailBurst != 2 {
		t.Errorf("cotas = %+v, esperado 30/9/7/2", cfg)
	}
}

func TestLoad_CotasDefault(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\n")
	semVariavel(t, "DATABASE_URL")
	for _, nome := range []string{"RATE_LIMIT_IP_PER_MIN", "RATE_LIMIT_IP_BURST", "RATE_LIMIT_EMAIL_PER_MIN", "RATE_LIMIT_EMAIL_BURST"} {
		semVariavel(t, nome)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RateLimitIPPerMin != defaultIPPerMin || cfg.RateLimitIPBurst != defaultIPBurst ||
		cfg.RateLimitEmailPerMin != defaultEmailPerMin || cfg.RateLimitEmailBurst != defaultEmailBurst {
		t.Errorf("defaults = %+v, esperado %d/%d/%d/%d", cfg,
			defaultIPPerMin, defaultIPBurst, defaultEmailPerMin, defaultEmailBurst)
	}
}

// Cota zero ou negativa derruba o boot: limite desligado silenciosamente é pior
// que subir quebrado.
func TestLoad_CotaZeroOuNegativaRetornaErro(t *testing.T) {
	casos := []struct {
		nome  string
		valor string
	}{
		{"RATE_LIMIT_IP_PER_MIN", "0"},
		{"RATE_LIMIT_IP_PER_MIN", "-1"},
		{"RATE_LIMIT_IP_BURST", "0"},
		{"RATE_LIMIT_EMAIL_PER_MIN", "-5"},
		{"RATE_LIMIT_EMAIL_BURST", "abc"},
	}

	for _, c := range casos {
		t.Run(c.nome+"="+c.valor, func(t *testing.T) {
			t.Chdir(t.TempDir())
			escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\n"+c.nome+"="+c.valor+"\n")
			semVariavel(t, "DATABASE_URL")
			semVariavel(t, c.nome)

			if _, err := Load(); err == nil {
				t.Fatalf("Load deveria falhar com %s=%s", c.nome, c.valor)
			}
		})
	}
}

func TestLoad_ListasDeCORSEProxies(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\n"+
		"CORS_ORIGINS=https://a.com, https://b.com\nTRUSTED_PROXIES=10.0.0.0/8,192.168.0.0/16\n")
	semVariavel(t, "DATABASE_URL")
	semVariavel(t, "CORS_ORIGINS")
	semVariavel(t, "TRUSTED_PROXIES")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[0] != "https://a.com" || cfg.CORSOrigins[1] != "https://b.com" {
		t.Errorf("CORSOrigins = %v", cfg.CORSOrigins)
	}
	if len(cfg.TrustedProxies) != 2 || cfg.TrustedProxies[0] != "10.0.0.0/8" {
		t.Errorf("TrustedProxies = %v", cfg.TrustedProxies)
	}
}

func TestLoad_TrustedProxyInvalidoRetornaErro(t *testing.T) {
	t.Chdir(t.TempDir())
	escreverDotenv(t, "DATABASE_URL=postgres://localhost:5432/loja\nTRUSTED_PROXIES=nao-e-cidr\n")
	semVariavel(t, "DATABASE_URL")
	semVariavel(t, "TRUSTED_PROXIES")

	if _, err := Load(); err == nil {
		t.Fatal("Load deveria falhar com TRUSTED_PROXIES inválido")
	}
}
