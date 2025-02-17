package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

var (
	jwtsecret = []byte(os.Getenv("JWT_SECRET"))
	algo      = string(os.Getenv("JWT_ALGO"))
)

func validateJWT(tokenString string, expectedAlg string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token: %v", token.Header["alg"])
		}
		if token.Header["alg"] != expectedAlg {
			return nil, fmt.Errorf("incorrect alg: %v", token.Header["alg"])
		}
		return jwtsecret, nil
	})
	if err != nil || !token.Valid {
		return false, err
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	}
	return false, fmt.Errorf("invalid token")
}

func Verifyrequest(r *http.Request) bool {
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
	isValid, _ := validateJWT(token, algo)
	return isValid
}
