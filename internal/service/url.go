package service

import (
	"context"
	"errors"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/service/random"
)

type URLRepository interface {
	Save(ctx context.Context, url *model.URL) error
	SaveMany(ctx context.Context, urls []*model.URL) error
	GetByID(ctx context.Context, id string) (*model.URL, bool)
	GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error)
	DeleteURLs(ctx context.Context, ids []model.URL) error
}

type URLService struct {
	urlRepository URLRepository
}

func NewURLService(urlRepository URLRepository) *URLService {
	return &URLService{urlRepository: urlRepository}
}

func (s *URLService) SaveURL(ctx context.Context, url string, userID string) (*model.URL, error) {
	newID := random.NewRandomString(10)
	newURL := &model.URL{
		Original: url,
		Short:    newID,
		UserID:   userID,
	}
	err := s.urlRepository.Save(ctx, newURL)
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func (s *URLService) SaveManyURL(ctx context.Context, urls []string, userID string) ([]model.URL, error) {
	newURLs := make([]*model.URL, 0, len(urls))
	for _, url := range urls {
		newURLs = append(newURLs, &model.URL{
			Original: url,
			Short:    random.NewRandomString(10),
			UserID:   userID,
		})
	}

	if err := s.urlRepository.SaveMany(ctx, newURLs); err != nil {
		return nil, err
	}

	result := make([]model.URL, 0, len(newURLs))
	for _, u := range newURLs {
		result = append(result, *u)
	}
	return result, nil
}

func (s *URLService) GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	return s.urlRepository.GetAllByUserID(ctx, userID)
}

func (s *URLService) GetByID(ctx context.Context, id string) (*model.URL, error) {
	url, exists := s.urlRepository.GetByID(ctx, id)
	if !exists {
		return nil, &model.ErrURLNotFound{Short: id}
	}

	if url.DeletedFlag {
		return nil, &model.ErrUrlDeleted{Short: id}
	}

	return url, nil
}

func (s *URLService) DeleteURLs(ctx context.Context, shortUrls []string, userID string) error {
	// получить URL по каждому id и проверить, что они принадлежат пользователю
	var urls []model.URL
	for _, shortUrl := range shortUrls {
		url, err := s.GetByID(ctx, shortUrl)
		if err != nil {
			return err
		}

		if url.UserID != userID {
			return errors.New("URL с id " + shortUrl + " не принадлежит пользователю")
		}
		urls = append(urls, *url)
	}

	return s.urlRepository.DeleteURLs(ctx, urls)
}
