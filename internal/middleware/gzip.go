package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		contentType := r.Header.Get("Content-Type")
		gzippableContentType := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")

		if supportsGzip && gzippableContentType {
			gzw := NewGzipCompressWriter(w)
			ow = gzw
			defer gzw.Close()
		}

		contentEncodign := r.Header.Get("Content-Encoding")
		isRequestGzipeed := strings.Contains(contentEncodign, "gzip")
		if isRequestGzipeed {
			cr, err := NewGzipCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

type gzipCompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewGzipCompressWriter(w http.ResponseWriter) *gzipCompressWriter {
	return &gzipCompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (gz *gzipCompressWriter) Write(p []byte) (int, error) {
	return gz.zw.Write(p)
}
func (gz *gzipCompressWriter) Header() http.Header {
	return gz.w.Header()
}
func (gz *gzipCompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gz.w.Header().Set("Content-Encoding", "gzip")
	}
	gz.w.WriteHeader(statusCode)
}
func (gz *gzipCompressWriter) Close() error {
	return gz.zw.Close()
}

type gzipCompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewGzipCompressReader(r io.ReadCloser) (*gzipCompressReader, error) {

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipCompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (gz *gzipCompressReader) Read(p []byte) (int, error) {
	return gz.zr.Read(p)
}
func (gz *gzipCompressReader) Close() error {
	if err := gz.r.Close(); err != nil {
		return err
	}
	return gz.r.Close()
}
