package repository

import (
	"fmt"

	"github.com/KV2013/url-shortner-go/internal/model"
)

type URLCollection struct {
	urlsById map[string]model.URL
	urlByURL map[string]model.URL
}

func NewURLCollection() URLCollection {
	return URLCollection{
		urlsById: make(map[string]model.URL),
		urlByURL: make(map[string]model.URL),
	}
}

func (uc *URLCollection) FindByURL(url string) (model.URL, bool) {
	found, exists := uc.urlByURL[url]
	return found, exists
}
func (uc *URLCollection) FindByID(id string) (model.URL, bool) {
	found, exists := uc.urlsById[id]
	return found, exists
}

func (uc *URLCollection) Set(url model.URL) error {
	if _, exists := uc.urlByURL[url.Original]; exists {
		return fmt.Errorf("already exists")
	}
	if _, exists := uc.urlsById[url.Short]; exists {
		return fmt.Errorf("already exists")
	}
	uc.urlsById[url.Short] = url
	uc.urlByURL[url.Original] = url

	return nil
}
