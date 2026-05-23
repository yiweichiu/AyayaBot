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
	now := time.Now()
	mockResponse := model.NewsAPIResponse{}
	
	// Item 1: Oldest
	item1 := struct {
		ID             string    `json:"id"`
		Subject        string    `json:"subject"`
		Category       string    `json:"category"`
		PublishedAt    time.Time `json:"publishedAt"`
		ContentPreview string    `json:"contentPreview"`
	}{
		ID:             "id1",
		Subject:        "Test News 1",
		PublishedAt:    now.Add(-1 * time.Hour),
		ContentPreview: "Preview 1",
	}
	mockResponse.Items = append(mockResponse.Items, item1)

	// Item 2: Newer
	item2 := struct {
		ID             string    `json:"id"`
		Subject        string    `json:"subject"`
		Category       string    `json:"category"`
		PublishedAt    time.Time `json:"publishedAt"`
		ContentPreview string    `json:"contentPreview"`
	}{
		ID:             "id2",
		Subject:        "Test News 2 (Newer)",
		PublishedAt:    now,
		ContentPreview: "Preview 2",
	}
	mockResponse.Items = append(mockResponse.Items, item2)

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
	if news[0].ID != "id2" {
		t.Errorf("Expected first item ID to be 'id2' (newer), got %s", news[0].ID)
	}
	if news[1].ID != "id1" {
		t.Errorf("Expected second item ID to be 'id1', got %s", news[1].ID)
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
