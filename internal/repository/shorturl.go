package repository

import (
	"fmt"

	"github.com/KV2013/url-shortner-go/internal/model"
)

type URLCollection struct {
	UrlsByID  map[string]model.URL
	UrlsByURL map[string]model.URL
}

func NewURLCollection() URLCollection {
	return URLCollection{
		UrlsByID:  make(map[string]model.URL),
		UrlsByURL: make(map[string]model.URL),
	}
}

func (uc *URLCollection) FindByURL(url string) (model.URL, bool) {
	found, exists := uc.UrlsByURL[url]
	return found, exists
}
func (uc *URLCollection) FindByID(id string) (model.URL, bool) {
	found, exists := uc.UrlsByID[id]
	return found, exists
}

func (uc *URLCollection) Set(url model.URL) error {
	if _, exists := uc.UrlsByURL[url.Original]; exists {
		return fmt.Errorf("already exists")
	}
	if _, exists := uc.UrlsByID[url.Short]; exists {
		return fmt.Errorf("already exists")
	}
	uc.UrlsByID[url.Short] = url
	uc.UrlsByURL[url.Original] = url

	return nil
}
