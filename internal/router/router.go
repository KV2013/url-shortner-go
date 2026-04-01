package router

import (
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Init(handler *handler.URLHandler, logger *zap.Logger) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.GzipCompression)
	r.Post("/", handler.Create)
	r.Post("/api/shorten", handler.ApiCreate)
	r.Get("/{id}", handler.Redirect)

	return r
}
