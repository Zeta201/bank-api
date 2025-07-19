package config

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("your_secret_key")

func GenerateJWT(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString(JWTSecret)
}
