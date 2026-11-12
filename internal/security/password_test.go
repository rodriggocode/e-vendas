package security

import (
	"strings"
	"testing"
)

func TestHashPassword_NaoDevolveTextoPlano(t *testing.T) {
	const senha = "senha-secreta-123"

	hash, err := HashPassword(senha)
	if err != nil {
		t.Fatalf("HashPassword retornou erro inesperado: %v", err)
	}
	if string(hash) == senha {
		t.Fatal("hash gerada é idêntica à senha em texto plano")
	}
}

func TestHashPassword_UsaSaltAleatorio(t *testing.T) {
	const senha = "senha-secreta-123"

	primeira, err := HashPassword(senha)
	if err != nil {
		t.Fatalf("primeira HashPassword: %v", err)
	}
	segunda, err := HashPassword(senha)
	if err != nil {
		t.Fatalf("segunda HashPassword: %v", err)
	}
	if string(primeira) == string(segunda) {
		t.Fatal("hashes da mesma senha deveriam ser diferentes (salt aleatório)")
	}
}

func TestCheckPassword_AceitaSenhaCorreta(t *testing.T) {
	const senha = "senha-secreta-123"

	hash, err := HashPassword(senha)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := CheckPassword(string(hash), senha); err != nil {
		t.Fatalf("CheckPassword rejeitou a senha correta: %v", err)
	}
}

func TestCheckPassword_RejeitaSenhaErrada(t *testing.T) {
	hash, err := HashPassword("senha-secreta-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := CheckPassword(string(hash), "senha-errada"); err == nil {
		t.Fatal("CheckPassword aceitou uma senha errada")
	}
}

func TestCheckPassword_RejeitaHashInvalida(t *testing.T) {
	if err := CheckPassword("valor-que-nao-e-hash", "qualquer-senha"); err == nil {
		t.Fatal("CheckPassword aceitou um hash inválido")
	}
}

func TestHashPassword_SenhaComMaisDe72BytesRetornaErro(t *testing.T) {
	longa := strings.Repeat("a", 100)
	if _, err := HashPassword(longa); err == nil {
		t.Fatal("HashPassword deveria falhar para senha com mais de 72 bytes")
	}
}
