package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/service/auth"
	"github.com/google/uuid"
)

const tokenCookieName = "token"

func AuthJWT(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie(tokenCookieName)
			if err == nil {
				id, parseErr := auth.GetUserID(cookie.Value, cfg.JWTSecretKey)
				if parseErr == nil && id != "" {
					userID = id
				}
			}

			if userID == "" {
				userID = generateUserID()

				tokenString, tokenErr := auth.GenerateAccessToken(userID, cfg.JWTSecretKey)
				if tokenErr == nil {
					http.SetCookie(w, &http.Cookie{
						Name:     tokenCookieName,
						Value:    tokenString,
						Expires:  time.Now().Add(auth.TOKEN_EXP),
						HttpOnly: true,
						Path:     "/",
					})
				}
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateUserID() string {
	return uuid.New().String()
}
