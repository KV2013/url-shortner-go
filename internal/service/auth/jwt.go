package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TOKEN_EXP = time.Hour * 3

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

var unexpectedSigningMethodError = errors.New("unexpected signing method")
var invalidTokenError = errors.New("invalid token")

func GenerateAccessToken(userID string, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})
	return token.SignedString([]byte(secretKey))
}

func GetUserID(tokenString string, secretKey string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, unexpectedSigningMethodError
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return "", invalidTokenError
	}
	return claims.UserID, nil
}
