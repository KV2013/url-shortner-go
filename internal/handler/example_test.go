package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/model"
)

// exampleURLService — заглушка URLService для демонстрации в godoc-примерах.
type exampleURLService struct{}

func (s *exampleURLService) SaveURL(_ context.Context, url, _ string) (*model.URL, error) {
	return &model.URL{Short: "abc123", Original: url}, nil
}

func (s *exampleURLService) SaveManyURL(_ context.Context, urls []string, _ string) ([]model.URL, error) {
	result := make([]model.URL, len(urls))
	for i, u := range urls {
		result[i] = model.URL{Short: "short" + strconv.Itoa(i), Original: u}
	}
	return result, nil
}

func (s *exampleURLService) GetByID(_ context.Context, id string) (*model.URL, error) {
	if id == "abc123" {
		return &model.URL{Short: "abc123", Original: "https://example.com/page"}, nil
	}
	return nil, &model.ErrURLNotFound{Short: id}
}

func (s *exampleURLService) GetAllByUserID(_ context.Context, _ string) ([]model.URL, error) {
	return []model.URL{
		{Short: "abc", Original: "https://example.com/1"},
		{Short: "def", Original: "https://example.com/2"},
	}, nil
}

func (s *exampleURLService) DeleteURLs(_ context.Context, _ []string, _ string) error {
	return nil
}

// ExampleURLHandler_Create демонстрирует создание короткой ссылки через POST /.
// В теле запроса передаётся оригинальный URL как text/plain.
func ExampleURLHandler_Create() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	h := New(svc, nil, cfg, log)

	body := strings.NewReader("https://example.com/very-long-url")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = req.WithContext(ctxWithUserID("user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 201
	// text/plain
	// http://localhost:8080/abc123
}

// ExampleURLHandler_APICreate демонстрирует создание короткой ссылки через POST /api/shorten.
// Принимает JSON с полем url, возвращает JSON с коротким URL.
func ExampleURLHandler_APICreate() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	h := New(svc, nil, cfg, log)

	body := strings.NewReader(`{"url":"https://example.com/very-long-url"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	req = req.WithContext(ctxWithUserID("user-1"))
	rec := httptest.NewRecorder()

	h.APICreate(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 201
	// application/json
	// {"result":"http://localhost:8080/abc123"}
}

// ExampleURLHandler_APICreateBatch демонстрирует пакетное создание коротких ссылок
// через POST /api/shorten/batch. Каждый элемент связывает correlation_id с short_url.
func ExampleURLHandler_APICreateBatch() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	h := New(svc, nil, cfg, log)

	body := strings.NewReader(`[
		{"correlation_id":"1","original_url":"https://example.com/first"},
		{"correlation_id":"2","original_url":"https://example.com/second"}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
	req = req.WithContext(ctxWithUserID("user-1"))
	rec := httptest.NewRecorder()

	h.APICreateBatch(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 201
	// application/json
	// [{"correlation_id":"1","short_url":"http://localhost:8080/short0"},{"correlation_id":"2","short_url":"http://localhost:8080/short1"}]
}

// ExampleURLHandler_Redirect демонстрирует перенаправление на оригинальный URL
// через GET /{id}. Успешный запрос возвращает 307 TemporaryRedirect.
func ExampleURLHandler_Redirect() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	h := New(svc, nil, cfg, log)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	req.SetPathValue("id", "abc123")
	rec := httptest.NewRecorder()

	h.Redirect(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Location"))
	// Output:
	// 307
	// https://example.com/page
}

// ExampleURLHandler_GetUserURLs демонстрирует получение всех URL пользователя
// через GET /api/user/urls. Требует userID в контексте запроса.
func ExampleURLHandler_GetUserURLs() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	h := New(svc, nil, cfg, log)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(ctxWithUserID("user-1"))
	rec := httptest.NewRecorder()

	h.GetUserURLs(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 200
	// application/json
	// [{"short_url":"http://localhost:8080/abc","original_url":"https://example.com/1"},{"short_url":"http://localhost:8080/def","original_url":"https://example.com/2"}]
}

// ExampleURLHandler_APIDeleteURLs демонстрирует soft-delete URL пользователя
// через DELETE /api/user/urls. Принимает JSON-массив идентификаторов.
func ExampleURLHandler_APIDeleteURLs() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	h := New(svc, nil, cfg, log)

	body := strings.NewReader(`["abc123","def456"]`)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	req = req.WithContext(ctxWithUserID("user-1"))
	rec := httptest.NewRecorder()

	h.APIDeleteURLs(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 202
}

// ExampleURLHandler_Ping демонстрирует проверку работоспособности хранилища
// через GET /ping.
func ExampleURLHandler_Ping() {
	svc := &exampleURLService{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.New("error")
	pinger := &examplePinger{}
	h := New(svc, pinger, cfg, log)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	h.Ping(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}

// examplePinger — заглушка Pinger для демонстрации Ping.
type examplePinger struct{}

func (p *examplePinger) Ping() error {
	return nil
}
