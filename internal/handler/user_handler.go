package handler

import (
	"e-vendas/internal/domain"
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"
	"encoding/json"
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
		responses.ResponseError(w, http.StatusUnprocessableEntity, err)
		return
	}

	var user domain.User
	user.Name = req.Name
	user.Email = req.Email
	user.PasswordHash = req.Password

	inserUser, err := us.userCase.CreateUser(r.Context(), user)

	if err != nil {
		responses.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusCreated, inserUser)

}
