package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

var algo = string(os.Getenv("JWT_ALGO"))

func GenerateJWT(userID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(algo), claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func validateJWT(tokenString string, expectedAlg string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token: %v", token.Header["alg"])
		}
		if token.Method != jwt.GetSigningMethod(expectedAlg) {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("JWT validation failed: invalid claims")
	}
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return nil, fmt.Errorf("JWT validation failed: token expired")
		}
	}

	return claims, nil
}

func VerifyRequest(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("Missing Authorization header in request")
	}

	// expected: "Bearer <token>"
	authSplit := strings.Split(authHeader, " ")
	if len(authSplit) != 2 || authSplit[0] != "Bearer" {
		return nil, fmt.Errorf("Malformed Authorization header in request")
	}

	token := authSplit[1]
	claims, err := validateJWT(token, algo)
	if err != nil {
		fmt.Printf("JWT error: %v for token: %s\n", err, token)
		return nil, err
	}

	return claims, nil
}
