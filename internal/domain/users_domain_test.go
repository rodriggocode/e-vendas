package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUser_JSONNaoExpoePasswordHash(t *testing.T) {
	u := User{Name: "João", PasswordHash: "hash-secreta"}

	b, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if strings.Contains(string(b), "hash-secreta") {
		t.Errorf("JSON expôs o hash da senha: %s", b)
	}
	if strings.Contains(string(b), "password_hash") {
		t.Errorf("JSON expôs o campo password_hash: %s", b)
	}
}

func TestUser_JSONContemCamposDoContrato(t *testing.T) {
	u := User{ID: 1, Name: "João", Email: "joao@example.com", Active: true}

	b, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, campo := range []string{"id", "name", "email", "active", "is_platform_admin", "email_verified_at", "created_at", "updated_at"} {
		if _, ok := body[campo]; !ok {
			t.Errorf("JSON sem o campo %q: %s", campo, b)
		}
	}
}
