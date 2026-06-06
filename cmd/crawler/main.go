package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/infralens/infralens/internal/client"
	"github.com/infralens/infralens/internal/crawler"
	"github.com/infralens/infralens/internal/store"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://infralens:infralens@localhost:5432/infralens?sslmode=disable"
	}

	db, err := store.New(dsn)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	c := client.New()
	if err := c.Authenticate(ctx); err != nil {
		log.Fatalf("auth: %v", err)
	}
	log.Println("authenticated with MahaRERA")

	startID := intEnv("START_ID", 1)
	endID := intEnv("END_ID", 100000)

	log.Printf("crawling project IDs %d → %d", startID, endID)

	cr := crawler.New(c, db)
	cr.Run(ctx, startID, endID)
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
