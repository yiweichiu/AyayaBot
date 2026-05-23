package model

import "time"

// DefaultNewsLimit defines the default number of news items to fetch and store.
const DefaultNewsLimit = 20

// NewsAPIResponse represents the overall structure of the new news API response.
type NewsAPIResponse struct {
	Items []struct {
		ID             string    `json:"id"`
		Subject        string    `json:"subject"`
		Category       string    `json:"category"`
		PublishedAt    time.Time `json:"publishedAt"`
		ContentPreview string    `json:"contentPreview"`
	} `json:"items"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
	TotalPages int `json:"totalPages"`
}

// NewsDetailResponse represents the structure of the news detail API response.
type NewsDetailResponse struct {
	ID          string    `json:"id"`
	Subject     string    `json:"subject"`
	Category    string    `json:"category"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
	ContentHtml string    `json:"contentHtml"`
}

// NewsItem represents a simplified structure of a single news item for storage and comparison.
type NewsItem struct {
	ID          string    `json:"id"`
	Subject     string    `json:"subject"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}
