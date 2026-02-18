package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	Cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Cron: cron.New(),
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

func (s *Scheduler) Start() {
	s.Cron.Start()
	log.Println("Scheduler started.")
}

func (s *Scheduler) Stop() {
	s.Cron.Stop()
	log.Println("Scheduler stopped.")
}
