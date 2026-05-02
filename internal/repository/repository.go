package repository

import (
	"context"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/repository/file"
	"github.com/KV2013/url-shortner-go/internal/repository/inmemory"
	sqlxrepo "github.com/KV2013/url-shortner-go/internal/repository/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	Save(ctx context.Context, url *model.URL) error
	SaveMany(ctx context.Context, urls []*model.URL) error
	GetByID(ctx context.Context, id string) (*model.URL, bool)
	GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error)
	DeleteUserURLs(ctx context.Context, userID string, urls []string) error
	Ping() error
	Close() error
}

func New(cfg *config.Config, logger *zap.Logger) (Repository, error) {
	if cfg.DatabaseDSN != "" {
		return sqlxrepo.NewRepository(cfg.DatabaseDSN, logger)
	}
	if cfg.FileStoragePath != "" {
		return file.NewRepository(cfg.FileStoragePath, logger)
	}
	return inmemory.NewRepository()
}
