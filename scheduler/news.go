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

	oldNewsIDs, err := loadNewsFromFile(s.Config.News.StoragePath)
	if err != nil {
		log.Printf("Error loading old news from file: %v", err)
		// Continue with empty oldNews if file load fails to allow saving new news
		oldNewsIDs = []string{}
	}

	err = compareAndNotify(s.DiscordBot, channelID, oldNewsIDs, newNews, s.Config.News.SendContent, s.Config.News.HideEmbed, s.Config.News.MentionRoleID)
	if err != nil {
		log.Printf("Error comparing and notifying news: %v", err)
	}

	err = saveNewsToFile(s.Config.News.StoragePath, newNews)
	if err != nil {
		log.Printf("Error saving new news to file: %v", err)
	}
}

// loadNewsFromFile loads news IDs from news.json.
func loadNewsFromFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // File doesn't exist, return empty slice
		}
		return nil, fmt.Errorf("failed to read news file: %w", err)
	}

	var newsIDs []string
	err = json.Unmarshal(data, &newsIDs)
	if err != nil {
		// Try to handle old format (slice of objects) for backward compatibility
		type oldNewsItem struct {
			ID string `json:"id"`
		}
		var oldItems []oldNewsItem
		if err2 := json.Unmarshal(data, &oldItems); err2 == nil {
			ids := make([]string, len(oldItems))
			for i, item := range oldItems {
				ids[i] = item.ID
			}
			return ids, nil
		}
		return nil, fmt.Errorf("failed to unmarshal news data from file: %w", err)
	}
	return newsIDs, nil
}

// saveNewsToFile saves news IDs to news.json.
func saveNewsToFile(filePath string, newsItems []model.NewsItem) error {
	// Limit the number of news items to save to model.DefaultNewsLimit
	if len(newsItems) > model.DefaultNewsLimit {
		newsItems = newsItems[:model.DefaultNewsLimit]
	}

	toSave := make([]string, len(newsItems))
	for i, item := range newsItems {
		toSave[i] = item.ID
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

// compareAndNotify compares old news IDs and new news items and sends notifications for new ones.
func compareAndNotify(bot discord.Messenger, channelID string, oldNewsIDs []string, newNews []model.NewsItem, sendContent bool, hideEmbed bool, mentionRoleID string) error {
	oldNewsMap := make(map[string]struct{})
	for _, id := range oldNewsIDs {
		oldNewsMap[id] = struct{}{}
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
			newsURL := fmt.Sprintf("https://www.browndust2.com/zh-tw/news/view?id=%s", newAnnc.ID)
			if hideEmbed {
				newsURL = "<" + newsURL + ">"
			}

			message := fmt.Sprintf("📢 **[新公告](%s) %s**\n**%s**\n",
				newsURL, newAnnc.PublishedAt.Local().Format("2006-01-02"), newAnnc.Subject)

			if tag := GetMentionTag(mentionRoleID); tag != "" {
				message = fmt.Sprintf("%s\n%s", tag, message)
			}

			if sendContent {
				// Fetch full content from detail API
				detail, err := bd2news.FetchNewsDetail(context.Background(), newAnnc.ID)
				if err != nil {
					log.Printf("Error fetching detail for news %s: %v", newAnnc.ID, err)
					// Fallback to preview content
					if newAnnc.Content != "" {
						content, _ := htmltomarkdown.ConvertString(newAnnc.Content)
						message += fmt.Sprintf("\n%s\n", TruncateString(content, 1800))
					}
				} else if detail.ContentHtml != "" {
					content, err := htmltomarkdown.ConvertString(detail.ContentHtml)
					if err != nil {
						log.Printf("Error converting HTML to Markdown for news %s: %v", newAnnc.ID, err)
						content = detail.ContentHtml
					}
					message += fmt.Sprintf("\n%s\n", TruncateString(content, 1800))
				}
			}

			err := bot.SendMessage(channelID, message)
			if err != nil {
				log.Printf("Error sending Discord message for new announcement %s to channel %s: %v", newAnnc.ID, channelID, err)
			}
			time.Sleep(1 * time.Second) // Avoid hitting Discord rate limits
		}
	} else {
		log.Println("No new announcements found.")
	}

	return nil
}
