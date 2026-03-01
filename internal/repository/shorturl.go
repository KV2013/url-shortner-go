package repository

import (
	"fmt"

	"github.com/KV2013/url-shortner-go/internal/model"
)

type UrlCollection struct {
	urlsById map[string]model.Url
	urlByURL map[string]model.Url
}

func NewUrlCollection() UrlCollection {
	return UrlCollection{
		urlsById: make(map[string]model.Url),
		urlByURL: make(map[string]model.Url),
	}
}

func (this *UrlCollection) FindByUrl(url string) (model.Url, bool) {
	found, exists := this.urlByURL[url]
	return found, exists
}
func (this *UrlCollection) FindById(id string) (model.Url, bool) {
	found, exists := this.urlsById[id]
	return found, exists
}

func (this *UrlCollection) Set(url model.Url) error {
	if _, exists := this.urlByURL[url.Original]; exists {
		return fmt.Errorf("already exists")
	}
	if _, exists := this.urlsById[url.Short]; exists {
		return fmt.Errorf("already exists")
	}
	this.urlsById[url.Short] = url
	this.urlByURL[url.Original] = url

	return nil
}
