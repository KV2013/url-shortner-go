package inmemory

import (
	"errors"

	"github.com/KV2013/url-shortner-go/internal/model"
)

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
		return errors.New("id uzje zanyat")
	}
	r.UrlsByID[url.Short] = *url

	return nil
}
