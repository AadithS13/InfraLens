package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"strings"

	"github.com/infralens/infralens/internal/client"
	"github.com/infralens/infralens/internal/core"
	"github.com/infralens/infralens/internal/crawler"
	"github.com/infralens/infralens/internal/notifier"
	"github.com/infralens/infralens/internal/repo"
	"github.com/infralens/infralens/internal/scheduler"
	"github.com/infralens/infralens/internal/server"
	"github.com/infralens/infralens/internal/server/handler"
	"github.com/infralens/infralens/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Dev default — matches docker-compose.yml. Set DATABASE_URL in production.
		dsn = "postgres://infralens:infralens@127.0.0.1:5433/infralens?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	crawlSchedule := os.Getenv("CRAWL_SCHEDULE")
	if crawlSchedule == "" {
		crawlSchedule = "0 2 * * *" // daily at 2am
	}
	startID := intEnv("START_ID", 1)
	endID := intEnv("END_ID", 100000)

	// Read pool (API queries)
	readPool, err := repo.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("read db: %v", err)
	}
	defer readPool.Close()

	// Write pool (crawler / scheduler writes)
	writeStore, err := store.New(dsn)
	if err != nil {
		log.Fatalf("write db: %v", err)
	}
	defer writeStore.Close()

	// Core / handler layers
	projectRepo := repo.NewProjectRepo(readPool)
	projectSvc := core.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectSvc)
	crawlHandler := handler.NewCrawlHandler(projectSvc)
	analyticsHandler := handler.NewAnalyticsHandler(projectSvc)
	searchHandler := handler.NewSearchHandler(projectSvc)
	nlSearchHandler := handler.NewNLSearchHandler(projectSvc)

	// Notifier — always log; add email/webhook if env vars are present
	adapters := []notifier.Adapter{notifier.NewLogAdapter()}

	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		emailTo := strings.Split(os.Getenv("NOTIFY_EMAIL_TO"), ",")
		if len(emailTo) > 0 && emailTo[0] != "" {
			smtpPort := os.Getenv("SMTP_PORT")
			if smtpPort == "" {
				smtpPort = "587"
			}
			adapters = append(adapters, notifier.NewEmailAdapter(notifier.EmailConfig{
				Host: smtpHost,
				Port: smtpPort,
				User: os.Getenv("SMTP_USER"),
				Pass: os.Getenv("SMTP_PASS"),
				From: os.Getenv("NOTIFY_EMAIL_FROM"),
				To:   emailTo,
			}))
			log.Printf("[NOTIFIER] email adapter enabled → %s", os.Getenv("NOTIFY_EMAIL_TO"))
		}
	}

	if webhookURL := os.Getenv("NOTIFY_WEBHOOK_URL"); webhookURL != "" {
		adapters = append(adapters, notifier.NewWebhookAdapter(webhookURL))
		log.Printf("[NOTIFIER] webhook adapter enabled → %s", webhookURL)
	}

	notify := notifier.New(writeStore, adapters...)

	// Crawler + scheduler
	mahaClient := client.New()
	if err := mahaClient.Authenticate(ctx); err != nil {
		log.Fatalf("maharera auth: %v", err)
	}
	cr := crawler.New(mahaClient, writeStore)
	sched := scheduler.New(cr, writeStore, notify, startID, endID)
	if err := sched.Start(crawlSchedule); err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	defer sched.Stop()

	// HTTP server
	srv := server.New(projectHandler, crawlHandler, analyticsHandler, searchHandler, nlSearchHandler, port)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
