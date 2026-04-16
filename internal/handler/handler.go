package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/mailru/easyjson"
)

//go:generate go run go.uber.org/mock/mockgen -source=handler.go -destination=mocks/handler_mock.go -package=mocks -typed
type URLService interface {
	SaveURL(ctx context.Context, url string) (*model.URL, error)
	GetByID(ctx context.Context, id string) (*model.URL, bool)
}

type Pinger interface {
	Ping() error
}

type URLHandler struct {
	urlService URLService
	pinger     Pinger
	config     *config.Config
}

func New(urlService URLService, pinger Pinger, config *config.Config) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		pinger:     pinger,
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
		http.Error(res, "URL не задан", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	storedURL, err := h.urlService.SaveURL(ctx, reqURL)
	if err != nil {
		http.Error(res, "Не удалось сохранить url "+err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := h.config.BaseURL + "/" + storedURL.Short

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	io.WriteString(res, shortURL)
}

func (h *URLHandler) APICreate(res http.ResponseWriter, req *http.Request) {
	var decoded model.CreateURLRequest
	res.Header().Set("Content-Type", "application/json")

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	err = easyjson.Unmarshal(reqBody, &decoded)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if decoded.URL == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	storedURL, err := h.urlService.SaveURL(req.Context(), decoded.URL)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	resp := model.CreateURLResponse{
		Result: h.config.BaseURL + "/" + storedURL.Short,
	}
	jsonBody, err := easyjson.Marshal(resp)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Write(jsonBody)
}

func (h *URLHandler) Redirect(res http.ResponseWriter, req *http.Request) {
	urlID := req.PathValue("id")
	if urlID == "" {
		http.Error(res, "не задан id", http.StatusBadRequest)
		return
	}

	url, exists := h.urlService.GetByID(req.Context(), urlID)
	if !exists {
		http.NotFound(res, req)
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Redirect(res, req, url.Original, http.StatusTemporaryRedirect)
}

func (h *URLHandler) Ping(res http.ResponseWriter, req *http.Request) {
	if err := h.pinger.Ping(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
