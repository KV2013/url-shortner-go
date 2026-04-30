package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen -source=handler.go -destination=mocks/handler_mock.go -package=mocks -typed
type URLService interface {
	SaveURL(ctx context.Context, url string, userID string) (*model.URL, error)
	SaveManyURL(ctx context.Context, urls []string, userID string) ([]model.URL, error)
	GetByID(ctx context.Context, id string) (*model.URL, bool)
	GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error)
	DeleteURLs(ctx context.Context, ids []string, userID string) error
}

func userIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(middleware.UserIDContextKey).(string)
	return id
}

type Pinger interface {
	Ping() error
}

type URLHandler struct {
	urlService URLService
	pinger     Pinger
	config     *config.Config
	logger     *zap.Logger
}

func (h *URLHandler) writeJSONError(res http.ResponseWriter, errMsg string, status int) {
	body, _ := json.Marshal(model.APIErrorResponse{Error: errMsg})
	res.WriteHeader(status)
	if _, err := res.Write(body); err != nil {
		h.logger.Error("ошибка при записи ответа", zap.Error(err))
	}
}

func New(urlService URLService, pinger Pinger, config *config.Config, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		pinger:     pinger,
		config:     config,
		logger:     logger,
	}
}

func (h *URLHandler) Create(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "ошибка в теле запроса "+err.Error(), http.StatusBadRequest)
		return
	}

	reqURL := string(reqBody)
	if reqURL == "" {
		http.Error(res, "URL не задан", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	storedURL, err := h.urlService.SaveURL(ctx, reqURL, userIDFromContext(ctx))
	if err != nil {
		var urlExists *model.ErrURLAlreadyExists
		if errors.As(err, &urlExists) {
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusConflict)
			_, err := io.WriteString(res, h.config.BaseURL+"/"+urlExists.URL.Short)
			if err != nil {
				h.logger.Error("ошибка при записи ответа", zap.Error(err))
				http.Error(res, "ошибка при записи ответа", http.StatusInternalServerError)
				return
			}
			return
		}
		http.Error(res, "Не удалось сохранить url "+err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := h.config.BaseURL + "/" + storedURL.Short

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	_, err = io.WriteString(res, shortURL)
	if err != nil {
		h.logger.Error("ошибка при записи ответа", zap.Error(err))
		http.Error(res, "ошибка при записи ответа", http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) APICreate(res http.ResponseWriter, req *http.Request) {
	var decoded model.CreateURLRequest
	res.Header().Set("Content-Type", "application/json")

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		h.writeJSONError(res, "ошибка чтения тела запроса: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = easyjson.Unmarshal(reqBody, &decoded)
	if err != nil {
		h.writeJSONError(res, "ошибка парсинга запроса: "+err.Error(), http.StatusBadRequest)
		return
	}
	if decoded.URL == "" {
		h.writeJSONError(res, "URL не задан", http.StatusBadRequest)
		return
	}
	storedURL, err := h.urlService.SaveURL(req.Context(), decoded.URL, userIDFromContext(req.Context()))
	if err != nil {
		var urlExists *model.ErrURLAlreadyExists
		if errors.As(err, &urlExists) {
			resp := model.CreateURLResponse{Result: h.config.BaseURL + "/" + urlExists.URL.Short}
			jsonBody, _ := easyjson.Marshal(resp)
			res.WriteHeader(http.StatusConflict)
			if _, err = res.Write(jsonBody); err != nil {
				h.logger.Error("ошибка при записи ответа", zap.Error(err))
			}
			return
		}
		h.writeJSONError(res, "не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
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
	_, err = res.Write(jsonBody)
	if err != nil {
		h.logger.Error("ошибка при записи ответа", zap.Error(err))
		http.Error(res, "ошибка при записи ответа", http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) APICreateBatch(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		h.writeJSONError(res, "ошибка чтения тела запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	var requestItems []model.CreateURLBatchRequestItem
	if err := json.Unmarshal(reqBody, &requestItems); err != nil {
		h.writeJSONError(res, "ошибка парсинга запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	originalURLs := make([]string, 0, len(requestItems))
	for _, item := range requestItems {
		originalURLs = append(originalURLs, item.OriginalURL)
	}

	ctx := req.Context()
	savedURLs, err := h.urlService.SaveManyURL(ctx, originalURLs, userIDFromContext(ctx))
	if err != nil {
		var urlExists *model.ErrURLAlreadyExists
		if errors.As(err, &urlExists) {
			h.writeJSONError(res,
				"URL уже существует: original_url: "+urlExists.URL.Original+", short_url: "+urlExists.URL.Short,
				http.StatusConflict,
			)
			return
		}
		h.writeJSONError(res, "ошибка при сохранении URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	responseItems := make([]model.CreateURLBatchResponseItem, 0, len(requestItems))
	for i, savedURL := range savedURLs {
		responseItems = append(responseItems, model.CreateURLBatchResponseItem{
			CorrelationID: requestItems[i].CorrelationID,
			ShortURL:      h.config.BaseURL + "/" + savedURL.Short,
		})
	}

	jsonBody, err := json.Marshal(responseItems)
	if err != nil {
		h.logger.Error("ошибка при подготовке ответа", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(jsonBody)
	if err != nil {
		h.logger.Error("ошибка при записи ответа", zap.Error(err))
		http.Error(res, "ошибка при записи ответа", http.StatusInternalServerError)
		return
	}
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

	if url.DeletedFlag {
		http.Error(res, "URL удалён", http.StatusGone)
		return
	}

	http.Redirect(res, req, url.Original, http.StatusTemporaryRedirect)
}

func (h *URLHandler) Ping(res http.ResponseWriter, req *http.Request) {
	if err := h.pinger.Ping(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (h *URLHandler) GetUserURLs(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	userID := userIDFromContext(req.Context())
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.urlService.GetAllByUserID(req.Context(), userID)
	if err != nil {
		h.writeJSONError(res, "ошибка получения URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	type responseItem struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	items := make([]responseItem, 0, len(urls))
	for _, u := range urls {
		items = append(items, responseItem{
			ShortURL:    h.config.BaseURL + "/" + u.Short,
			OriginalURL: u.Original,
		})
	}

	body, err := json.Marshal(items)
	if err != nil {
		h.writeJSONError(res, "ошибка сериализации", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(body)
}

func (h *URLHandler) APIDeleteURLs(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	userID := userIDFromContext(req.Context())
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		h.writeJSONError(res, "ошибка чтения тела запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	var urlIDs []string
	if err := json.Unmarshal(reqBody, &urlIDs); err != nil {
		h.writeJSONError(res, "ошибка парсинга запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.urlService.DeleteURLs(req.Context(), urlIDs, userID); err != nil {
		h.writeJSONError(res, "ошибка при удалении URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)

}
