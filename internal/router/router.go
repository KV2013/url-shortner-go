package router

import (
	"context"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Init(ctx context.Context, handler *handler.URLHandler, logger *zap.Logger, cfg *config.Config) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.GzipCompression)
	r.Use(middleware.AuthJWT(cfg, logger))
	r.Use(middleware.RequestAuditor(ctx, cfg, logger))
	r.Post("/", handler.Create)
	r.Get("/{id}", handler.Redirect)

	r.Get("/ping", handler.Ping)

	r.Post("/api/shorten", handler.APICreate)
	r.Post("/api/shorten/batch", handler.APICreateBatch)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.Delete("/api/user/urls", handler.APIDeleteURLs)

	return r
}
