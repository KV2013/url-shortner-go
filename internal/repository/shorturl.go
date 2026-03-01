package repository

import (
	"fmt"

	"github.com/KV2013/url-shortner-go/internal/model"
)

type URLCollection struct {
	urlsByID map[string]model.URL
	urlByURL map[string]model.URL
}

func NewURLCollection() URLCollection {
	return URLCollection{
		urlsByID: make(map[string]model.URL),
		urlByURL: make(map[string]model.URL),
	}
}

func (uc *URLCollection) FindByURL(url string) (model.URL, bool) {
	found, exists := uc.urlByURL[url]
	return found, exists
}
func (uc *URLCollection) FindByID(id string) (model.URL, bool) {
	found, exists := uc.urlsByID[id]
	return found, exists
}

func (uc *URLCollection) Set(url model.URL) error {
	if _, exists := uc.urlByURL[url.Original]; exists {
		return fmt.Errorf("already exists")
	}
	if _, exists := uc.urlsByID[url.Short]; exists {
		return fmt.Errorf("already exists")
	}
	uc.urlsByID[url.Short] = url
	uc.urlByURL[url.Original] = url

	return nil
}
