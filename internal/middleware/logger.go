package middleware

import (
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func ZapLogger(l *zap.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&logger.ZapLogFormatter{Logger: l, UserIDKey: UserIDContextKey})
}
