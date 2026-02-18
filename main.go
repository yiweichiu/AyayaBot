package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil" // Use ioutil for ReadFile and WriteFile
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yiweichiu/AyayaBot/api"
	"github.com/yiweichiu/AyayaBot/config"
	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/scheduler"
)

const redeemFilePath = "redeem.json"

func loadRedeemCodesFromFile(filePath string) ([]string, error) {
	data, err := ioutil.ReadFile(filePath) // Use ioutil
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

func saveRedeemCodesToFile(filePath string, codes []string) error {
	data, err := json.MarshalIndent(codes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal redeem codes to JSON: %w", err)
	}

	err = ioutil.WriteFile(filePath, data, 0644) // Use ioutil
	if err != nil {
		return fmt.Errorf("failed to write redeem codes to file: %w", err)
	}
	return nil
}

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	discordBot, err := discord.NewBot(cfg.Discord.Token, cfg.Discord.ChannelID)
	if err != nil {
		log.Fatalf("Failed to create Discord bot: %v", err)
	}

	err = discordBot.Start()
	if err != nil {
		log.Fatalf("Failed to start Discord bot: %v", err)
	}
	defer discordBot.Stop()

	scheduler := scheduler.NewScheduler()
	scheduler.Start()
	defer scheduler.Stop()

	task := func() {
		log.Println("Fetching redeem codes...")
		fetchedCodesInfo, err := api.GetRedeemCodes(cfg.API.URL, cfg.API.APIKey) // Changed return type
		if err != nil {
			log.Printf("Failed to fetch redeem codes: %v", err)
			discordBot.SendMessage(fmt.Sprintf("Failed to fetch redeem codes: %v", err))
			return
		}

		// Load previously sent codes (just the code strings)
		previouslySentCodes, err := loadRedeemCodesFromFile(redeemFilePath)
		if err != nil {
			log.Printf("Error loading previously sent redeem codes: %v", err)
			// Continue without old codes if there's an error
		}

		// Convert previously sent codes to a map for efficient lookup
		sentCodesMap := make(map[string]bool)
		for _, code := range previouslySentCodes {
			sentCodesMap[code] = true
		}

		var newMessages []string
		var allCurrentCodesForSave []string // To save all current valid codes back to file

		for _, codeInfo := range fetchedCodesInfo { // Iterate over RedeemCodeInfo
			allCurrentCodesForSave = append(allCurrentCodesForSave, codeInfo.Code) // Save only the code string
			if !sentCodesMap[codeInfo.Code] {
				newMessages = append(newMessages, fmt.Sprintf("- %s: %s", codeInfo.Code, codeInfo.Reward))
			}
		}

		if len(newMessages) > 0 {
			message := fmt.Sprintf("## [New Brown Dust 2 Redeem Codes](https://thebd2pulse.com/)\n%s", strings.Join(newMessages, "\n"))
			err := discordBot.SendMessage(message)
			if err != nil {
				log.Printf("Failed to send new redeem codes to Discord: %v", err)
			} else {
				log.Printf("Sent %d new redeem codes to Discord.", len(newMessages))
				// After successfully sending new codes, update the redeem.json file
				if err := saveRedeemCodesToFile(redeemFilePath, allCurrentCodesForSave); err != nil {
					log.Printf("Error saving current redeem codes to file: %v", err)
				}
			}
		} else {
			log.Println("No new redeem codes available.")
			// Even if no new codes, ensure redeem.json is up-to-date with currently valid codes
			if err := saveRedeemCodesToFile(redeemFilePath, allCurrentCodesForSave); err != nil {
				log.Printf("Error saving current redeem codes to file: %v", err)
			}
		}
	}

	for _, spec := range cfg.Schedule {
		_, err := scheduler.AddJob(spec, task)
		if err != nil {
			log.Fatalf("Failed to add schedule job %s: %v", spec, err)
		}
	}

	log.Println("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Shutting down bot...")
}
