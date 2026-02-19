package model

import "time"

// NewsAPIResponse represents the overall structure of the news API response.
type NewsAPIResponse struct {
	Data []struct {
		ID         int `json:"id"`
		Attributes struct {
			Subject     string    `json:"subject"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
			PublishedAt time.Time `json:"publishedAt"`
			Content     *string   `json:"content"` // Can be null
			Tag         string    `json:"tag"`
			Locale      string    `json:"locale"`
			NewContent  string    `json:"NewContent"`
		} `json:"attributes"`
	} `json:"data"`
	Meta struct {
		Pagination struct {
			Page    int `json:"page"`
			PageSize int `json:"pageSize"`
			PageCount int `json:"pageCount"`
			Total   int `json:"total"`
		} `json:"pagination"`
	} `json:"meta"`
}

// NewsItem represents a simplified structure of a single news item for storage and comparison.
type NewsItem struct {
	ID        int       `json:"id"`
	Subject   string    `json:"subject"`
	PublishedAt time.Time `json:"publishedAt"`
}
