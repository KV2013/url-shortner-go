package repository

import (
	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/repository/file"
	"github.com/KV2013/url-shortner-go/internal/repository/inmemory"
	sqlxrepo "github.com/KV2013/url-shortner-go/internal/repository/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	Save(url *model.URL) error
	GetByID(id string) (*model.URL, bool)
	Ping() error
	Close() error
}

func New(cfg *config.Config, logger *zap.Logger) (Repository, error) {
	if cfg.DatabaseDSN != "" {
		return sqlxrepo.NewRepository(cfg.DatabaseDSN)
	}
	if cfg.FileStoragePath != "" {
		return file.NewRepository(cfg.FileStoragePath, logger)
	}
	return inmemory.NewRepository()
}
