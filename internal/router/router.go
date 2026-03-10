package router

import (
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Init(handler *handler.URLHandler) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/", handler.Create)
	r.Get("/{id}", handler.Redirect)

	return r
}
