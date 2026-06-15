package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler/mocks"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/KV2013/url-shortner-go/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

var benchCfg = &config.Config{
	ServerAddress: "localhost:8080",
	BaseURL:       "http://localhost:8080",
}

var benchLogger = zap.NewNop()

func BenchmarkCreate(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockService := mocks.NewMockURLService(ctrl)
	mockService.EXPECT().
		SaveURL(gomock.Any(), "http://example.com", gomock.Any()).
		Return(&model.URL{Short: "abc123", Original: "http://example.com"}, nil).
		AnyTimes()

	h := New(mockService, nil, benchCfg, benchLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := strings.NewReader("http://example.com")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		rec := httptest.NewRecorder()
		h.Create(rec, req)
	}
}

func BenchmarkAPICreate(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockService := mocks.NewMockURLService(ctrl)
	mockService.EXPECT().
		SaveURL(gomock.Any(), "http://example.com", gomock.Any()).
		Return(&model.URL{Short: "abc123", Original: "http://example.com"}, nil).
		AnyTimes()

	h := New(mockService, nil, benchCfg, benchLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := strings.NewReader(`{"url":"http://example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		rec := httptest.NewRecorder()
		h.APICreate(rec, req)
	}
}

func BenchmarkAPICreateBatch(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockService := mocks.NewMockURLService(ctrl)
	mockService.EXPECT().
		SaveManyURL(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, urls []string, _ string) ([]model.URL, error) {
			result := make([]model.URL, len(urls))
			for i, u := range urls {
				result[i] = model.URL{Short: "abc123", Original: u}
			}
			return result, nil
		}).
		AnyTimes()

	h := New(mockService, nil, benchCfg, benchLogger)

	batchBody := `[` +
		`{"correlation_id":"1","original_url":"http://example.com/1"},` +
		`{"correlation_id":"2","original_url":"http://example.com/2"},` +
		`{"correlation_id":"3","original_url":"http://example.com/3"},` +
		`{"correlation_id":"4","original_url":"http://example.com/4"},` +
		`{"correlation_id":"5","original_url":"http://example.com/5"},` +
		`{"correlation_id":"6","original_url":"http://example.com/6"},` +
		`{"correlation_id":"7","original_url":"http://example.com/7"},` +
		`{"correlation_id":"8","original_url":"http://example.com/8"},` +
		`{"correlation_id":"9","original_url":"http://example.com/9"},` +
		`{"correlation_id":"10","original_url":"http://example.com/10"}` +
		`]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := strings.NewReader(batchBody)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
		rec := httptest.NewRecorder()
		h.APICreateBatch(rec, req)
	}
}

func BenchmarkAPICreateBatch_Size(b *testing.B) {
	sizes := []int{1, 5, 10, 50}

	for _, sz := range sizes {
		b.Run(strconv.Itoa(sz), func(b *testing.B) {
			ctrl := gomock.NewController(b)
			mockService := mocks.NewMockURLService(ctrl)
			mockService.EXPECT().
				SaveManyURL(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, urls []string, _ string) ([]model.URL, error) {
					result := make([]model.URL, len(urls))
					for i, u := range urls {
						result[i] = model.URL{Short: "abc123", Original: u}
					}
					return result, nil
				}).
				AnyTimes()

			h := New(mockService, nil, benchCfg, benchLogger)

			batchBody := buildBatchBody(sz)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				body := strings.NewReader(batchBody)
				req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
				rec := httptest.NewRecorder()
				h.APICreateBatch(rec, req)
			}
		})
	}
}

func buildBatchBody(size int) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < size; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{"correlation_id":"`)
		sb.WriteByte(byte('a' + (i % 26)))
		sb.WriteString(`","original_url":"http://example.com/`)
		sb.WriteByte(byte('0' + (i % 10)))
		sb.WriteString(`"}`)
	}
	sb.WriteByte(']')
	return sb.String()
}

func BenchmarkRedirect(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockService := mocks.NewMockURLService(ctrl)
	mockService.EXPECT().
		GetByID(gomock.Any(), "abc123").
		Return(&model.URL{Short: "abc123", Original: "https://example.com"}, nil).
		AnyTimes()

	h := New(mockService, nil, benchCfg, benchLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		req.SetPathValue("id", "abc123")
		rec := httptest.NewRecorder()
		h.Redirect(rec, req)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	urls := make([]model.URL, 10)
	for i := 0; i < 10; i++ {
		urls[i] = model.URL{Short: "abc123", Original: "https://example.com/", UserID: "user-1"}
	}

	ctrl := gomock.NewController(b)
	mockService := mocks.NewMockURLService(ctrl)
	mockService.EXPECT().
		GetAllByUserID(gomock.Any(), "user-1").
		Return(urls, nil).
		AnyTimes()

	h := New(mockService, nil, benchCfg, benchLogger)

	ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "user-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		h.GetUserURLs(rec, req)
	}
}

func BenchmarkAPIDeleteURLs(b *testing.B) {
	deleteIDs := []string{"abc", "def", "ghi", "jkl", "mno"}

	ctrl := gomock.NewController(b)
	mockService := mocks.NewMockURLService(ctrl)
	mockService.EXPECT().
		DeleteURLs(gomock.Any(), deleteIDs, "user-1").
		Return(nil).
		AnyTimes()

	h := New(mockService, nil, benchCfg, benchLogger)

	ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "user-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := strings.NewReader(`["abc","def","ghi","jkl","mno"]`)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body).WithContext(ctx)
		rec := httptest.NewRecorder()
		h.APIDeleteURLs(rec, req)
	}
}
