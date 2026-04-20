package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/model" // Import model package
	"github.com/yiweichiu/AyayaBot/repository/bd2news"
)

// RunNewsTask executes the full flow of fetching, comparing, and notifying news.
func (s *Scheduler) RunNewsTask() {
	if !s.Config.News.Service {
		return
	}
	log.Println("Executing news fetch job...")

	channelID := s.GetChannelID(s.Config.News.Channel)
	if channelID == "" {
		log.Printf("Error: Channel %s not found for news task", s.Config.News.Channel)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newNews, err := bd2news.FetchNews(ctx, s.Config.News.API.URL)
	if err != nil {
		log.Printf("Error fetching news: %v", err)
		return
	}

	if len(newNews) == 0 {
		log.Println("No news items fetched from API.")
		return
	}

	oldNews, err := loadNewsFromFile(s.Config.News.StoragePath)
	if err != nil {
		log.Printf("Error loading old news from file: %v", err)
		// Continue with empty oldNews if file load fails to allow saving new news
		oldNews = []model.NewsItem{}
	}

	err = compareAndNotify(s.DiscordBot, channelID, oldNews, newNews, s.Config.News.SendContent, s.Config.News.HideEmbed, s.Config.News.MentionRoleID)
	if err != nil {
		log.Printf("Error comparing and notifying news: %v", err)
	}

	err = saveNewsToFile(s.Config.News.StoragePath, newNews)
	if err != nil {
		log.Printf("Error saving new news to file: %v", err)
	}
}

// loadNewsFromFile loads news items from news.json.
func loadNewsFromFile(filePath string) ([]model.NewsItem, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.NewsItem{}, nil // File doesn't exist, return empty slice
		}
		return nil, fmt.Errorf("failed to read news file: %w", err)
	}

	var newsItems []model.NewsItem
	err = json.Unmarshal(data, &newsItems)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal news data from file: %w", err)
	}
	return newsItems, nil
}

// saveNewsToFile saves news items to news.json, only storing ID, Subject, and PublishedAt.
func saveNewsToFile(filePath string, newsItems []model.NewsItem) error {
	type minimalNewsItem struct {
		ID          int       `json:"id"`
		Subject     string    `json:"subject"`
		PublishedAt time.Time `json:"publishedAt"`
	}

	toSave := make([]minimalNewsItem, len(newsItems))
	for i, item := range newsItems {
		toSave[i] = minimalNewsItem{
			ID:          item.ID,
			Subject:     item.Subject,
			PublishedAt: item.PublishedAt,
		}
	}

	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal news data for file: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write news data to file: %w", err)
	}
	return nil
}

// compareAndNotify compares old and new news items and sends notifications for new ones.
func compareAndNotify(bot discord.Messenger, channelID string, oldNews, newNews []model.NewsItem, sendContent bool, hideEmbed bool, mentionRoleID string) error {
	oldNewsMap := make(map[int]struct{})
	for _, item := range oldNews {
		oldNewsMap[item.ID] = struct{}{}
	}

	var newAnnouncements []model.NewsItem
	for _, newItem := range newNews {
		if _, exists := oldNewsMap[newItem.ID]; !exists {
			newAnnouncements = append(newAnnouncements, newItem)
		}
	}

	if len(newAnnouncements) > 0 {
		log.Printf("Found %d new announcements.", len(newAnnouncements))
		// Iterate in reverse to send from oldest to newest
		for i := len(newAnnouncements) - 1; i >= 0; i-- {
			newAnnc := newAnnouncements[i]
			newsURL := fmt.Sprintf("https://www.browndust2.com/zh-tw/news/view?id=%d", newAnnc.ID)
			if hideEmbed {
				newsURL = "<" + newsURL + ">"
			}

			message := fmt.Sprintf("📢 **[新公告](%s) %s**\n**%s**\n",
				newsURL, newAnnc.PublishedAt.Local().Format("2006-01-02"), newAnnc.Subject)

			if tag := GetMentionTag(mentionRoleID); tag != "" {
				message = fmt.Sprintf("%s\n%s", tag, message)
			}

			if sendContent && newAnnc.Content != "" {
				content, err := htmltomarkdown.ConvertString(newAnnc.Content)
				if err != nil {
					log.Printf("Error converting HTML to Markdown for news %d: %v", newAnnc.ID, err)
					content = newAnnc.Content // Fallback to raw content if conversion fails
				}
				// Use TruncateString for safe UTF-8 and Markdown-aware truncation
				content = TruncateString(content, 1800)
				message += fmt.Sprintf("\n%s\n", content)
			}

			err := bot.SendMessage(channelID, message)
			if err != nil {
				log.Printf("Error sending Discord message for new announcement %d to channel %s: %v", newAnnc.ID, channelID, err)
			}
			time.Sleep(1 * time.Second) // Avoid hitting Discord rate limits
		}
	} else {
		log.Println("No new announcements found.")
	}

	return nil
}
