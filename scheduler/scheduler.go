package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
	"github.com/yiweichiu/AyayaBot/config"  // Import config package
	"github.com/yiweichiu/AyayaBot/discord" // Import discord package
)

// Scheduler manages scheduled tasks for various functionalities.
type Scheduler struct {
	Cron       *cron.Cron
	Config     *config.Config
	DiscordBot *discord.Bot
}

// NewScheduler creates and initializes a new Scheduler instance.
func NewScheduler(cfg *config.Config, bot *discord.Bot) *Scheduler {
	return &Scheduler{
		Cron:       cron.New(),
		Config:     cfg,
		DiscordBot: bot,
	}
}

func (s *Scheduler) AddJob(spec string, cmd func()) (cron.EntryID, error) {
	id, err := s.Cron.AddFunc(spec, cmd)
	if err != nil {
		return 0, err
	}
	log.Printf("Added scheduled job with spec: %s", spec)
	return id, nil
}

func (s *Scheduler) RemoveJob(id cron.EntryID) {
	s.Cron.Remove(id)
	log.Printf("Removed scheduled job with ID: %v", id)
}

func (s *Scheduler) Start() {
	s.Cron.Start()
	log.Println("Scheduler started.")
}

func (s *Scheduler) Stop() {
	s.Cron.Stop()
	log.Println("Scheduler stopped.")
}

// GetChannelID returns the Discord channel ID for the given channel name from config.
func (s *Scheduler) GetChannelID(name string) string {
	if s.Config.ChannelMap == nil {
		return ""
	}
	return s.Config.ChannelMap[name]
}
