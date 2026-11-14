package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Println("Erro ao gerar JSON", err)
		}
	}
}

func ResponseError(w http.ResponseWriter, statusCode int, code string, message string) {
	JSON(w, statusCode, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}
