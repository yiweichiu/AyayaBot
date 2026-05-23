package scheduler

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yiweichiu/AyayaBot/model"
)

// MockMessenger mocks the discord.Messenger interface for testing.
type MockMessenger struct {
	Messages []string
}

func (m *MockMessenger) SendMessage(channelID, message string) error {
	m.Messages = append(m.Messages, message)
	return nil
}

func TestCompareAndNotify(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()
	channelID := "test-channel"

	oldNewsIDs := []string{"id1", "id2"}

	newNews := []model.NewsItem{
		{ID: "id3", Subject: "New News 3 (Newer)", PublishedAt: now},
		{ID: "id2", Subject: "Old News 2", PublishedAt: now.Add(-1 * time.Hour)},
		{ID: "id1", Subject: "Old News 1", PublishedAt: now.Add(-2 * time.Hour)},
	}

	err := compareAndNotify(mockBot, channelID, oldNewsIDs, newNews, false, false, "")
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
	expectedMessage := fmt.Sprintf("📢 **[新公告](https://www.browndust2.com/zh-tw/news/view?id=id3) %s**\n**New News 3 (Newer)**\n", expectedDate)
	if mockBot.Messages[0] != expectedMessage {
		t.Errorf("Expected message:\n%q\nGot:\n%q", expectedMessage, mockBot.Messages[0])
	}
}

func TestCompareAndNotify_HideEmbed(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNewsIDs := []string{}
	newNews := []model.NewsItem{
		{ID: "id1", Subject: "News 1", PublishedAt: now},
	}

	err := compareAndNotify(mockBot, "test-channel", oldNewsIDs, newNews, false, true, "")
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	expectedDate := now.Local().Format("2006-01-02")
	expectedMessage := fmt.Sprintf("📢 **[新公告](<https://www.browndust2.com/zh-tw/news/view?id=id1>) %s**\n**News 1**\n", expectedDate)
	if mockBot.Messages[0] != expectedMessage {
		t.Errorf("Expected message with < >:\n%q\nGot:\n%q", expectedMessage, mockBot.Messages[0])
	}
}

func TestCompareAndNotify_Mention(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNewsIDs := []string{}
	newNews := []model.NewsItem{
		{ID: "id1", Subject: "News 1", PublishedAt: now},
	}

	err := compareAndNotify(mockBot, "test-channel", oldNewsIDs, newNews, false, false, "123456789")
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	if !strings.HasPrefix(mockBot.Messages[0], "<@&123456789>\n") {
		t.Errorf("Expected message to start with role mention and newline, got: %s", mockBot.Messages[0])
	}
}

func TestCompareAndNotify_WithContent(t *testing.T) {
	mockBot := &MockMessenger{}
	now := time.Now()

	oldNewsIDs := []string{}
	newNews := []model.NewsItem{
		{ID: "id1", Subject: "News 1", PublishedAt: now, Content: "This is the content"},
	}

	err := compareAndNotify(mockBot, "test-channel", oldNewsIDs, newNews, true, true, "")
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

	oldNewsIDs := []string{} // No old news
	newNews := []model.NewsItem{
		{ID: "id2", Subject: "News 2 (Newer)", PublishedAt: now},
		{ID: "id1", Subject: "News 1 (Older)", PublishedAt: now.Add(-1 * time.Hour)},
	}

	err := compareAndNotify(mockBot, "test-channel", oldNewsIDs, newNews, false, true, "")
	if err != nil {
		t.Fatalf("compareAndNotify failed: %v", err)
	}

	if len(mockBot.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(mockBot.Messages))
	}

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

func TestNewsFileStorage(t *testing.T) {
	tempFile := "test_news.json"
	defer os.Remove(tempFile)

	now := time.Now().Truncate(time.Second) // Truncate to avoid precision issues in JSON
	newsItems := []model.NewsItem{
		{ID: "id1", Subject: "Subject 1", PublishedAt: now, Content: "Content 1"},
		{ID: "id2", Subject: "Subject 2", PublishedAt: now.Add(time.Hour), Content: "Content 2"},
	}

	// Save news items
	err := saveNewsToFile(tempFile, newsItems)
	if err != nil {
		t.Fatalf("saveNewsToFile failed: %v", err)
	}

	// Load news IDs
	loadedIDs, err := loadNewsFromFile(tempFile)
	if err != nil {
		t.Fatalf("loadNewsFromFile failed: %v", err)
	}

	if len(loadedIDs) != 2 {
		t.Fatalf("Expected 2 news IDs, got %d", len(loadedIDs))
	}

	for i, id := range loadedIDs {
		if id != newsItems[i].ID {
			t.Errorf("Item %d: Expected ID %s, got %s", i, newsItems[i].ID, id)
		}
	}
}

func TestNewsFileStorage_BackwardCompatibility(t *testing.T) {
	tempFile := "test_news_compat.json"
	defer os.Remove(tempFile)

	// Old format data
	oldData := `[
		{"id": "id1", "subject": "Sub 1", "publishedAt": "2024-01-01T00:00:00Z"},
		{"id": "id2", "subject": "Sub 2", "publishedAt": "2024-01-01T01:00:00Z"}
	]`
	err := os.WriteFile(tempFile, []byte(oldData), 0644)
	if err != nil {
		t.Fatalf("failed to write old format file: %v", err)
	}

	// Load news IDs
	loadedIDs, err := loadNewsFromFile(tempFile)
	if err != nil {
		t.Fatalf("loadNewsFromFile failed: %v", err)
	}

	if len(loadedIDs) != 2 {
		t.Fatalf("Expected 2 news IDs, got %d", len(loadedIDs))
	}

	if loadedIDs[0] != "id1" || loadedIDs[1] != "id2" {
		t.Errorf("Incorrect IDs loaded: %v", loadedIDs)
	}
}
