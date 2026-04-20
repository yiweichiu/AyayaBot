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

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		systray.Quit()
		return
	}

	mNewsParent := systray.AddMenuItem("新聞", "新聞通知設定")
	mNewsService := mNewsParent.AddSubMenuItemCheckbox("啟用", "開關新聞通知功能", cfg.News.Service)
	mNewsContent := mNewsParent.AddSubMenuItemCheckbox("發送內文", "是否在通知中包含新聞內容", cfg.News.SendContent)
	mNewsEmbed := mNewsParent.AddSubMenuItemCheckbox("隱藏預覽", "是否隱藏 Discord 連結預覽", cfg.News.HideEmbed)
	mNewsMention := mNewsParent.AddSubMenuItem("標註設定", "設定標註對象")
	mNewsMentionNone := mNewsMention.AddSubMenuItemCheckbox("無", "", cfg.News.MentionRoleID == "")
	mNewsMentionHere := mNewsMention.AddSubMenuItemCheckbox("@here", "", cfg.News.MentionRoleID == "here")
	mNewsMentionID := mNewsMention.AddSubMenuItemCheckbox("自訂 ID", "", cfg.News.MentionRoleID != "" && cfg.News.MentionRoleID != "here")

	mRedeemParent := systray.AddMenuItem("兌換碼", "兌換碼通知設定")
	mRedeemService := mRedeemParent.AddSubMenuItemCheckbox("啟用", "開關兌換碼通知功能", cfg.Redeem.Service)
	mRedeemEmbed := mRedeemParent.AddSubMenuItemCheckbox("隱藏預覽", "是否隱藏 Discord 連結預覽", cfg.Redeem.HideEmbed)
	mRedeemMention := mRedeemParent.AddSubMenuItem("標註設定", "設定標註對象")
	mRedeemMentionNone := mRedeemMention.AddSubMenuItemCheckbox("無", "", cfg.Redeem.MentionRoleID == "")
	mRedeemMentionHere := mRedeemMention.AddSubMenuItemCheckbox("@here", "", cfg.Redeem.MentionRoleID == "here")
	mRedeemMentionID := mRedeemMention.AddSubMenuItemCheckbox("自訂 ID", "", cfg.Redeem.MentionRoleID != "" && cfg.Redeem.MentionRoleID != "here")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("關閉", "停止機器人並結束程式")

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

	setupSignals(mQuit, mNewsService, mNewsContent, mNewsEmbed, mRedeemService, mRedeemEmbed,
		mNewsMentionNone, mNewsMentionHere, mNewsMentionID,
		mRedeemMentionNone, mRedeemMentionHere, mRedeemMentionID,
		cfg)
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
	if s.GetChannelID(cfg.News.Channel) != "" {
		for _, spec := range cfg.News.Schedule {
			if _, err := s.AddJob(spec, s.RunNewsTask); err != nil {
				log.Printf("Failed to add news job: %v", err)
			}
		}
	} else {
		log.Printf("News channel %s not found. News task will not run.", cfg.News.Channel)
	}

	// Register Redeem jobs
	if s.GetChannelID(cfg.Redeem.Channel) != "" {
		for _, spec := range cfg.Redeem.Schedule {
			if _, err := s.AddJob(spec, s.RunRedeemTask); err != nil {
				log.Printf("Failed to add redeem job: %v", err)
			}
		}
	} else {
		log.Printf("Redeem channel %s not found. Redeem task will not run.", cfg.Redeem.Channel)
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

func updateMentionChecks(none, here, custom *systray.MenuItem, roleID string) {
	none.Uncheck()
	here.Uncheck()
	custom.Uncheck()

	switch roleID {
	case "":
		none.Check()
	case "here":
		here.Check()
	default:
		custom.Check()
	}
}

func setupSignals(mQuit, mNews, mNewsContent, mNewsEmbed, mRedeem, mRedeemEmbed,
	mNewsMNone, mNewsMHere, mNewsMID,
	mRedeemMNone, mRedeemMHere, mRedeemMID *systray.MenuItem,
	cfg *config.Config) {
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

		for {
			select {
			case <-sc:
				log.Println("Received termination signal.")
				systray.Quit()
				return
			case <-mQuit.ClickedCh:
				log.Println("Quit clicked.")
				systray.Quit()
				return
			case <-mNews.ClickedCh:
				cfg.News.Service = !cfg.News.Service
				if cfg.News.Service {
					mNews.Check()
					log.Println("News service enabled.")
				} else {
					mNews.Uncheck()
					log.Println("News service disabled.")
				}
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsContent.ClickedCh:
				cfg.News.SendContent = !cfg.News.SendContent
				if cfg.News.SendContent {
					mNewsContent.Check()
					log.Println("News send content enabled.")
				} else {
					mNewsContent.Uncheck()
					log.Println("News send content disabled.")
				}
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsEmbed.ClickedCh:
				cfg.News.HideEmbed = !cfg.News.HideEmbed
				if cfg.News.HideEmbed {
					mNewsEmbed.Check()
					log.Println("News hide embed enabled.")
				} else {
					mNewsEmbed.Uncheck()
					log.Println("News hide embed disabled.")
				}
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeem.ClickedCh:
				cfg.Redeem.Service = !cfg.Redeem.Service
				if cfg.Redeem.Service {
					mRedeem.Check()
					log.Println("Redeem service enabled.")
				} else {
					mRedeem.Uncheck()
					log.Println("Redeem service disabled.")
				}
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemEmbed.ClickedCh:
				cfg.Redeem.HideEmbed = !cfg.Redeem.HideEmbed
				if cfg.Redeem.HideEmbed {
					mRedeemEmbed.Check()
					log.Println("Redeem hide embed enabled.")
				} else {
					mRedeemEmbed.Uncheck()
					log.Println("Redeem hide embed disabled.")
				}
				_ = config.SaveConfig("config.yaml", cfg)

			// News Mention Handlers
			case <-mNewsMNone.ClickedCh:
				cfg.News.MentionRoleID = ""
				updateMentionChecks(mNewsMNone, mNewsMHere, mNewsMID, cfg.News.MentionRoleID)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsMHere.ClickedCh:
				cfg.News.MentionRoleID = "here"
				updateMentionChecks(mNewsMNone, mNewsMHere, mNewsMID, cfg.News.MentionRoleID)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsMID.ClickedCh:
				if id, ok := showInputDialog("自訂標註 ID", "請輸入 Discord 身分組 ID:", cfg.News.MentionRoleID); ok {
					cfg.News.MentionRoleID = id
					updateMentionChecks(mNewsMNone, mNewsMHere, mNewsMID, cfg.News.MentionRoleID)
					_ = config.SaveConfig("config.yaml", cfg)
				}

			// Redeem Mention Handlers
			case <-mRedeemMNone.ClickedCh:
				cfg.Redeem.MentionRoleID = ""
				updateMentionChecks(mRedeemMNone, mRedeemMHere, mRedeemMID, cfg.Redeem.MentionRoleID)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemMHere.ClickedCh:
				cfg.Redeem.MentionRoleID = "here"
				updateMentionChecks(mRedeemMNone, mRedeemMHere, mRedeemMID, cfg.Redeem.MentionRoleID)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemMID.ClickedCh:
				if id, ok := showInputDialog("自訂標註 ID", "請輸入 Discord 身分組 ID:", cfg.Redeem.MentionRoleID); ok {
					cfg.Redeem.MentionRoleID = id
					updateMentionChecks(mRedeemMNone, mRedeemMHere, mRedeemMID, cfg.Redeem.MentionRoleID)
					_ = config.SaveConfig("config.yaml", cfg)
				}
			}
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
