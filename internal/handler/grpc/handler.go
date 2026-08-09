package grpc

import (
	"context"
	"errors"

	shortnerpb "github.com/KV2013/url-shortner-go/api/shortner"
	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/KV2013/url-shortner-go/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type URLService interface {
	SaveURL(ctx context.Context, url string, userID string) (*model.URL, error)
	GetByID(ctx context.Context, id string) (*model.URL, error)
	GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error)
}

type ShortenerHandler struct {
	shortnerpb.UnimplementedShortenerServiceServer
	urlService URLService
	config     *config.Config
	logger     *zap.Logger
}

func New(urlService URLService, config *config.Config, logger *zap.Logger) *ShortenerHandler {
	return &ShortenerHandler{
		urlService: urlService,
		config:     config,
		logger:     logger,
	}
}

func (h *ShortenerHandler) ShortenURL(ctx context.Context, req *shortnerpb.URLShortenRequest) (*shortnerpb.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "URL не задан")
	}

	storedURL, err := h.urlService.SaveURL(ctx, req.GetUrl(), middleware.UserIDFromContext(ctx))
	if err != nil {
		var urlExists *model.ErrURLAlreadyExists
		if errors.As(err, &urlExists) {
			shortURL := h.config.BaseURL + "/" + urlExists.URL.Short
			return nil, status.Errorf(codes.AlreadyExists, "%s", shortURL)
		}
		h.logger.Error("ShortenURL: ошибка сохранения URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "не удалось сохранить URL")
	}

	resp := &shortnerpb.URLShortenResponse{}
	resp.SetResult(h.config.BaseURL + "/" + storedURL.Short)
	return resp, nil
}

func (h *ShortenerHandler) ExpandURL(ctx context.Context, req *shortnerpb.URLExpandRequest) (*shortnerpb.URLExpandResponse, error) {
	url, err := h.urlService.GetByID(ctx, req.GetId())
	if err != nil {
		var errNotFound *model.ErrURLNotFound
		if errors.As(err, &errNotFound) {
			return nil, status.Errorf(codes.NotFound, "URL не найден: %s", req.GetId())
		}
		var errDeleted *model.ErrURLDeleted
		if errors.As(err, &errDeleted) {
			return nil, status.Errorf(codes.NotFound, "URL удалён: %s", req.GetId())
		}
		h.logger.Error("ExpandURL: ошибка получения URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "ошибка при получении URL")
	}

	resp := &shortnerpb.URLExpandResponse{}
	resp.SetResult(url.Original)
	return resp, nil
}

func (h *ShortenerHandler) ListUserURLs(ctx context.Context, _ *shortnerpb.ListUserURLsRequest) (*shortnerpb.UserURLsResponse, error) {
	userID := middleware.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "пользователь не авторизован")
	}

	urls, err := h.urlService.GetAllByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("ListUserURLs: ошибка получения URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "ошибка получения URL")
	}

	items := make([]*shortnerpb.URLData, 0, len(urls))
	for _, u := range urls {
		item := &shortnerpb.URLData{}
		item.SetShortUrl(h.config.BaseURL + "/" + u.Short)
		item.SetOriginalUrl(u.Original)
		items = append(items, item)
	}

	resp := &shortnerpb.UserURLsResponse{}
	resp.SetUrl(items)
	return resp, nil
}
