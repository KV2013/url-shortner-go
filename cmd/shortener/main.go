package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/repository"
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

	repo, repoErr := repository.New(config, Logger)
	if repoErr != nil {
		Logger.Fatal("Ошибка при создании репозитория")
	}
	defer repo.Close()

	urlService := service.NewURLService(repo)
	handler := handler.New(urlService, repo, config, Logger)
	mux := router.Init(handler, Logger)

	srv := &http.Server{
		Addr:         config.ServerAddress,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		Logger.Info("Сервер запущен", zap.String("serverAddress", config.ServerAddress), zap.String("logLevel", config.LogLevel))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Logger.Fatal("Не удалось запустить сервер", zap.Error(err))
		}
	}()

	// Ожидаем сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	Logger.Info("Получен сигнал завершения. Начинаем graceful shutdown...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		Logger.Fatal("Graceful shutdown не удался", zap.Error(err))
	}

	Logger.Info("Сервер успешно остановлен")
}
