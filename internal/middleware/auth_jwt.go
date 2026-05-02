package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/service/auth"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tokenCookieName = "token"

func AuthJWT(cfg *config.Config, logger *zap.Logger) func(next http.Handler) http.Handler {
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

			logger.Debug("middleware.AuthJWT", zap.String("userID", userID), zap.String("cookie", fmt.Sprint(cookie)), zap.Error(err))
			if userID == "" {
				userID = generateUserID()
				var parts []string
				for _, c := range r.Cookies() {
					parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
				}

				logger.Debug(
					"middleware.AuthJWT: generated new userID",
					zap.String("userID", userID),
					zap.String("cookies", strings.Join(parts, "; ")),
				)

				tokenString, tokenErr := auth.GenerateAccessToken(userID, cfg.JWTSecretKey)
				if tokenErr == nil {
					http.SetCookie(w, &http.Cookie{
						Name:     tokenCookieName,
						Value:    tokenString,
						Expires:  time.Now().Add(auth.TokenExp),
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
