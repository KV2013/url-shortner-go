package main

import (
	"log"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/repository/file"
	"github.com/KV2013/url-shortner-go/internal/router"
	"github.com/KV2013/url-shortner-go/internal/service"
	"go.uber.org/zap"
)

func main() {

	config, cfgErr := config.NewConfig()
	if cfgErr != nil {
		log.Fatal("Ошибка при сборке конфига")
	}
	Logger, loggerErr := logger.New(config.LogLevel)
	if loggerErr != nil {
		log.Fatal("Ошибка при создании логгера")
	}

	repo, repoErr := file.NewRepository(config.FileStoragePath, Logger)
	if repoErr != nil {
		Logger.Fatal("Ошибка при создании репозитория")
	}
	defer repo.Close()

	urlService := service.NewURLService(repo)
	handler := handler.New(urlService, config)
	mux := router.Init(handler, Logger)

	Logger.Info("Сервер запущен", zap.String("serverAddress", config.ServerAddress), zap.String("logLevel", config.LogLevel))

	err := http.ListenAndServe(config.ServerAddress, mux)

	if err != nil {
		Logger.Fatal("Не удалось запустить сервер", zap.Error(err))
	}
}
