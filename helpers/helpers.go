package helpers

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var JwtSecret string

func init() {

	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}
	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		log.Fatal("Error loading JWT_SECRET from .env file")
	}

}

func CreateJWT(userId uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userId,
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour * 7 * 24)),
	})

	tokenString, err := token.SignedString([]byte(JwtSecret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return tokenString, nil
}

func VerifyJWT(tokenString string) (float64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if errors.Is(err, jwt.ErrTokenExpired) {
		return 0, errors.New("Token is expired")
	}

	if err != nil || !token.Valid {
		return 0, errors.New("Token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("Error parsing claims")
	}

	userId := claims["id"].(float64)
	return userId, nil
}
