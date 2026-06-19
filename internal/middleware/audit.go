package middleware

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/KV2013/url-shortner-go/internal/config"
	"go.uber.org/zap"
)

type AuditEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type AuditEventOption func(*AuditEvent)

func WithAction(action string) AuditEventOption {
	return func(e *AuditEvent) { e.Action = action }
}

func WithUserID(userID string) AuditEventOption {
	return func(e *AuditEvent) { e.UserID = userID }
}

func WithURL(url string) AuditEventOption {
	return func(e *AuditEvent) { e.URL = url }
}

func NewAuditEvent(opts ...AuditEventOption) AuditEvent {
	e := AuditEvent{TS: time.Now().Unix()}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

type AuditObserver interface {
	Notify(ctx context.Context, event AuditEvent) error
}

type LocalAuditor struct {
	file   *os.File
	writer *bufio.Writer
	mu     sync.Mutex
}

func NewLocalAuditor(filePath string) (*LocalAuditor, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &LocalAuditor{
		file:   f,
		writer: bufio.NewWriter(f),
	}, nil
}

func (a *LocalAuditor) Notify(_ context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.writer.Write(data)
	a.writer.WriteByte('\n')
	return a.writer.Flush()
}

type RemoteAuditor struct {
	url    string
	client *http.Client
}

func NewRemoteAuditor(url string) *RemoteAuditor {
	return &RemoteAuditor{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *RemoteAuditor) Notify(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

type responseWriterDecorator struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перехватывает статус-код до отправки клиенту, чтобы middleware мог
// определить успешность запроса (201/307) и решить, создавать ли аудит-событие.
// http.ResponseWriter не позволяет прочитать статус после вызова WriteHeader.
func (rw *responseWriterDecorator) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestAuditor(ctx context.Context, cfg *config.Config, logger *zap.Logger) func(http.Handler) http.Handler {
	var observers []AuditObserver

	if cfg.AuditFile != "" {
		la, err := NewLocalAuditor(cfg.AuditFile)
		if err != nil {
			logger.Error("RequestAuditor: failed to create LocalAuditor: %v", zap.Error(err))
		} else {
			observers = append(observers, la)
			logger.Debug("Local request auditor added", zap.String("AuditFile", cfg.AuditFile))
		}
	}

	if cfg.AuditURL != "" {
		observers = append(observers, NewRemoteAuditor(cfg.AuditURL))
		logger.Debug("Remote request auditor added", zap.String("AuditURL", cfg.AuditURL))
	}

	if len(observers) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var requestURL string

			isShortenPlain := r.Method == http.MethodPost && r.URL.Path == "/"
			isShortenAPI := r.Method == http.MethodPost && r.URL.Path == "/api/shorten"
			isFollow := r.Method == http.MethodGet &&
				r.URL.Path != "/" &&
				r.URL.Path != "/ping" &&
				!strings.HasPrefix(r.URL.Path, "/api/")

			if isShortenPlain {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					requestURL = string(body)
					r.Body = io.NopCloser(bytes.NewReader(body))
				}
			}

			if isShortenAPI {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					r.Body = io.NopCloser(bytes.NewReader(body))
					var req struct {
						URL string `json:"url"`
					}
					if json.Unmarshal(body, &req) == nil {
						requestURL = req.URL
					}
				}
			}

			rw := &responseWriterDecorator{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			var action string
			switch {
			case (isShortenPlain || isShortenAPI) && rw.statusCode == http.StatusCreated:
				action = "shorten"
			case isFollow && rw.statusCode == http.StatusTemporaryRedirect:
				action = "follow"
				if requestURL == "" {
					requestURL = rw.Header().Get("Location")
				}
			}

			if action == "" {
				return
			}

			userID, _ := r.Context().Value(UserIDContextKey).(string)

			event := NewAuditEvent(
				WithAction(action),
				WithUserID(userID),
				WithURL(requestURL),
			)

			for _, obs := range observers {
				go obs.Notify(ctx, event)
			}
		})
	}
}
