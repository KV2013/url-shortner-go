package inmemory

import (
	"context"
	"errors"
	"sync"
	"time"

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

func (r *InMemoryRepository) GetByID(_ context.Context, id string) (*model.URL, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	found, exists := r.UrlsByID[id]
	if !exists {
		return nil, false
	}

	return &found, true
}

func (r *InMemoryRepository) SaveMany(_ context.Context, urls []*model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range urls {
		if _, exists := r.UrlsByID[url.Short]; exists {
			return ErrIDAlreadyExists
		}
		r.UrlsByID[url.Short] = *url
	}
	return nil
}

func (r *InMemoryRepository) GetAllByUserID(_ context.Context, userID string) ([]model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.URL
	for _, url := range r.UrlsByID {
		if url.UserID == userID {
			result = append(result, url)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) Ping() error {
	return nil
}

func (r *InMemoryRepository) Close() error {
	return nil
}

func (r *InMemoryRepository) Save(_ context.Context, url *model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.UrlsByID[url.Short]; exists {
		return ErrIDAlreadyExists
	}
	r.UrlsByID[url.Short] = *url

	return nil
}

func (r *InMemoryRepository) GetStats(_ context.Context) (int, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	urlsCount := 0
	usersSet := make(map[string]struct{})
	for _, url := range r.UrlsByID {
		if !url.DeletedFlag {
			urlsCount++
		}
		if url.UserID != "" {
			usersSet[url.UserID] = struct{}{}
		}
	}
	return urlsCount, len(usersSet), nil
}

func (r *InMemoryRepository) DeleteUserURLs(_ context.Context, userID string, urls []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, short := range urls {
		if url, exists := r.UrlsByID[short]; exists && url.UserID == userID {
			now := time.Now()
			url.DeletedFlag = true
			url.DeletedAt = &now
			r.UrlsByID[short] = url
		}
	}
	return nil
}
