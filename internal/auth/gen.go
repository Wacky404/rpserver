package auth

import (
    "fmt"
    "time"
    "os"
    "github.com/goland-jwt/jwt/v5"
)

var jwtkey = []byte(os.Getenv("JWT_KEY"))

func GenerateJWT() (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user": "testuser",
        "exp": time.Now().Add(time.Hour * 1).Unix(), // set it to expire after one hour
    })

    return token.SignedString(jwtSecret)
}
