package main

import (
	_ "embed"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
	"github.com/yiweichiu/AyayaBot/config"
	"github.com/yiweichiu/AyayaBot/discord"
	"github.com/yiweichiu/AyayaBot/logger"
	"github.com/yiweichiu/AyayaBot/scheduler"
)

//go:embed assets/icon.ico
var iconData []byte

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex = kernel32.NewProc("CreateMutexW")
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBox  = user32.NewProc("MessageBoxW")

	globalBot       *discord.Bot
	globalScheduler *scheduler.Scheduler
)

const (
	errorAlreadyExists syscall.Errno = 183
)

const (
	mbOk          uint32 = 0x00000000
	mbIconWarning uint32 = 0x00000030
)

func main() {
	// Check for single instance using Windows Named Mutex
	mutexName, _ := syscall.UTF16PtrFromString("Local\\AyayaBot-SingleInstance-Mutex")
	// ret is the handle to the mutex; we need to keep it alive during program execution
	ret, _, err := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))
	if err != nil && err.(syscall.Errno) == errorAlreadyExists {
		showAlert("AyayaBot", "程式已經在運行中！\n請檢查系統工作列。")
		return
	}
	// Prevent GC from collecting 'ret' handle if it were a managed object
	// In this case, 'ret' is a uintptr, so it stays alive as long as it's in scope.
	defer func(h uintptr) {
		// Optional: CloseHandle if needed, but the OS cleans up on exit
	}(ret)

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Use systray.Run to manage life cycle
	systray.Run(onReady, onExit)
}

func showAlert(title, message string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(message)
	_, _, _ = procMessageBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), uintptr(mbOk|mbIconWarning))
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("AyayaBot")
	systray.SetTooltip("AyayaBot")

	mQuit := systray.AddMenuItem("關閉 (Quit)", "停止機器人並結束程式")

	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		systray.Quit()
		return
	}

	// Create bot
	discordBot, err := discord.NewBot(cfg.Discord.Token)
	if err != nil {
		log.Printf("Failed to create Discord bot: %v", err)
		systray.Quit()
		return
	}

	// Start bot
	if err := discordBot.Start(); err != nil {
		log.Printf("Failed to start Discord bot: %v", err)
		systray.Quit()
		return
	}

	// Scheduler
	s := scheduler.NewScheduler(cfg, discordBot)
	s.Start()

	// Add jobs
	if cfg.News.Service && s.GetChannelID(cfg.News.Channel) != "" {
		for _, spec := range cfg.News.Schedule {
			if _, err := s.AddJob(spec, s.RunNewsTask); err != nil {
				log.Printf("Failed to add news job: %v", err)
			}
		}
	} else {
		log.Printf("News service is disabled or channel %s not found.", cfg.News.Channel)
	}

	if cfg.Redeem.Service && s.GetChannelID(cfg.Redeem.Channel) != "" {
		for _, spec := range cfg.Redeem.Schedule {
			if _, err := s.AddJob(spec, s.RunRedeemTask); err != nil {
				log.Printf("Failed to add redeem job: %v", err)
			}
		}
	} else {
		log.Printf("Redeem service is disabled or channel %s not found.", cfg.Redeem.Channel)
	}
	if _, err := s.AddJob("1 0 * * *", func() {
		if err := logger.Rotate(); err != nil {
			log.Printf("Failed to rotate log: %v", err)
		}
	}); err != nil {
		log.Printf("Failed to add log rotation job: %v", err)
	}

	log.Println("AyayaBot is running in background.")

	// Global assignments for cleanup
	globalBot = discordBot
	globalScheduler = s

	// Listen for signals and menu items in background
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
