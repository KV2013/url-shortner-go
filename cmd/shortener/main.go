package main

import (
	"log"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/repository/inmemory"
	"github.com/KV2013/url-shortner-go/internal/router"
	"github.com/KV2013/url-shortner-go/internal/service"
)

func main() {

	config, cfgErr := config.NewConfig()
	if cfgErr != nil {
		log.Fatal("Oshibka pri sborke konfiga")
	}
	Logger, loggerErr := logger.New(config.LogLevel)
	if loggerErr != nil {
		log.Fatal("Oshibka pri sozdanii logger")
	}

	repo := inmemory.NewRepository()
	urlService := service.NewURLService(repo)
	handler := handler.New(urlService, config)
	mux := router.Init(handler, Logger)

	log.Println("server zapuchen na " + config.ServerAddress)

	err := http.ListenAndServe(config.ServerAddress, mux)

	if err != nil {
		log.Fatal("Ne udalos zapustit server ", err)
	}
}
