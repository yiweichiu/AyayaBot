package bd2news

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/yiweichiu/AyayaBot/model" // Import model package
)

// FetchNews fetches news from the given API URL and returns the latest 10 items.
func FetchNews(apiURL string) ([]model.NewsItem, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response body: %w", err)
	}

	var apiResponse model.NewsAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	var newsItems []model.NewsItem
	for _, data := range apiResponse.Data {
		newsItems = append(newsItems, model.NewsItem{
			ID:          data.ID,
			Subject:     data.Attributes.Subject,
			PublishedAt: data.Attributes.PublishedAt,
			Content:     data.Attributes.NewContent,
		})
	}

	// Sort news items by PublishedAt in descending order (newest first).
	sort.Slice(newsItems, func(i, j int) bool {
		return newsItems[i].PublishedAt.After(newsItems[j].PublishedAt)
	})

	// Take the top 10 latest news items.
	if len(newsItems) > 10 {
		newsItems = newsItems[:10]
	}

	return newsItems, nil
}
