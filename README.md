# InfraLens

A construction intelligence platform that crawls public RERA registries, normalizes project and promoter metadata, and maintains a structured Postgres database for analysis.

Inspired by platforms like Biltrax — built to demonstrate real-world data engineering: web crawling, API reverse engineering, change detection, and idempotent ingestion pipelines.

---

## What it does

InfraLens crawls the **MahaRERA** (Maharashtra Real Estate Regulatory Authority) public API and stores structured data about:

- Real estate projects (name, type, status, registration details, completion dates)
- Promoters / developers (name, PAN, GSTIN, type)
- Project and promoter addresses
- Professional contacts (architects, engineers, agents)
- Per-crawl snapshots with MD5 checksums for change detection

---

## Architecture

```
MahaRERA API
     │
     ▼
┌─────────────────────────────────┐
│  Crawler (Go)                   │
│  ├── Auth (Keycloak JWT)        │
│  ├── Worker Pool (5 goroutines) │
│  ├── Rate Limiting (300ms)      │
│  └── Upsert Pipeline           │
└──────────────┬──────────────────┘
               │
               ▼
        PostgreSQL
   ┌────────────────────┐
   │  promoters         │
   │  projects          │
   │  addresses         │
   │  contacts          │
   │  project_snapshots │
   └────────────────────┘
```

---

## DB Schema

| Table | Description |
|---|---|
| `projects` | Core project data, deduplicated by `maha_id` |
| `promoters` | Builder details, deduplicated by `user_profile_id` |
| `addresses` | Shared table for project + promoter addresses via `entity_type` / `entity_id` |
| `contacts` | Architects, engineers, agents linked to a project |
| `project_snapshots` | MD5 checksum of each crawl — foundation for change detection |

---

## Versioning Roadmap

| Version | Feature |
|---|---|
| **V1** ✅ | Data ingestion — crawl MahaRERA, normalize, store |
| **V2** | Idempotent crawling — checksum-based change detection, upserts |
| **V3** | Scheduled crawling — cron, retries, incremental updates |
| **V4** | Change history — `project_changes` table, before/after state |
| **V5** | Search — filter by city, budget, status |
| **V6** | Notifications — email/webhook on new registrations or status changes |

---

## Tech Stack

- **Go 1.22** — crawler, HTTP client, worker pool
- **PostgreSQL 16** — primary store
- **Docker** — local Postgres via docker-compose

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker

### 1. Start Postgres

```bash
docker compose up -d
```

This starts Postgres on port `5433` and runs `migrations/001_init.sql` automatically.

### 2. Run the crawler

```bash
DATABASE_URL="postgres://infralens:infralens@127.0.0.1:5433/infralens?sslmode=disable" \
  START_ID=1 \
  END_ID=1000 \
  go run ./cmd/crawler/
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://infralens:infralens@localhost:5432/infralens?sslmode=disable` | Postgres connection string |
| `START_ID` | `1` | First MahaRERA project ID to crawl |
| `END_ID` | `100000` | Last MahaRERA project ID to crawl |

---

## How the MahaRERA API works

MahaRERA's public site (`maharerait.maharashtra.gov.in`) is an Angular SPA backed by a REST API secured with Keycloak JWT. The site uses a hardcoded public-view user for unauthenticated browsing.

Project IDs are sequential integers. For each ID the crawler calls:

| Endpoint | Data |
|---|---|
| `getProjectGeneralDetailsByProjectId` | Name, RERA number, status, type, dates |
| `getProjectAndAssociatedPromoterDetails` | Promoter name (fallback) |
| `fetchPromoterGeneralDetails` | PAN, GSTIN, promoter type |
| `getProjectLandAddressDetails` | Project land address |
| `getPromoterAddressDetails` | Promoter office address |
| `getProjectProfessionalByType` | Architects, engineers |
| `getAgentByProjectId` | Registered agents |

The Bearer token is auto-refreshed every 98 minutes.

---

## Project Structure

```
InfraLens/
├── cmd/
│   └── crawler/
│       └── main.go              # Entry point
├── internal/
│   ├── client/
│   │   └── maharera.go          # HTTP client, auth, all API methods
│   ├── model/
│   │   └── project.go           # API response structs + DB models
│   ├── store/
│   │   └── postgres.go          # DB upserts
│   └── crawler/
│       └── crawler.go           # Worker pool, orchestration
├── migrations/
│   └── 001_init.sql             # Schema
└── docker-compose.yml
```
