package main

import (
	"log"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/handler/create"
	"github.com/KV2013/url-shortner-go/internal/handler/redirect"
	"github.com/KV2013/url-shortner-go/internal/repository"
)

func main() {

	mux := http.NewServeMux()

	urls := repository.NewURLCollection()

	BaseURL := "http://localhost:8080/"
	mux.HandleFunc(`POST /`, create.New(urls, BaseURL))
	mux.HandleFunc(`GET /{id}`, redirect.New(urls))
	http.ListenAndServe(":8080", mux)

	log.Println("server zapuchen")
	err := http.ListenAndServe(`:8080`, mux)
	log.Fatal(err)
}
