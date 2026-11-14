package responses

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_DefineContentTypeEStatus(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusCreated, map[string]string{"ok": "sim"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, esperado %d", w.Code, http.StatusCreated)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, esperado application/json", ct)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("corpo não é JSON válido: %v", err)
	}
	if body["ok"] != "sim" {
		t.Errorf("corpo = %v", body)
	}
}

func TestJSON_DataNuloEscreveApenasStatus(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusNoContent, nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, esperado %d", w.Code, http.StatusNoContent)
	}
	if w.Body.Len() != 0 {
		t.Errorf("corpo = %q, esperado vazio", w.Body.String())
	}
}

// Regressão: ARQUITETURA.md 6.4 define o envelope
// {"error": {"code": "...", "message": "..."}}.
func TestResponseError_SegueEnvelopeDocumentado(t *testing.T) {
	w := httptest.NewRecorder()

	ResponseError(w, http.StatusBadRequest, "validation_error", "nome é obrigatório")

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("corpo não é JSON válido: %v", err)
	}

	errorObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("resposta = %s; esperado envelope {\"error\":{\"code\":...,\"message\":...}}", w.Body.String())
	}
	if _, ok := errorObj["code"]; !ok {
		t.Errorf("envelope sem campo code: %s", w.Body.String())
	}
	if _, ok := errorObj["message"]; !ok {
		t.Errorf("envelope sem campo message: %s", w.Body.String())
	}
}
