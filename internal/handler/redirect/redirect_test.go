package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/repository"
	"github.com/magiconair/properties/assert"
)

func TestNewRedirectHandler(t *testing.T) {

	url1 := model.URL{
		Original: "https://foo.bar",
		Short:    "q123",
	}
	urls1 := repository.NewURLCollection()
	urls1.Set(url1)
	baseURL := "http://localhost:8080/"

	type want struct {
		contentType string
		statusCode  int
		Location    string
	}
	tests := []struct {
		name    string
		request string
		want    want
		urls    repository.URLCollection
		baseURL string
	}{
		{
			name:    "status 307",
			request: baseURL + url1.Short,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusTemporaryRedirect,
				Location:    url1.Original,
			},
			urls: urls1,
		},
		{
			name:    "not found",
			request: baseURL + "qwe",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				Location:    "",
			},
			urls: urls1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.request, nil)
			res := httptest.NewRecorder()
			h := NewRedirectHandler(tt.urls)
			// h(res, req)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", h) // например, "GET /foo/{id}"

			// Тестируем через маршрутизатор
			mux.ServeHTTP(res, req)

			result := res.Result()

			assert.Equal(t, result.StatusCode, tt.want.statusCode, "request:"+tt.request)
			assert.Equal(t, result.Header.Get("Content-Type"), tt.want.contentType)
			assert.Equal(t, result.Header.Get("Location"), tt.want.Location)
		})
	}
}
