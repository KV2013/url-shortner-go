package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler/mocks"
	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/middleware"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/magiconair/properties/assert"
	"go.uber.org/mock/gomock"
)

func TestCreate(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
		response    string
	}

	tests := []struct {
		name          string
		request       string
		url           string
		saveURLError  error
		storedURL     *model.URL
		expectedError bool
		config        *config.Config
		want          want
	}{
		{
			name:    "201 Created - successful save",
			request: "http://localhost:8080",
			url:     "http://example.com",
			storedURL: &model.URL{
				Short:    "abc123",
				Original: "http://example.com",
			},
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				response:    "http://localhost:8080/abc123", // ожидаемый короткий URL
			},
		},
		{
			name:    "201 Created - with different baseurl",
			request: "http://localhost:8080",
			url:     "http://example.com",
			storedURL: &model.URL{
				Short:    "abc123",
				Original: "http://example.com",
			},
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "https://foo.bar:45000",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				response:    "https://foo.bar:45000/abc123", // ожидаемый короткий URL
			},
		},
		{
			name:    "400 Bad Request - save error",
			request: "http://localhost:8080",
			url:     "http://example.com",
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			saveURLError:  errors.New("failed to save"),
			expectedError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "400 Bad Request - empty URL",
			request: "http://localhost:8080",
			url:     "",
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			expectedError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	Logger, loggerErr := logger.New("debug")
	if loggerErr != nil {
		t.Fatalf("не удалось создать логгер: %v", loggerErr)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)

			if tt.url != "" && !tt.expectedError {
				mockService.EXPECT().
					SaveURL(gomock.Any(), tt.url, gomock.Any()).
					Return(tt.storedURL, tt.saveURLError)
			} else if tt.url != "" {
				mockService.EXPECT().
					SaveURL(gomock.Any(), tt.url, gomock.Any()).
					Return(nil, tt.saveURLError)
			}

			handler := New(mockService, nil, tt.config, Logger)

			body := strings.NewReader(tt.url)
			req := httptest.NewRequest(http.MethodPost, tt.request, body)
			res := httptest.NewRecorder()

			handler.Create(res, req)

			// Проверяем результаты
			assert.Equal(t, res.Code, tt.want.statusCode)
			if tt.want.contentType != "" {
				assert.Equal(t, res.Header().Get("Content-Type"), tt.want.contentType)
			}

			if tt.want.response != "" {
				actualResponse := strings.TrimSpace(res.Body.String())
				assert.Equal(t, actualResponse, tt.want.response)
			}
		})
	}
}

func TestAPICreate(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}

	tests := []struct {
		name          string
		request       string
		url           string
		body          string
		saveURLError  error
		storedURL     *model.URL
		expectedError bool
		config        *config.Config
		want          want
	}{
		{
			name:    "201 Created - successful save",
			request: "http://localhost:8080/api/shorten",
			url:     "http://example.com",
			body:    `{"url":"http://example.com"}`,
			storedURL: &model.URL{
				Short:    "abc123",
				Original: "http://example.com",
			},
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
				response:    `{"result":"http://localhost:8080/abc123"}`,
			},
		},
		// ---
		{
			name:    "201 Created - with different baseurl",
			request: "http://localhost:8080/api/shorten",
			url:     "http://example.com",
			body:    `{"url":"http://example.com"}`,
			storedURL: &model.URL{
				Short:    "abc123",
				Original: "http://example.com",
			},
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "https://foo.bar:45000",
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
				response:    `{"result":"https://foo.bar:45000/abc123"}`,
			},
		},
		{
			name:    "400 Bad Request - save error",
			request: "http://localhost:8080/api/shorten",
			url:     "http://example.com",
			body:    `{"url":"http://example.com"}`,
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			saveURLError:  errors.New("failed to save"),
			expectedError: true,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:    "400 Bad Request - empty URL",
			request: "http://localhost:8080/api/shorten",
			url:     "",
			body:    `{"url":""}`,
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
			},
			expectedError: true,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
			},
		},

		// ---
	}

	Logger, loggerErr := logger.New("debug")
	if loggerErr != nil {
		t.Fatalf("не удалось создать логгер: %v", loggerErr)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)

			if tt.body != "" && !tt.expectedError {
				mockService.EXPECT().
					SaveURL(gomock.Any(), tt.url, gomock.Any()).
					Return(tt.storedURL, tt.saveURLError)
			} else if tt.url != "" {
				mockService.EXPECT().
					SaveURL(gomock.Any(), tt.url, gomock.Any()).
					Return(nil, tt.saveURLError)
			}

			handler := New(mockService, nil, tt.config, Logger)

			body := strings.NewReader(tt.body)
			req := httptest.NewRequest(http.MethodPost, tt.request, body)
			res := httptest.NewRecorder()

			handler.APICreate(res, req)

			// Проверяем результаты
			assert.Equal(t, res.Code, tt.want.statusCode)
			if tt.want.contentType != "" {
				assert.Equal(t, res.Header().Get("Content-Type"), tt.want.contentType)
			}

			if tt.want.response != "" {
				actualResponse := strings.TrimSpace(res.Body.String())
				assert.Equal(t, actualResponse, tt.want.response)
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	cfg := &config.Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080"}

	tests := []struct {
		name               string
		id                 string
		foundURL           *model.URL
		getURLErr          error
		expectsCallGetByID bool
		expectedCode       int
		expectedLocation   string
	}{
		{
			name: "307 Temporary Redirect - URL found",
			id:   "abc123",
			foundURL: &model.URL{
				Short:    "abc123",
				Original: "https://example.com",
			},
			getURLErr:          nil,
			expectsCallGetByID: true,
			expectedCode:       http.StatusTemporaryRedirect,
			expectedLocation:   "https://example.com",
		},
		{
			name:               "404 Not Found - URL not found",
			id:                 "unknown-id",
			getURLErr:          &model.ErrURLNotFound{Short: "unknown-id"},
			expectsCallGetByID: true,
			expectedCode:       http.StatusNotFound,
		},
		{
			name:               "400 Bad Request - empty id",
			id:                 "",
			expectsCallGetByID: false,
			expectedCode:       http.StatusBadRequest,
		},
		{
			name:               "410 Gone - URL deleted",
			id:                 "deleted1",
			getURLErr:          &model.ErrURLDeleted{Short: "deleted1"},
			expectsCallGetByID: true,
			expectedCode:       http.StatusGone,
		},
	}

	Logger, loggerErr := logger.New("debug")
	if loggerErr != nil {
		t.Fatalf("не удалось создать логгер: %v", loggerErr)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)

			if tt.expectsCallGetByID {
				mockService.EXPECT().
					GetByID(gomock.Any(), tt.id).
					Return(tt.foundURL, tt.getURLErr)
			}

			handler := New(mockService, nil, cfg, Logger)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			res := httptest.NewRecorder()

			handler.Redirect(res, req)

			assert.Equal(t, res.Code, tt.expectedCode)
			if tt.expectedLocation != "" {
				assert.Equal(t, res.Header().Get("Location"), tt.expectedLocation)
			}
		})
	}
}

func ctxWithUserID(userID string) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDContextKey, userID)
}

func TestGetUserURLs(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	tests := []struct {
		name         string
		userID       string
		serviceURLs  []model.URL
		serviceErr   error
		expectedCode int
		expectedBody string
	}{
		{
			name:   "200 OK - has URLs",
			userID: "user-1",
			serviceURLs: []model.URL{
				{Short: "abc", Original: "https://example.com", UserID: "user-1"},
			},
			expectedCode: http.StatusOK,
			expectedBody: `[{"short_url":"http://localhost:8080/abc","original_url":"https://example.com"}]`,
		},
		{
			name:         "204 No Content - no URLs",
			userID:       "user-2",
			serviceURLs:  []model.URL{},
			expectedCode: http.StatusNoContent,
		},
		{
			name:         "401 Unauthorized - no userID in context",
			userID:       "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "500 Internal Server Error - service error",
			userID:       "user-3",
			serviceErr:   errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	Logger, _ := logger.New("debug")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)
			if tt.userID != "" {
				mockService.EXPECT().
					GetAllByUserID(gomock.Any(), tt.userID).
					Return(tt.serviceURLs, tt.serviceErr)
			}

			h := New(mockService, nil, cfg, Logger)
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			req = req.WithContext(ctxWithUserID(tt.userID))
			res := httptest.NewRecorder()

			h.GetUserURLs(res, req)

			assert.Equal(t, res.Code, tt.expectedCode)
			if tt.expectedBody != "" {
				assert.Equal(t, strings.TrimSpace(res.Body.String()), tt.expectedBody)
			}
		})
	}
}

func TestAPIDeleteURLs(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	tests := []struct {
		name         string
		userID       string
		body         string
		ids          []string
		serviceErr   error
		expectedCode int
	}{
		{
			name:         "202 Accepted - successful delete",
			userID:       "user-1",
			body:         `["abc","def"]`,
			ids:          []string{"abc", "def"},
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "401 Unauthorized - no userID",
			userID:       "",
			body:         `["abc"]`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "400 Bad Request - invalid JSON",
			userID:       "user-1",
			body:         `not-json`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "500 Internal Server Error - service error",
			userID:       "user-1",
			body:         `["abc"]`,
			ids:          []string{"abc"},
			serviceErr:   errors.New("delete failed"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	Logger, _ := logger.New("debug")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)
			if tt.userID != "" && tt.ids != nil {
				mockService.EXPECT().
					DeleteURLs(gomock.Any(), tt.ids, tt.userID).
					Return(tt.serviceErr)
			}

			h := New(mockService, nil, cfg, Logger)
			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(tt.body))
			req = req.WithContext(ctxWithUserID(tt.userID))
			res := httptest.NewRecorder()

			h.APIDeleteURLs(res, req)

			assert.Equal(t, res.Code, tt.expectedCode)
		})
	}
}
