package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/robfig/cron/v3"
	"github.com/yiweichiu/AyayaBot/config"
	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/logger"
	"github.com/yiweichiu/AyayaBot/scheduler"
	"github.com/yiweichiu/AyayaBot/updater"
)

var (
	globalBot       *discord.Bot
	globalScheduler *scheduler.Scheduler
	newsJobIDs      []cron.EntryID
	redeemJobIDs    []cron.EntryID
	shouldRestart   bool
)

func main() {
	// Check for single instance
	cleanup, ok := checkSingleInstance()
	if !ok {
		showAlert("AyayaBot", "程式已經在運行中！\n請檢查系統工作列。")
		return
	}

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	systray.Run(onReady, onExit)

	if cleanup != nil {
		cleanup()
	}

	if shouldRestart {
		restartApp()
	}
}

func onReady() {
	setupSystray()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		systray.Quit()
		return
	}

	systray.AddMenuItem(fmt.Sprintf("版本: %s", updater.CurrentVersion), "當前程式版本").Disable()
	systray.AddSeparator()

	mNewsParent := systray.AddMenuItem("新聞", "新聞通知設定")
	mNewsService := mNewsParent.AddSubMenuItemCheckbox("啟用", "開關新聞通知功能", cfg.News.Service)
	mNewsContent := mNewsParent.AddSubMenuItemCheckbox("發送內文", "是否在通知中包含新聞內容", cfg.News.SendContent)
	mNewsEmbed := mNewsParent.AddSubMenuItemCheckbox("隱藏預覽", "是否隱藏 Discord 連結預覽", cfg.News.HideEmbed)
	mNewsMention := mNewsParent.AddSubMenuItem("標註設定", "設定標註對象")
	mNewsMentionNone := mNewsMention.AddSubMenuItemCheckbox("無", "", cfg.News.MentionRoleID == "")
	mNewsMentionHere := mNewsMention.AddSubMenuItemCheckbox("@here", "", cfg.News.MentionRoleID == "here")
	mNewsMentionID := mNewsMention.AddSubMenuItemCheckbox("自訂 ID", "", cfg.News.MentionRoleID != "" && cfg.News.MentionRoleID != "here")

	mNewsFreq := mNewsParent.AddSubMenuItem("檢查頻率", "設定新聞檢查頻率")
	mNewsFreq1 := mNewsFreq.AddSubMenuItemCheckbox("每小時", "", len(cfg.News.Schedule) > 0 && cfg.News.Schedule[0] == "0 * * * *")
	mNewsFreq2 := mNewsFreq.AddSubMenuItemCheckbox("每 2 小時", "", len(cfg.News.Schedule) > 0 && cfg.News.Schedule[0] == "0 */2 * * *")
	mNewsFreq4 := mNewsFreq.AddSubMenuItemCheckbox("每 4 小時", "", len(cfg.News.Schedule) > 0 && cfg.News.Schedule[0] == "0 */4 * * *")
	mNewsFreq8 := mNewsFreq.AddSubMenuItemCheckbox("每 8 小時", "", len(cfg.News.Schedule) > 0 && cfg.News.Schedule[0] == "0 */8 * * *")
	mNewsFreq12 := mNewsFreq.AddSubMenuItemCheckbox("每 12 小時", "", len(cfg.News.Schedule) > 0 && cfg.News.Schedule[0] == "0 */12 * * *")
	mNewsFreqDay := mNewsFreq.AddSubMenuItemCheckbox("每天", "", len(cfg.News.Schedule) > 0 && cfg.News.Schedule[0] == "0 9 * * *")

	mRedeemParent := systray.AddMenuItem("兌換碼", "兌換碼通知設定")
	mRedeemService := mRedeemParent.AddSubMenuItemCheckbox("啟用", "開關兌換碼通知功能", cfg.Redeem.Service)
	mRedeemEmbed := mRedeemParent.AddSubMenuItemCheckbox("隱藏預覽", "是否隱藏 Discord 連結預覽", cfg.Redeem.HideEmbed)
	mRedeemMention := mRedeemParent.AddSubMenuItem("標註設定", "設定標註對象")
	mRedeemMentionNone := mRedeemMention.AddSubMenuItemCheckbox("無", "", cfg.Redeem.MentionRoleID == "")
	mRedeemMentionHere := mRedeemMention.AddSubMenuItemCheckbox("@here", "", cfg.Redeem.MentionRoleID == "here")
	mRedeemMentionID := mRedeemMention.AddSubMenuItemCheckbox("自訂 ID", "", cfg.Redeem.MentionRoleID != "" && cfg.Redeem.MentionRoleID != "here")

	mRedeemFreq := mRedeemParent.AddSubMenuItem("檢查頻率", "設定兌換碼檢查頻率")
	mRedeemFreq1 := mRedeemFreq.AddSubMenuItemCheckbox("每小時", "", len(cfg.Redeem.Schedule) > 0 && cfg.Redeem.Schedule[0] == "0 * * * *")
	mRedeemFreq2 := mRedeemFreq.AddSubMenuItemCheckbox("每 2 小時", "", len(cfg.Redeem.Schedule) > 0 && cfg.Redeem.Schedule[0] == "0 */2 * * *")
	mRedeemFreq4 := mRedeemFreq.AddSubMenuItemCheckbox("每 4 小時", "", len(cfg.Redeem.Schedule) > 0 && cfg.Redeem.Schedule[0] == "0 */4 * * *")
	mRedeemFreq8 := mRedeemFreq.AddSubMenuItemCheckbox("每 8 小時", "", len(cfg.Redeem.Schedule) > 0 && cfg.Redeem.Schedule[0] == "0 */8 * * *")
	mRedeemFreq12 := mRedeemFreq.AddSubMenuItemCheckbox("每 12 小時", "", len(cfg.Redeem.Schedule) > 0 && cfg.Redeem.Schedule[0] == "0 */12 * * *")
	mRedeemFreqDay := mRedeemFreq.AddSubMenuItemCheckbox("每天", "", len(cfg.Redeem.Schedule) > 0 && cfg.Redeem.Schedule[0] == "0 9 * * *")

	systray.AddSeparator()
	mUpdate := systray.AddMenuItem("檢查更新", "檢查是否有新版本")
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

	setupSignals(mQuit, mUpdate, mNewsService, mNewsContent, mNewsEmbed, mRedeemService, mRedeemEmbed,
		mNewsMentionNone, mNewsMentionHere, mNewsMentionID,
		mRedeemMentionNone, mRedeemMentionHere, mRedeemMentionID,
		mNewsFreq1, mNewsFreq2, mNewsFreq4, mNewsFreq8, mNewsFreq12, mNewsFreqDay,
		mRedeemFreq1, mRedeemFreq2, mRedeemFreq4, mRedeemFreq8, mRedeemFreq12, mRedeemFreqDay,
		cfg)
	log.Println("AyayaBot is running.")
}

func runUpdateCheck() {
	info, err := updater.CheckUpdate()
	if err != nil {
		log.Printf("Failed to check update: %v", err)
		showAlert("檢查更新失敗", fmt.Sprintf("無法獲取更新資訊：\n%v", err))
		return
	}

	if info == nil {
		showAlert("檢查更新", "目前已經是最新版本！")
		return
	}

	msg := fmt.Sprintf("發現新版本 %s，是否要開始更新？\n更新完成後程式將自動重新啟動。", info.Version)
	if showConfirmDialog("發現新版本", msg) {
		if err := updater.DoUpdate(info.DownloadURL); err != nil {
			log.Printf("Failed to apply update: %v", err)
			showAlert("更新失敗", fmt.Sprintf("更新過程中發生錯誤：\n%v", err))
		} else {
			showAlert("更新成功", "更新已完成，程式即將自動重新啟動。")
			shouldRestart = true
			systray.Quit()
		}
	}
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
			id, err := s.AddJob(spec, s.RunNewsTask)
			if err != nil {
				log.Printf("Failed to add news job: %v", err)
			} else {
				newsJobIDs = append(newsJobIDs, id)
			}
		}
	} else {
		log.Printf("News channel %s not found. News task will not run.", cfg.News.Channel)
	}

	// Register Redeem jobs
	if s.GetChannelID(cfg.Redeem.Channel) != "" {
		for _, spec := range cfg.Redeem.Schedule {
			id, err := s.AddJob(spec, s.RunRedeemTask)
			if err != nil {
				log.Printf("Failed to add redeem job: %v", err)
			} else {
				redeemJobIDs = append(redeemJobIDs, id)
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

func updateFrequencyChecks(m1, m2, m4, m8, m12, mDay *systray.MenuItem, spec string) {
	m1.Uncheck()
	m2.Uncheck()
	m4.Uncheck()
	m8.Uncheck()
	m12.Uncheck()
	mDay.Uncheck()

	switch spec {
	case "0 * * * *":
		m1.Check()
	case "0 */2 * * *":
		m2.Check()
	case "0 */4 * * *":
		m4.Check()
	case "0 */8 * * *":
		m8.Check()
	case "0 */12 * * *":
		m12.Check()
	case "0 9 * * *":
		mDay.Check()
	}
}

func reloadNewsJobs(cfg *config.Config) {
	if globalScheduler == nil {
		return
	}
	for _, id := range newsJobIDs {
		globalScheduler.RemoveJob(id)
	}
	newsJobIDs = nil
	for _, spec := range cfg.News.Schedule {
		id, err := globalScheduler.AddJob(spec, globalScheduler.RunNewsTask)
		if err == nil {
			newsJobIDs = append(newsJobIDs, id)
		}
	}
}

func reloadRedeemJobs(cfg *config.Config) {
	if globalScheduler == nil {
		return
	}
	for _, id := range redeemJobIDs {
		globalScheduler.RemoveJob(id)
	}
	redeemJobIDs = nil
	for _, spec := range cfg.Redeem.Schedule {
		id, err := globalScheduler.AddJob(spec, globalScheduler.RunRedeemTask)
		if err == nil {
			redeemJobIDs = append(redeemJobIDs, id)
		}
	}
}

func setupSignals(mQuit, mUpdate, mNews, mNewsContent, mNewsEmbed, mRedeem, mRedeemEmbed,
	mNewsMNone, mNewsMHere, mNewsMID,
	mRedeemMNone, mRedeemMHere, mRedeemMID,
	mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay,
	mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay *systray.MenuItem,
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
			case <-mUpdate.ClickedCh:
				runUpdateCheck()
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

			// News Frequency Handlers
			case <-mNewsF1.ClickedCh:
				cfg.News.Schedule = []string{"0 * * * *"}
				updateFrequencyChecks(mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay, cfg.News.Schedule[0])
				reloadNewsJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsF2.ClickedCh:
				cfg.News.Schedule = []string{"0 */2 * * *"}
				updateFrequencyChecks(mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay, cfg.News.Schedule[0])
				reloadNewsJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsF4.ClickedCh:
				cfg.News.Schedule = []string{"0 */4 * * *"}
				updateFrequencyChecks(mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay, cfg.News.Schedule[0])
				reloadNewsJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsF8.ClickedCh:
				cfg.News.Schedule = []string{"0 */8 * * *"}
				updateFrequencyChecks(mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay, cfg.News.Schedule[0])
				reloadNewsJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsF12.ClickedCh:
				cfg.News.Schedule = []string{"0 */12 * * *"}
				updateFrequencyChecks(mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay, cfg.News.Schedule[0])
				reloadNewsJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mNewsFDay.ClickedCh:
				cfg.News.Schedule = []string{"0 9 * * *"}
				updateFrequencyChecks(mNewsF1, mNewsF2, mNewsF4, mNewsF8, mNewsF12, mNewsFDay, cfg.News.Schedule[0])
				reloadNewsJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)

			// Redeem Frequency Handlers
			case <-mRedeemF1.ClickedCh:
				cfg.Redeem.Schedule = []string{"0 * * * *"}
				updateFrequencyChecks(mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay, cfg.Redeem.Schedule[0])
				reloadRedeemJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemF2.ClickedCh:
				cfg.Redeem.Schedule = []string{"0 */2 * * *"}
				updateFrequencyChecks(mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay, cfg.Redeem.Schedule[0])
				reloadRedeemJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemF4.ClickedCh:
				cfg.Redeem.Schedule = []string{"0 */4 * * *"}
				updateFrequencyChecks(mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay, cfg.Redeem.Schedule[0])
				reloadRedeemJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemF8.ClickedCh:
				cfg.Redeem.Schedule = []string{"0 */8 * * *"}
				updateFrequencyChecks(mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay, cfg.Redeem.Schedule[0])
				reloadRedeemJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemF12.ClickedCh:
				cfg.Redeem.Schedule = []string{"0 */12 * * *"}
				updateFrequencyChecks(mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay, cfg.Redeem.Schedule[0])
				reloadRedeemJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
			case <-mRedeemFDay.ClickedCh:
				cfg.Redeem.Schedule = []string{"0 9 * * *"}
				updateFrequencyChecks(mRedeemF1, mRedeemF2, mRedeemF4, mRedeemF8, mRedeemF12, mRedeemFDay, cfg.Redeem.Schedule[0])
				reloadRedeemJobs(cfg)
				_ = config.SaveConfig("config.yaml", cfg)
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
