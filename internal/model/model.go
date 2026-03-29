package model

type URL struct {
	Original string `json:"original"`
	Short    string `json:"short"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
}
