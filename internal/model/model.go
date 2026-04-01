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
