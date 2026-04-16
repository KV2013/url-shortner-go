package service

import (
	"context"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/service/random"
)

type URLRepository interface {
	Save(ctx context.Context, url *model.URL) error
	GetByID(ctx context.Context, id string) (*model.URL, bool)
}

type URLService struct {
	urlRepository URLRepository
}

func NewURLService(urlRepository URLRepository) *URLService {
	return &URLService{urlRepository: urlRepository}
}

func (s *URLService) SaveURL(ctx context.Context, url string) (*model.URL, error) {
	newID := random.NewRandomString(10)
	newURL := &model.URL{
		Original: url,
		Short:    newID,
	}
	err := s.urlRepository.Save(ctx, newURL)
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func (s *URLService) GetByID(ctx context.Context, id string) (*model.URL, bool) {
	url, exists := s.urlRepository.GetByID(ctx, id)
	if !exists {
		return nil, false
	}

	return url, true
}
