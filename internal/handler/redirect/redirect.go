package redirect

import (
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/repository"
)

/*
2. Эндпоинт с методом GET и путём /{id}, где id — идентификатор сокращённого URL (например,
/EwHXdJfB). В случае успешной обработки запроса сервер возвращает ответ с кодом 307 и
оригинальным URL в HTTP-заголовке Location.

Пример запроса к серверу:

	GET /EwHXdJfB HTTP/1.1
	Host: localhost:8080
	Content-Type: text/plain

Пример ответа от сервера:

	HTTP/1.1 307 Temporary Redirect
	Location: https://practicum.yandex.ru/
*/
func NewRedirectHandler(urls repository.URLCollection) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		urlID := req.PathValue("id")
		if urlID == "" {
			http.Error(res, "empty id", http.StatusBadRequest)
			return
		}

		url, exists := urls.FindByID(urlID)
		if !exists {
			http.NotFound(res, req)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Redirect(res, req, url.Original, http.StatusTemporaryRedirect)
	}
}
