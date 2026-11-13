package handler

import (
	"e-vendas/internal/domain"
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"
	"encoding/json"
	"errors"
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

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		responses.ResponseError(w, http.StatusBadRequest, err)
		return
	}

	insertedUser, err := us.userCase.CreateUser(r.Context(), req.Name, req.Email, req.Password)

	if err != nil {
		var vErr *domain.ValidationError

		if errors.As(err, &vErr) {
			responses.ResponseError(w, http.StatusUnprocessableEntity, errors.New(vErr.Message))
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			responses.ResponseError(w, http.StatusConflict, errors.New("e-mail já cadastrado"))
			return
		}
		log.Println(err)
		responses.ResponseError(w, http.StatusInternalServerError, errors.New("erro interno"))
		return
	}

	resp := RegisterResponse{
		ID:    insertedUser.ID,
		Name:  insertedUser.Name,
		Email: insertedUser.Email,
	}

	responses.JSON(w, http.StatusCreated, resp)

}
