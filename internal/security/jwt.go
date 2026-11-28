package security

import (
	"e-vendas/internal/config"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func GenerationToken(id int) (string, error) {
	permission := jwt.MapClaims{}
	permission["authorized"] = true
	permission["exp"] = time.Now().Add(time.Hour * 6).Unix()
	permission["id"] = id
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, permission)

	return token.SignedString([]byte(config.SecretKey))
}

func ValidateToken(r *http.Request) error {
	tokenString := extractToken(r)
	token, err := jwt.Parse(tokenString, returnKey)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	}

	return errors.New("Invalid token")
}

func ExtractUserID(r *http.Request) (int, error) {
    tokenString := extractToken(r)
    token, err := jwt.Parse(tokenString, returnKey)
	if err != nil {
		return 0, err
	}

	if permission, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, err := strconv.ParseInt(fmt.Sprintf("%.0f", permission["id"]), 10, 0)
		if err != nil {
			return 0, err
		}

		return int(userID), nil
	}

	return 0, errors.New("Invalid token")
}

func extractToken(r *http.Request) string {
	token := r.Header.Get("Authorized")
	if len(strings.Split(token, " ")) == 2 {
		return strings.Split(token,"")[1]
	}

	return ""
}

func returnKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Metodo de assinatura inseperado! %v", token.Header["alg"])
	}

	return config.SecretKey, nil
}
