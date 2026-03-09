package router

import (
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/handler"
)

func Init(handler *handler.URLHandler) *http.ServeMux {

	mux := http.NewServeMux()
	mux.HandleFunc(`POST /`, handler.NewCreate())
	mux.HandleFunc(`GET /{id}`, handler.NewRedirect())

	return mux
}
