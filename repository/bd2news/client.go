package bd2news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"

	"github.com/yiweichiu/AyayaBot/model" // Import model package
)

// FetchNews fetches news from the given API URL and returns the latest items.
func FetchNews(ctx context.Context, apiURL string) ([]model.NewsItem, error) {
	// Parse the provided API URL and hardcode limit=20 and locale=zh-tw
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}
	q := u.Query()
	q.Set("limit", fmt.Sprintf("%d", model.DefaultNewsLimit))
	q.Set("locale", "zh-tw")
	u.RawQuery = q.Encode()
	finalURL := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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
	for _, item := range apiResponse.Items {
		newsItems = append(newsItems, model.NewsItem{
			ID:          item.ID,
			Subject:     item.Subject,
			PublishedAt: item.PublishedAt,
			Content:     item.ContentPreview,
		})
	}

	// Sort news items by PublishedAt in descending order (newest first).
	sort.Slice(newsItems, func(i, j int) bool {
		return newsItems[i].PublishedAt.After(newsItems[j].PublishedAt)
	})

	return newsItems, nil
}

// FetchNewsDetail fetches the full content of a news item by its ID.
func FetchNewsDetail(ctx context.Context, id string) (*model.NewsDetailResponse, error) {
	apiURL := fmt.Sprintf("https://webapi.browndust2.com/api/notices/%s?locale=zh-tw", id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news detail from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response body: %w", err)
	}

	var detailResponse model.NewsDetailResponse
	err = json.Unmarshal(body, &detailResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	return &detailResponse, nil
}
