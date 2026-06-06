DB_DSN=postgres://postgres:postgres@localhost:5432/infralens?sslmode=disable

.PHONY: up down migrate-up migrate-down migrate-status migrate-create crawl crawl-dry build

up:
	docker compose up -d

down:
	docker compose down

migrate-up:
	goose -dir migrations postgres "$(DB_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(DB_DSN)" down

migrate-status:
	goose -dir migrations postgres "$(DB_DSN)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql

build:
	go build -o bin/crawler ./cmd/crawler

crawl:
	go run ./cmd/crawler --state=karnataka

crawl-dry:
	go run ./cmd/crawler --state=karnataka --dry-run
