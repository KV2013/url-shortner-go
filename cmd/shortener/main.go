package main

import (
	"log"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/repository/in_memory"
	"github.com/KV2013/url-shortner-go/internal/router"
	"github.com/KV2013/url-shortner-go/internal/service"
)

func main() {

	repo := in_memory.NewRepository()
	urlService := service.NewURLService(repo)
	handler := handler.New(urlService)
	mux := router.Init(handler)

	log.Println("server zapuchen na portu 8080")

	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		log.Fatal("Ne udalos zapustit server ", err)
	}
}
