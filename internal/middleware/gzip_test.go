package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestGzipCompression_ResponseWriter(t *testing.T) {
	const responseBody = `{"result":"ok"}`

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	})

	handler := GzipCompression(nextHandler)

	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		wantGzip       bool
	}{
		{
			name:           "сжимает application/json",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			wantGzip:       true,
		},
		{
			name:           "сжимает text/html",
			acceptEncoding: "gzip",
			contentType:    "text/html",
			wantGzip:       true,
		},
		{
			name:           "не сжимает без Accept-Encoding",
			acceptEncoding: "",
			contentType:    "application/json",
			wantGzip:       false,
		},
		{
			name:           "не сжимает text/plain",
			acceptEncoding: "gzip",
			contentType:    "text/plain",
			wantGzip:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			req.Header.Set("Content-Type", tt.contentType)
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			assert.Equal(t, res.Code, http.StatusOK)

			if tt.wantGzip {
				assert.Equal(t, res.Header().Get("Content-Encoding"), "gzip")

				gr, err := gzip.NewReader(res.Body)
				if err != nil {
					t.Fatalf("не удалось создать gzip reader: %v", err)
				}
				defer gr.Close()

				body, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("не удалось прочитать gzip тело: %v", err)
				}
				assert.Equal(t, string(body), responseBody)
			} else {
				assert.Equal(t, res.Header().Get("Content-Encoding"), "")
				assert.Equal(t, res.Body.String(), responseBody)
			}
		})
	}
}

func TestGzipCompression_RequestDecompression(t *testing.T) {
	const originalBody = `{"url":"http://example.com"}`

	tests := []struct {
		name            string
		contentEncoding string
		buildBody       func() *bytes.Buffer
	}{
		{
			name:            "распаковывает gzip тело запроса",
			contentEncoding: "gzip",
			buildBody: func() *bytes.Buffer {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				gz.Write([]byte(originalBody))
				gz.Close()
				return &buf
			},
		},
		{
			name:            "не трогает обычное тело",
			contentEncoding: "",
			buildBody: func() *bytes.Buffer {
				return bytes.NewBufferString(originalBody)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody string

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				receivedBody = string(b)
				w.WriteHeader(http.StatusCreated)
			})

			handler := GzipCompression(nextHandler)

			req := httptest.NewRequest(http.MethodPost, "/", tt.buildBody())
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			assert.Equal(t, receivedBody, originalBody)
		})
	}
}
