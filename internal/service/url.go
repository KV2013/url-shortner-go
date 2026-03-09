package service

import (
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/service/random"
)

type URLRepository interface {
	Save(url *model.URL) error
	GetByID(id string) (*model.URL, bool)
}

type URLService struct {
	urlRepository URLRepository
}

func NewURLService(urlRepository URLRepository) *URLService {
	return &URLService{urlRepository: urlRepository}
}

func (s *URLService) SaveURL(url string) (*model.URL, error) {
	newID := random.NewRandomString(10)
	newURL := &model.URL{
		Original: url,
		Short:    newID,
	}
	err := s.urlRepository.Save(newURL)
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func (s *URLService) GetByID(id string) (*model.URL, bool) {
	url, exists := s.urlRepository.GetByID(id)
	if !exists {
		return nil, false
	}

	return url, true
}
