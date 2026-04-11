package inmemory

import (
	"errors"
	"sync"

	"github.com/KV2013/url-shortner-go/internal/model"
)

var ErrIDAlreadyExists = errors.New("id уже занят")

type InMemoryRepository struct {
	mu       sync.RWMutex
	UrlsByID map[string]model.URL
}

func NewRepository() (*InMemoryRepository, error) {
	return &InMemoryRepository{
		UrlsByID: make(map[string]model.URL),
	}, nil
}

func (r *InMemoryRepository) GetByID(id string) (*model.URL, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	found, exists := r.UrlsByID[id]
	if !exists {
		return nil, false
	}

	return &found, true
}

func (r *InMemoryRepository) Save(url *model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.UrlsByID[url.Short]; exists {
		return ErrIDAlreadyExists
	}
	r.UrlsByID[url.Short] = *url

	return nil
}
