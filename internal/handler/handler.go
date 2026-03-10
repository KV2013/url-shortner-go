package handler

import (
	"io"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/model"
)

//go:generate go run go.uber.org/mock/mockgen -source=handler.go -destination=mocks/handler_mock.go -package=mocks -typed
type URLService interface {
	SaveURL(url string) (*model.URL, error)
	GetByID(id string) (*model.URL, bool)
}

type URLHandler struct {
	urlService URLService
	config     *config.Config
}

func New(urlService URLService, config *config.Config) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		config:     config,
	}
}

func (h *URLHandler) Create(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	reqURL := string(reqBody)
	if reqURL == "" {
		http.Error(res, "no url provided", http.StatusBadRequest)
		return
	}

	storedURL, err := h.urlService.SaveURL(reqURL)
	if err != nil {
		http.Error(res, "Ne udalos sohranit url "+err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := h.config.BaseURL + "/" + storedURL.Short

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	io.WriteString(res, shortURL)
}

func (h *URLHandler) Redirect(res http.ResponseWriter, req *http.Request) {
	urlID := req.PathValue("id")
	if urlID == "" {
		http.Error(res, "empty id", http.StatusBadRequest)
		return
	}

	url, exists := h.urlService.GetByID(urlID)
	if !exists {
		http.NotFound(res, req)
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Redirect(res, req, url.Original, http.StatusTemporaryRedirect)
}
