package handler

import (
	"e-vendas/internal/domain"
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type userHandler struct {
	userCase *usecase.UserCase
}

func NewUserHandler(usercase *usecase.UserCase) userHandler {
	return userHandler{
		userCase: usercase,
	}
}

func (us *userHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)

	if err != nil {
		responses.ResponseError(w, http.StatusBadRequest, "invalid_json", "JSON invalido")
		return
	}

	var trash any
	erro := decoder.Decode(&trash)
	if !errors.Is(erro, io.EOF) {
		responses.ResponseError(w, http.StatusBadRequest, "invalid_json", "JSON invalido")
		return
	}

	insertedUser, err := us.userCase.CreateUser(r.Context(), req.Name, req.Email, req.Phone, req.Password)

	if err != nil {
		var vErr *domain.ValidationError

		if errors.As(err, &vErr) {
			responses.ResponseError(w, http.StatusUnprocessableEntity, "validation_error", vErr.Message)
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			responses.ResponseError(w, http.StatusConflict, "conflict", "e-mail já cadastrado")
			return
		}
		log.Println(err)
		responses.ResponseError(w, http.StatusInternalServerError, "internal_error", "erro interno")
		return
	}

	resp := RegisterResponse{
		ID:    insertedUser.ID,
		Name:  insertedUser.Name,
		Email: insertedUser.Email,
		Phone: insertedUser.Phone,
	}

	responses.JSON(w, http.StatusCreated, resp)

}
