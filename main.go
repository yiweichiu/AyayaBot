package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/yiweichiu/AyayaBot/config"
	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/logger"
	"github.com/yiweichiu/AyayaBot/scheduler"
)

var (
	globalBot       *discord.Bot
	globalScheduler *scheduler.Scheduler
)

func main() {
	// Check for single instance
	cleanup, ok := checkSingleInstance()
	if !ok {
		showAlert("AyayaBot", "程式已經在運行中！\n請檢查系統工作列。")
		return
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	systray.Run(onReady, onExit)
}

func onReady() {
	setupSystray()
	mQuit := systray.AddMenuItem("關閉 (Quit)", "停止機器人並結束程式")

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		systray.Quit()
		return
	}

	bot, err := setupDiscord(cfg)
	if err != nil {
		log.Printf("Failed to setup Discord: %v", err)
		systray.Quit()
		return
	}
	globalBot = bot

	s, err := setupScheduler(cfg, bot)
	if err != nil {
		log.Printf("Failed to setup Scheduler: %v", err)
		systray.Quit()
		return
	}
	globalScheduler = s

	setupSignals(mQuit)
	log.Println("AyayaBot is running.")
}

func setupSystray() {
	systray.SetIcon(iconData)
	systray.SetTitle("AyayaBot")
	systray.SetTooltip("AyayaBot")
}

func setupDiscord(cfg *config.Config) (*discord.Bot, error) {
	bot, err := discord.NewBot(cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	if err := bot.Start(); err != nil {
		return nil, err
	}
	return bot, nil
}

func setupScheduler(cfg *config.Config, bot *discord.Bot) (*scheduler.Scheduler, error) {
	s := scheduler.NewScheduler(cfg, bot)
	s.Start()

	// Register News jobs
	if cfg.News.Service && s.GetChannelID(cfg.News.Channel) != "" {
		for _, spec := range cfg.News.Schedule {
			if _, err := s.AddJob(spec, s.RunNewsTask); err != nil {
				log.Printf("Failed to add news job: %v", err)
			}
		}
	} else {
		log.Printf("News service is disabled or channel %s not found.", cfg.News.Channel)
	}

	// Register Redeem jobs
	if cfg.Redeem.Service && s.GetChannelID(cfg.Redeem.Channel) != "" {
		for _, spec := range cfg.Redeem.Schedule {
			if _, err := s.AddJob(spec, s.RunRedeemTask); err != nil {
				log.Printf("Failed to add redeem job: %v", err)
			}
		}
	} else {
		log.Printf("Redeem service is disabled or channel %s not found.", cfg.Redeem.Channel)
	}

	// Register Log rotation job
	if _, err := s.AddJob("1 0 * * *", func() {
		if err := logger.Rotate(); err != nil {
			log.Printf("Failed to rotate log: %v", err)
		}
	}); err != nil {
		log.Printf("Failed to add log rotation job: %v", err)
	}

	return s, nil
}

func setupSignals(mQuit *systray.MenuItem) {
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

		select {
		case <-sc:
			log.Println("Received termination signal.")
			systray.Quit()
		case <-mQuit.ClickedCh:
			log.Println("Quit clicked.")
			systray.Quit()
		}
	}()
}

func onExit() {
	log.Println("Shutting down bot...")
	if globalScheduler != nil {
		globalScheduler.Stop()
	}
	if globalBot != nil {
		globalBot.Stop()
	}
}
