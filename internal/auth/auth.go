package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func validateJWT(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return false, err
	}

	return true, nil
}

func AuthRequest(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// expected: "Bearer <token>"
	authSplit := strings.Split(authHeader, " ")
	if len(authSplit) != 2 || authSplit[0] != "Bearer" {
		return false
	}

	token := authSplit[1]
	isValid, _ := validateJWT(token)
	return isValid
}
