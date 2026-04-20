package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/model"
	"github.com/yiweichiu/AyayaBot/repository/bd2redeem"
)

// RunRedeemTask executes the full flow of fetching and notifying redeem codes.
func (s *Scheduler) RunRedeemTask() {
	if !s.Config.Redeem.Service {
		return
	}
	log.Println("Fetching redeem codes...")

	channelID := s.GetChannelID(s.Config.Redeem.Channel)
	if channelID == "" {
		log.Printf("Error: Channel %s not found for redeem task", s.Config.Redeem.Channel)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fetchedCodesInfo, err := bd2redeem.GetRedeemCodes(ctx, s.Config.Redeem.API.URL, s.Config.Redeem.API.APIKey)
	if err != nil {
		log.Printf("Failed to fetch redeem codes: %v", err)
		sendErr := s.DiscordBot.SendMessage(channelID, fmt.Sprintf("Failed to fetch redeem codes: %v", err))
		if sendErr != nil {
			log.Printf("Error sending message about failed redeem code fetch: %v", sendErr)
		}
		return
	}

	previouslySentCodes, err := loadRedeemCodesFromFile(s.Config.Redeem.StoragePath)
	if err != nil {
		log.Printf("Error loading previously sent redeem codes: %v", err)
	}

	err = processRedeemTask(s.DiscordBot, channelID, fetchedCodesInfo, previouslySentCodes, s.Config.Redeem.HideEmbed, s.Config.Redeem.StoragePath, s.Config.Redeem.MentionRoleID)
	if err != nil {
		log.Printf("Error processing redeem task: %v", err)
	}
}

// processRedeemTask handles the comparison, notification, and saving logic for redeem codes.
func processRedeemTask(bot discord.Messenger, channelID string, fetchedCodesInfo []model.RedeemCodeInfo, previouslySentCodes []string, hideEmbed bool, storagePath string, mentionRoleID string) error {
	sentCodesMap := make(map[string]bool)
	for _, code := range previouslySentCodes {
		sentCodesMap[code] = true
	}

	var newMessages []string
	var allCurrentCodesForSave []string

	for _, codeInfo := range fetchedCodesInfo {
		allCurrentCodesForSave = append(allCurrentCodesForSave, codeInfo.Code)
		if !sentCodesMap[codeInfo.Code] {
			newMessages = append(newMessages, fmt.Sprintf("- %s: %s", codeInfo.Code, codeInfo.Reward))
		}
	}

	if len(newMessages) > 0 {
		redeemURL := "https://thebd2pulse.com/"
		if hideEmbed {
			redeemURL = "<" + redeemURL + ">"
		}
		message := fmt.Sprintf("📢 **[新兌換碼](%s)**\n%s", redeemURL, strings.Join(newMessages, "\n"))
		if tag := GetMentionTag(mentionRoleID); tag != "" {
			message = fmt.Sprintf("%s\n%s", tag, message)
		}
		if err := bot.SendMessage(channelID, message); err != nil {
			return fmt.Errorf("failed to send new redeem codes to Discord channel %s: %w", channelID, err)
		}
		log.Printf("Sent %d new redeem codes to Discord channel %s.", len(newMessages), channelID)
	} else {
		log.Println("No new redeem codes available.")
	}

	if err := saveRedeemCodesToFile(storagePath, allCurrentCodesForSave); err != nil {
		return fmt.Errorf("error saving current redeem codes to file: %w", err)
	}

	return nil
}

// loadRedeemCodesFromFile loads previously sent redeem code strings from a file.
func loadRedeemCodesFromFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // File doesn't exist, return empty slice
		}
		return nil, fmt.Errorf("failed to read redeem codes file: %w", err)
	}

	var codes []string
	err = json.Unmarshal(data, &codes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal redeem codes from file: %w", err)
	}
	return codes, nil
}

// saveRedeemCodesToFile saves redeem code strings to a file.
func saveRedeemCodesToFile(filePath string, codes []string) error {
	data, err := json.MarshalIndent(codes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal redeem codes to JSON: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write redeem codes to file: %w", err)
	}
	return nil
}
