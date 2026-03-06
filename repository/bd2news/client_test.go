package bd2news

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yiweichiu/AyayaBot/model"
)

func TestFetchNews(t *testing.T) {
	// ... (rest of the setup)

	now := time.Now()
	mockResponse := model.NewsAPIResponse{}
	
	// Item 1: Oldest
	item1 := struct {
		ID         int `json:"id"`
		Attributes struct {
			Subject     string    `json:"subject"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
			PublishedAt time.Time `json:"publishedAt"`
			Content     *string   `json:"content"`
			Tag         string    `json:"tag"`
			Locale      string    `json:"locale"`
			NewContent  string    `json:"NewContent"`
		} `json:"attributes"`
	}{
		ID: 1,
	}
	item1.Attributes.Subject = "Test News 1"
	item1.Attributes.PublishedAt = now.Add(-1 * time.Hour)
	mockResponse.Data = append(mockResponse.Data, item1)

	// Item 2: Newer
	item2 := struct {
		ID         int `json:"id"`
		Attributes struct {
			Subject     string    `json:"subject"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
			PublishedAt time.Time `json:"publishedAt"`
			Content     *string   `json:"content"`
			Tag         string    `json:"tag"`
			Locale      string    `json:"locale"`
			NewContent  string    `json:"NewContent"`
		} `json:"attributes"`
	}{
		ID: 2,
	}
	item2.Attributes.Subject = "Test News 2 (Newer)"
	item2.Attributes.PublishedAt = now
	mockResponse.Data = append(mockResponse.Data, item2)

	// 建立 Mock HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// 執行被測試函式
	news, err := FetchNews(context.Background(), server.URL)

	// 驗證結果
	if err != nil {
		t.Fatalf("FetchNews failed: %v", err)
	}
	if len(news) != 2 {
		t.Errorf("Expected 2 news items, got %d", len(news))
	}

	// 驗證排序 (Newer first)
	if news[0].ID != 2 {
		t.Errorf("Expected first item ID to be 2 (newer), got %d", news[0].ID)
	}
	if news[1].ID != 1 {
		t.Errorf("Expected second item ID to be 1, got %d", news[1].ID)
	}
}

func TestFetchNews_APIError(t *testing.T) {
	// 建立回傳 500 錯誤的 Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// 執行並驗證錯誤
	news, err := FetchNews(context.Background(), server.URL)
	if err == nil {
		t.Fatal("Expected error for 500 status, got nil")
	}
	if news != nil {
		t.Errorf("Expected nil news on error, got %v", news)
	}
	if !strings.Contains(err.Error(), "API returned non-OK status") {
		t.Errorf("Expected error message to contain 'API returned non-OK status', got: %v", err)
	}
}
