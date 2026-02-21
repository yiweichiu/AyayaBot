package scheduler

import (
	"strings"
	"testing"
	"time"

	"github.com/yiweichiu/AyayaBot/model"
)

// MockMessenger mocks the discord.Messenger interface for testing.
type MockMessenger struct {
	Messages []string
}

func (m *MockMessenger) SendMessage(message string) error {
	m.Messages = append(m.Messages, message)
	return nil
}

func TestCompareAndNotify(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNews := []model.NewsItem{
		{ID: 1, Subject: "Old News 1", PublishedAt: now.Add(-2 * time.Hour)},
		{ID: 2, Subject: "Old News 2", PublishedAt: now.Add(-1 * time.Hour)},
	}

	newNews := []model.NewsItem{
		{ID: 3, Subject: "New News 3 (Newer)", PublishedAt: now},
		{ID: 2, Subject: "Old News 2", PublishedAt: now.Add(-1 * time.Hour)},
		{ID: 1, Subject: "Old News 1", PublishedAt: now.Add(-2 * time.Hour)},
	}

	err := compareAndNotify(mockBot, oldNews, newNews)
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	// Should only notify for ID 3
	if len(mockBot.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(mockBot.Messages))
	}

	if mockBot.Messages[0] == "" {
		t.Error("Expected message content, got empty string")
	}

	expectedDate := now.Local().Format("2006-01-02")
	if !strings.Contains(mockBot.Messages[0], expectedDate) {
		t.Errorf("Expected message to contain date %s, got: %s", expectedDate, mockBot.Messages[0])
	}
}

func TestCompareAndNotify_Order(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNews := []model.NewsItem{} // No old news
	newNews := []model.NewsItem{
		{ID: 2, Subject: "News 2 (Newer)", PublishedAt: now},
		{ID: 1, Subject: "News 1 (Older)", PublishedAt: now.Add(-1 * time.Hour)},
	}

	err := compareAndNotify(mockBot, oldNews, newNews)
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	if len(mockBot.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(mockBot.Messages))
	}

	// Should be sent in reverse order (oldest to newest)
	// News items in newNews are usually sorted newest first from API.
	// So newNews[0] is ID: 2, newNews[1] is ID: 1.
	// compareAndNotify should iterate from newest to oldest in reverse:
	// newNews[1] (ID 1), then newNews[0] (ID 2).

	if !strings.Contains(mockBot.Messages[0], "[新公告]") {
		t.Errorf("Expected first message to contain markdown link '[新公告]', got: %s", mockBot.Messages[0])
	}
	if !strings.Contains(mockBot.Messages[1], "[新公告]") {
		t.Errorf("Expected second message to contain markdown link '[新公告]', got: %s", mockBot.Messages[1])
	}
}
