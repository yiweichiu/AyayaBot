package scheduler

import (
	"fmt"
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

	err := compareAndNotify(mockBot, oldNews, newNews, false, false)
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
	expectedMessage := fmt.Sprintf("📢 **[新公告](https://www.browndust2.com/zh-tw/news/view?id=3) %s**\n**New News 3 (Newer)**\n", expectedDate)
	if mockBot.Messages[0] != expectedMessage {
		t.Errorf("Expected message:\n%q\nGot:\n%q", expectedMessage, mockBot.Messages[0])
	}
}

func TestCompareAndNotify_HideEmbed(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNews := []model.NewsItem{}
	newNews := []model.NewsItem{
		{ID: 1, Subject: "News 1", PublishedAt: now},
	}

	err := compareAndNotify(mockBot, oldNews, newNews, false, true)
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	expectedDate := now.Local().Format("2006-01-02")
	expectedMessage := fmt.Sprintf("📢 **[新公告](<https://www.browndust2.com/zh-tw/news/view?id=1>) %s**\n**News 1**\n", expectedDate)
	if mockBot.Messages[0] != expectedMessage {
		t.Errorf("Expected message with < >:\n%q\nGot:\n%q", expectedMessage, mockBot.Messages[0])
	}
}

func TestCompareAndNotify_WithContent(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNews := []model.NewsItem{}
	newNews := []model.NewsItem{
		{ID: 1, Subject: "News 1", PublishedAt: now, Content: "This is the content"},
	}

	err := compareAndNotify(mockBot, oldNews, newNews, true, true)
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	if len(mockBot.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(mockBot.Messages))
	}

	if !strings.Contains(mockBot.Messages[0], "This is the content") {
		t.Errorf("Expected message to contain content, got: %s", mockBot.Messages[0])
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

	err := compareAndNotify(mockBot, oldNews, newNews, false, true)
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

	if !strings.Contains(mockBot.Messages[0], "News 1 (Older)") {
		t.Errorf("Expected first message to be 'News 1 (Older)', got: %s", mockBot.Messages[0])
	}
	if !strings.Contains(mockBot.Messages[1], "News 2 (Newer)") {
		t.Errorf("Expected second message to be 'News 2 (Newer)', got: %s", mockBot.Messages[1])
	}

	for i, msg := range mockBot.Messages {
		if !strings.HasPrefix(msg, "📢 **[新公告]") {
			t.Errorf("Message %d expected to start with '📢 **[新公告]', got: %s", i, msg)
		}
	}
}
