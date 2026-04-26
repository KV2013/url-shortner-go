package router

import (
	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Init(handler *handler.URLHandler, logger *zap.Logger, cfg *config.Config) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.GzipCompression)
	r.Use(middleware.AuthJWT(cfg))
	r.Post("/", handler.Create)
	r.Get("/{id}", handler.Redirect)

	r.Get("/ping", handler.Ping)

	r.Post("/api/shorten", handler.APICreate)
	r.Post("/api/shorten/batch", handler.APICreateBatch)
	r.Get("/api/user/urls", handler.GetUserURLs)

	return r
}
