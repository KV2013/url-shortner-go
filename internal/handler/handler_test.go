package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KV2013/url-shortner-go/internal/config"
	"github.com/KV2013/url-shortner-go/internal/handler/mocks"
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
		want          want
	}{
		{
			name:    "201 Created - successful save",
			request: "https://localhost:8080",
			url:     "https://example.com",
			storedURL: &model.URL{
				Short:    "abc123",
				Original: "https://example.com",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				response:    "https://localhost:8080/abc123", // ожидаемый короткий URL
			},
		},
		{
			name:          "400 Bad Request - save error",
			request:       "https://localhost:8080",
			url:           "https://example.com",
			saveURLError:  errors.New("failed to save"),
			expectedError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "400 Bad Request - empty URL",
			request:       "https://localhost:8080",
			url:           "",
			expectedError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)

			if tt.url != "" && !tt.expectedError {
				mockService.EXPECT().
					SaveURL(tt.url).
					Return(tt.storedURL, tt.saveURLError)
			} else if tt.url != "" {
				mockService.EXPECT().
					SaveURL(tt.url).
					Return(nil, tt.saveURLError)
			}

			config := config.NewConfig()
			handler := New(mockService, config)

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

func TestRedirect(t *testing.T) {
	tests := []struct {
		name               string
		id                 string
		getURLError        error
		foundURL           *model.URL
		exists             bool
		expectsCallGetByID bool
		expectedCode       int
	}{
		{
			name: "307 Temporary Redirect - URL found",
			id:   "abc123",
			foundURL: &model.URL{
				Short:    "abc123",
				Original: "https://example.com",
			},
			exists:             true,
			expectsCallGetByID: true,
			expectedCode:       http.StatusTemporaryRedirect,
		},
		{
			name:               "404 Not Found - URL not found",
			id:                 "unknown-id",
			exists:             false,
			expectsCallGetByID: true,
			expectedCode:       http.StatusNotFound,
		},
		{
			name:               "400 Bad Request - empty id",
			id:                 "",
			exists:             false,
			expectsCallGetByID: false,
			expectedCode:       http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём контроллер моков
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)

			if tt.expectsCallGetByID {
				mockService.EXPECT().
					GetByID(tt.id).
					Return(tt.foundURL, tt.exists)
			}

			config := config.NewConfig()
			handler := New(mockService, config)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			req.SetPathValue("id", tt.id)

			res := httptest.NewRecorder()

			handler.Redirect(res, req)

			// Проверяем результаты
			if res.Code != tt.expectedCode {
				assert.Equal(t, res.Code, tt.expectedCode)
			}

			if tt.exists {
				location := res.Header().Get("Location")
				assert.Equal(t, location, tt.foundURL.Original)
			}
		})
	}
}
