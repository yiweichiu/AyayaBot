package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/model" // Import model package
	"github.com/yiweichiu/AyayaBot/repository/bd2news"
)




// RunNewsTask executes the full flow of fetching, comparing, and notifying news.
func (s *Scheduler) RunNewsTask() {
	log.Println("Executing news fetch job...")

	newNews, err := bd2news.FetchNews(s.Config.News.API.URL)
	if err != nil {
		log.Printf("Error fetching news: %v", err)
		return
	}

	if len(newNews) == 0 {
		log.Println("No news items fetched from API.")
		return
	}

	oldNews, err := loadNewsFromFile()
	if err != nil {
		log.Printf("Error loading old news from file: %v", err)
		// Continue with empty oldNews if file load fails to allow saving new news
		oldNews = []model.NewsItem{}
	}

	err = compareAndNotify(s.DiscordBot, oldNews, newNews)
	if err != nil {
		log.Printf("Error comparing and notifying news: %v", err)
	}

	err = saveNewsToFile(newNews)
	if err != nil {
		log.Printf("Error saving new news to file: %v", err)
	}
}

// newsFileName is the name of the file to store news items.
const newsFileName = "news.json"

// loadNewsFromFile loads news items from news.json.
func loadNewsFromFile() ([]model.NewsItem, error) {
	data, err := os.ReadFile(newsFileName)
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

// saveNewsToFile saves news items to news.json.
func saveNewsToFile(newsItems []model.NewsItem) error {
	data, err := json.MarshalIndent(newsItems, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal news data for file: %w", err)
	}

	err = os.WriteFile(newsFileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write news data to file: %w", err)
	}
	return nil
}

// compareAndNotify compares old and new news items and sends notifications for new ones.
func compareAndNotify(bot *discord.Bot, oldNews, newNews []model.NewsItem) error {
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
			message := fmt.Sprintf("📢 **新公告**\n**標題:** %s\n**發佈時間:** %s\n**連結:** https://www.browndust2.com/zh-tw/news?page=0&type=all#%d",
				newAnnc.Subject, newAnnc.PublishedAt.Format("2006-01-02 15:04"), newAnnc.ID)
			err := bot.SendMessage(message)
			if err != nil {
				log.Printf("Error sending Discord message for new announcement %d: %v", newAnnc.ID, err)
			}
			time.Sleep(1 * time.Second) // Avoid hitting Discord rate limits
		}
	} else {
		log.Println("No new announcements found.")
	}

	return nil
}
