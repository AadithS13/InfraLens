package scheduler

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"

	"github.com/infralens/infralens/internal/crawler"
	"github.com/infralens/infralens/internal/store"
)

type Scheduler struct {
	cron    *cron.Cron
	crawler *crawler.Crawler
	store   *store.Store
	startID int
	endID   int
}

func New(cr *crawler.Crawler, st *store.Store, startID, endID int) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		crawler: cr,
		store:   st,
		startID: startID,
		endID:   endID,
	}
}

// Start registers the cron job and begins the scheduler. schedule is a standard 5-field cron expression.
func (s *Scheduler) Start(schedule string) error {
	_, err := s.cron.AddFunc(schedule, s.runCrawl)
	if err != nil {
		return err
	}
	s.cron.Start()
	log.Printf("[SCHEDULER] started, schedule=%q (IDs %d→%d)", schedule, s.startID, s.endID)
	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) runCrawl() {
	ctx := context.Background()

	runID, err := s.store.InsertCrawlRun(ctx, s.startID, s.endID)
	if err != nil {
		log.Printf("[SCHEDULER] failed to create crawl_run record: %v", err)
		return
	}

	log.Printf("[SCHEDULER] crawl_run %d started (IDs %d→%d)", runID, s.startID, s.endID)

	processed, failed := s.crawler.Run(ctx, s.startID, s.endID)

	status := "completed"
	if failed > 0 {
		status = "completed_with_errors"
	}

	if err := s.store.UpdateCrawlRun(ctx, runID, processed, failed, status, nil); err != nil {
		log.Printf("[SCHEDULER] failed to update crawl_run %d: %v", runID, err)
	}

	log.Printf("[SCHEDULER] crawl_run %d done: status=%s processed=%d failed=%d", runID, status, processed, failed)
}
