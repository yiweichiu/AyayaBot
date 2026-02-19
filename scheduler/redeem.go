package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yiweichiu/AyayaBot/repository/bd2redeem"
)

const redeemFilePath = "redeem.json"

// RunRedeemTask executes the full flow of fetching and notifying redeem codes.
func (s *Scheduler) RunRedeemTask() {
	log.Println("Fetching redeem codes...")
	fetchedCodesInfo, err := bd2redeem.GetRedeemCodes(s.Config.Redeem.API.URL, s.Config.Redeem.API.APIKey)
	if err != nil {
		log.Printf("Failed to fetch redeem codes: %v", err)
		sendErr := s.DiscordBot.SendMessage(fmt.Sprintf("Failed to fetch redeem codes: %v", err))
		if sendErr != nil {
			log.Printf("Error sending message about failed redeem code fetch: %v", sendErr)
		}
		return
	}

	previouslySentCodes, err := loadRedeemCodesFromFile(redeemFilePath)
	if err != nil {
		log.Printf("Error loading previously sent redeem codes: %v", err)
	}

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
		message := fmt.Sprintf("📢 **[新兌換碼](https://thebd2pulse.com/)**\n%s", strings.Join(newMessages, "\n"))
		if err := s.DiscordBot.SendMessage(message); err != nil {
			log.Printf("Failed to send new redeem codes to Discord: %v", err)
		} else {
			log.Printf("Sent %d new redeem codes to Discord.", len(newMessages))
			if err := saveRedeemCodesToFile(redeemFilePath, allCurrentCodesForSave); err != nil {
				log.Printf("Error saving current redeem codes to file: %v", err)
			}
		}
	} else {
		log.Println("No new redeem codes available.")
		if err := saveRedeemCodesToFile(redeemFilePath, allCurrentCodesForSave); err != nil {
			log.Printf("Error saving current redeem codes to file: %v", err)
		}
	}
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
