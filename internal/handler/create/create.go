package create

import (
	"io"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/repository"
	"github.com/KV2013/url-shortner-go/internal/service/random"
)

/*
*
Сервер должен быть доступен по адресу http://localhost:8080 и предоставлять два эндпоинта:
1. Эндпоинт с методом POST и путём /. Сервер принимает в теле запроса строку URL как
text/plain и возвращает ответ с кодом 201 и сокращённым URL как text/plain

	Пример запроса к серверу:

		POST / HTTP/1.1
		Host: localhost:8080
		Content-Type: text/plain

	https://practicum.yandex.ru/

	Пример ответа от сервера:

		HTTP/1.1 201 Created
		Content-Type: text/plain
		Content-Length: 30

		http://localhost:8080/EwHXdJfB
*/
func New(urls repository.URLCollection, baseURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

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

		_, exists := urls.FindByURL(reqURL)
		if exists {
			http.Error(res, "Url already exists", http.StatusBadRequest)
			return
		}

		urlID := random.NewRandomString(10)
		urls.Set(model.URL{
			Original: reqURL,
			Short:    urlID,
		})

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, baseURL+urlID)
	}
}
