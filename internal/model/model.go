package model

type URL struct {
	Original string `json:"original_url"`
	Short    string `json:"short_url"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
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
