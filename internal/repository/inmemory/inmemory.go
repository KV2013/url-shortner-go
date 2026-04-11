package inmemory

import (
	"errors"

	"github.com/KV2013/url-shortner-go/internal/model"
)

var ErrIDAlreadyExists = errors.New("id уже занят")

type InMemoryRepository struct {
	UrlsByID map[string]model.URL
}

func NewRepository() (*InMemoryRepository, error) {
	return &InMemoryRepository{
		UrlsByID: make(map[string]model.URL),
	}, nil
}

func (r *InMemoryRepository) GetByID(id string) (*model.URL, bool) {
	found, exists := r.UrlsByID[id]
	if !exists {
		return nil, false
	}

	return &found, true
}

func (r *InMemoryRepository) Save(url *model.URL) error {
	if _, exists := r.UrlsByID[url.Short]; exists {
		return ErrIDAlreadyExists
	}
	r.UrlsByID[url.Short] = *url

	return nil
}
