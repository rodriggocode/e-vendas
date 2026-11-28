package handler

import (
	"e-vendas/internal/responses"
	"e-vendas/internal/usecase"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"io"
)

type LoginHandler struct{
    loginUseCase *usecase.LoginUseCase
}

func NewLoginHandler(loginUsecase *usecase.LoginUseCase) *LoginHandler{
    return &LoginHandler{loginUseCase: loginUsecase}
}

func (h *LoginHandler)Login(w http.ResponseWriter, r *http.Request) {
	mediType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediType != "application/json" {
		responses.ResponseError(
			w,
			http.StatusUnsupportedMediaType,
			"unsupported_media_type",
			"o Content-Type deve ser application/json",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request LoginRequest

	err = decoder.Decode(&request)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			responses.ResponseError(w, http.StatusRequestEntityTooLarge, "request_too_large",
				"o corpo da requisição excede o limite de 1 MiB")
			return
		}

		responses.ResponseError(w, http.StatusBadRequest,   "invalid_request",
                "corpo da requisição inválido")
		return
	}


    var trash any
    err = decoder.Decode(&trash)
    if !errors.Is(err, io.EOF){
        var maxBytesErr *http.MaxBytesError
        if errors.As(err, &maxBytesErr){
            responses.ResponseError(
                w,
                http.StatusRequestEntityTooLarge,
                "request_too_large",
                "o corpo da requisicao excede o limite de 1 MiB",
            )
            return
        }
        responses.ResponseError(
            w,
            http.StatusBadRequest,
            "invalid_json",
            "corpo da requisicao invalido ou contendo dados adicionais",
        )
        return
    }

    if request.Email == "" || request.Password == ""{
        responses.ResponseError(
            w,
            http.StatusBadRequest,
            "validation_error",
            "email e senha sao obrigatorios",
        )
        return
    }

    user, token, err := h.loginUseCase.Login(
        r.Context(),
        request.Email,
        request.Password,
    )

    if err != nil{
        if errors.Is(err, usecase.ErrInvalidCredentials) ||
            errors.Is(err, usecase.ErrInactiveUser){
                responses.ResponseError(
                    w,
                    http.StatusUnauthorized,
                    "invalid_credentials",
                    "email ou senha invalidos",
                )

                return
            }

            responses.ResponseError(
                w,
                http.StatusInternalServerError,
                "internal_error",
                "erro ao realizar login",
            )
            return
    }

    response := LoginResponse{
        ID: user.ID,
        Name: user.Name,
        Email: user.Email,
        Token: token,
    }


   responses.JSON(w, http.StatusOK, response)

}
