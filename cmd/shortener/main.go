// URL Shortener — сервис сокращения ссылок.
//
// Принимает длинные URL и генерирует короткие идентификаторы для перенаправления.
// Поддерживает три типа хранилища: PostgreSQL, файл на диске и in-memory.
// Включает JWT-аутентификацию, gzip-сжатие, аудит-логирование и сбор метрик pprof.
//
// Запуск:
//
//	go run ./cmd/shortener/ -d "postgres://..." -a :8080 -p
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	shortnerpb "github.com/KV2013/url-shortner-go/api/shortner"
	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler"
	grpchandler "github.com/KV2013/url-shortner-go/internal/handler/grpc"
	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/KV2013/url-shortner-go/internal/repository"
	"github.com/KV2013/url-shortner-go/internal/router"
	"github.com/KV2013/url-shortner-go/internal/service"
	"github.com/KV2013/url-shortner-go/internal/tlscert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	printBuildInfo()

	config, cfgErr := config.NewConfig()
	if cfgErr != nil {
		log.Fatal("Ошибка при сборке конфига", zap.Error(cfgErr))
	}
	Logger, loggerErr := logger.New(config.LogLevel)
	if loggerErr != nil {
		log.Fatal("Ошибка при создании логгера")
	}

	repo, repoErr := repository.New(config, Logger)
	if repoErr != nil {
		Logger.Fatal("Ошибка при создании репозитория", zap.Error(repoErr))
	}
	defer repo.Close()

	urlService := service.NewURLService(repo, Logger)
	handler := handler.New(urlService, repo, config, Logger)
	mux := router.Init(context.Background(), handler, Logger, config)
	srv := &http.Server{
		Addr:         config.ServerAddress,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	grpcHandler := grpchandler.New(urlService, config, Logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthJWTUnaryInterceptor(config, Logger)),
	)
	shortnerpb.RegisterShortenerServiceServer(grpcServer, grpcHandler)

	go func() {
		if config.EnableHTTPS {
			certPaths, err := tlscert.ProvideCertAndKey()
			if err != nil {
				Logger.Fatal("Не удалось сгенерировать TLS сертификат", zap.Error(err))
			}

			Logger.Info("Сервер запущен (HTTPS)", zap.String("serverAddress", config.ServerAddress), zap.String("logLevel", config.LogLevel))
			if err := srv.ListenAndServeTLS(certPaths.CertPath, certPaths.KeyPath); err != nil && err != http.ErrServerClosed {
				Logger.Fatal("Не удалось запустить сервер", zap.Error(err))
			}
		} else {
			Logger.Info("Сервер запущен", zap.String("serverAddress", config.ServerAddress), zap.String("logLevel", config.LogLevel))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				Logger.Fatal("Не удалось запустить сервер", zap.Error(err))
			}
		}
	}()

	if config.EnablePprof {
		go func() {
			Logger.Info("pprof сервер запущен", zap.String("addr", ":8082"))
			if err := http.ListenAndServe(":8082", nil); err != nil {
				Logger.Error("ошибка pprof сервера", zap.Error(err))
			}
		}()
	}

	go func() {
		listener, err := net.Listen("tcp", config.GRPCPort)
		if err != nil {
			Logger.Fatal("Не удалось запустить gRPC сервер", zap.Error(err))
		}
		Logger.Info("gRPC сервер запущен", zap.String("grpcPort", config.GRPCPort))
		if err := grpcServer.Serve(listener); err != nil {
			Logger.Fatal("Не удалось запустить gRPC сервер", zap.Error(err))
		}
	}()

	// Ожидаем сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	Logger.Info("Получен сигнал завершения. Начинаем graceful shutdown...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		Logger.Fatal("Graceful shutdown не удался", zap.Error(err))
	}

	grpcServer.GracefulStop()

	Logger.Info("Сервер успешно остановлен")
}

func nA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", nA(buildVersion))
	fmt.Printf("Build date: %s\n", nA(buildDate))
	fmt.Printf("Build commit: %s\n", nA(buildCommit))
}
