package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yiweichiu/AyayaBot/config"
	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/logger"
	"github.com/yiweichiu/AyayaBot/scheduler"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

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

	s := scheduler.NewScheduler(cfg, discordBot) // Pass cfg and discordBot
	s.Start()
	defer s.Stop()

	// Add News fetching job
	for _, spec := range cfg.News.Schedule {
		if _, err := s.AddJob(spec, s.RunNewsTask); err != nil {
			log.Fatalf("Failed to add news job: %v", err)
		}
	}

	// Add Redeem Codes fetching job
	for _, spec := range cfg.Redeem.Schedule {
		if _, err := s.AddJob(spec, s.RunRedeemTask); err != nil {
			log.Fatalf("Failed to add redeem job: %v", err)
		}
	}

	// Add Log rotation job at 00:01 daily
	if _, err := s.AddJob("1 0 * * *", func() {
		if err := logger.Rotate(); err != nil {
			log.Printf("Failed to rotate log: %v", err)
		}
	}); err != nil {
		log.Fatalf("Failed to add log rotation job: %v", err)
	}

	log.Println("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Shutting down bot...")
}
