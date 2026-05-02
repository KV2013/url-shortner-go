package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TokenExp = time.Hour * 3

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

var errUnexpectedSigningMethodError = errors.New("unexpected signing method")
var errInvalidTokenError = errors.New("invalid token")

func GenerateAccessToken(userID string, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})
	return token.SignedString([]byte(secretKey))
}

func GetUserID(tokenString string, secretKey string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedSigningMethodError
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return "", errInvalidTokenError
	}
	return claims.UserID, nil
}
