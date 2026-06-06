package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/infralens/infralens/internal/core"
	"github.com/infralens/infralens/internal/repo"
	"github.com/infralens/infralens/internal/server"
	"github.com/infralens/infralens/internal/server/handler"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://infralens:infralens@127.0.0.1:5433/infralens?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Repo layer
	db, err := repo.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	// Core layer
	projectRepo := repo.NewProjectRepo(db)
	projectSvc := core.NewProjectService(projectRepo)

	// Server layer
	projectHandler := handler.NewProjectHandler(projectSvc)
	srv := server.New(projectHandler, port)

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}
