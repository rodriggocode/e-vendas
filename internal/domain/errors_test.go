package domain

import (
	"errors"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Message: "o nome é obrigatório"}

	if err.Error() != "o nome é obrigatório" {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestErrosSentinela_SaoIdentificaveis(t *testing.T) {
	if !errors.Is(ErrConflict, ErrConflict) {
		t.Error("ErrConflict não é identificável com errors.Is")
	}
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound não é identificável com errors.Is")
	}
	if errors.Is(ErrConflict, ErrNotFound) {
		t.Error("ErrConflict e ErrNotFound deveriam ser distintos")
	}
}
