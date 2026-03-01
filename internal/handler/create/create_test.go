package create

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/repository"
	"github.com/magiconair/properties/assert"
)

func TestNew(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	urls1 := repository.NewURLCollection()
	urls1.Set(model.URL{
		Original: "https://foo.bar",
		Short:    "123",
	})
	baseURL := "http://localhost:8080"

	tests := []struct {
		name    string
		request string
		url     string
		want    want
		urls    repository.URLCollection
		baseURL string
	}{
		{
			name:    "201 Created",
			request: "/",
			url:     "example.com",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
			urls:    repository.NewURLCollection(),
			baseURL: baseURL,
		},
		{
			name:    "400 Already exists",
			request: "/",
			url:     "https://foo.bar",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
			urls:    urls1,
			baseURL: baseURL,
		},
		{
			name:    "400 no url provided",
			request: "/",
			url:     "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
			urls:    urls1,
			baseURL: baseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.url)
			req := httptest.NewRequest(http.MethodPost, tt.request, body)
			res := httptest.NewRecorder()
			h := http.HandlerFunc(New(tt.urls, tt.baseURL))
			h(res, req)

			result := res.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
			assert.Equal(t, result.Header.Get("Content-Type"), tt.want.contentType)
		})
	}
}
