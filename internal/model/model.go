package model

import "time"

type URL struct {
	Original    string     `json:"original_url" db:"original"`
	Short       string     `json:"short_url"    db:"short"`
	UserID      string     `json:"user_id"      db:"user_id"`
	DeletedFlag bool       `json:"-"            db:"is_deleted"`
	DeletedAt   *time.Time `json:"-"            db:"deleted_at"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
}

type APIErrorResponse struct {
	Error string `json:"error"`
}

type CreateURLBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type CreateURLBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ErrURLAlreadyExists struct {
	URL URL
}

func (e *ErrURLAlreadyExists) Error() string {
	return "URL уже существует: " + e.URL.Short
}

type ErrURLNotFound struct {
	Short string
}

func (e *ErrURLNotFound) Error() string {
	return "URL не найден: " + e.Short
}

type ErrURLDeleted struct {
	Short string
}

func (e *ErrURLDeleted) Error() string {
	return "URL удален: " + e.Short
}
