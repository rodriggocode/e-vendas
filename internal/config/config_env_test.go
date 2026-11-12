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
