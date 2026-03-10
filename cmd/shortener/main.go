package main

import (
	"log"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/repository/inmemory"
	"github.com/KV2013/url-shortner-go/internal/router"
	"github.com/KV2013/url-shortner-go/internal/service"
)

func main() {

	config := config.NewConfig()

	repo := inmemory.NewRepository()
	urlService := service.NewURLService(repo)
	handler := handler.New(urlService, config)
	mux := router.Init(handler)

	log.Println("server zapuchen na " + config.ServerAddress)

	err := http.ListenAndServe(config.ServerAddress, mux)

	if err != nil {
		log.Fatal("Ne udalos zapustit server ", err)
	}
}
