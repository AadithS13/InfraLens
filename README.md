# InfraLens

A construction intelligence platform that crawls public RERA registries, normalizes project and promoter metadata, and maintains a structured Postgres database for analysis.

Inspired by platforms like Biltrax — built to demonstrate real-world data engineering: web crawling, API reverse engineering, idempotent ingestion, and field-level change detection.

---

## What it does

InfraLens crawls the **MahaRERA** (Maharashtra Real Estate Regulatory Authority) public API and stores structured data about:

- Real estate projects (name, type, status, registration details, completion dates)
- Promoters / developers (name, PAN, GSTIN, type)
- Project and promoter addresses
- Professional contacts (architects, engineers, agents)
- Per-crawl snapshots with MD5 checksums
- Field-level change history — what changed, from what value, to what value, when

---

## Architecture

```
MahaRERA API
     │
     ▼
┌──────────────────────────────────────┐
│  Crawler (Go)                        │
│  ├── Auth (Keycloak JWT, auto-refresh│
│  ├── Worker Pool (5 goroutines)      │
│  ├── Rate Limiting (300ms/worker)    │
│  ├── Idempotent Upserts             │
│  └── Change Detection Pipeline      │
└──────────────┬───────────────────────┘
               │
               ▼
         PostgreSQL
   ┌─────────────────────┐
   │  promoters          │
   │  projects           │
   │  addresses          │
   │  contacts           │
   │  project_snapshots  │
   │  project_changes    │
   └─────────────────────┘
```

---

## DB Schema

| Table | Description |
|---|---|
| `projects` | Core project data, deduplicated by `maha_id` (MahaRERA internal ID) |
| `promoters` | Builder details, deduplicated by `user_profile_id` |
| `addresses` | Shared table for project + promoter addresses via `entity_type` / `entity_id` |
| `contacts` | Architects, engineers, agents linked to a project |
| `project_snapshots` | Full raw JSON + MD5 checksum of every crawl per project |
| `project_changes` | Field-level change log — `field_name`, `old_value`, `new_value`, `detected_at` |

---

## How Change Detection Works

Every crawl per project goes through this flow:

```
Fetch from MahaRERA API
        │
        ▼
Generate MD5 checksum
        │
        ▼
Fetch latest snapshot from DB
        │
        ├── No snapshot?  → [NEW]  Insert everything
        │
        ├── Same checksum → [SAME] Skip, no changes
        │
        └── Different?   → [DIFF] Decode old snapshot JSON
                                  Compare tracked fields
                                  Write rows to project_changes
```

**Tracked fields** (one `project_changes` row per changed field):

| Field | Example change |
|---|---|
| `project_status` | `Under Approval` → `Ongoing` |
| `project_current_status` | `Submitted` → `Certificate Signed` |
| `proposed_completion_date` | `2024-12-31` → `2025-06-30` |
| `project_name` | Typo corrections by the builder |
| `total_units` | Units added after correction |
| `total_sold_units` | Booking progress |
| `rera_registration_no` | Rare, but tracked |

This makes queries like the following possible:

```sql
-- Projects whose status changed in the last 30 days
SELECT p.project_name, pc.old_value, pc.new_value, pc.detected_at
FROM project_changes pc
JOIN projects p ON p.id = pc.project_id
WHERE pc.field_name = 'project_status'
  AND pc.detected_at > NOW() - INTERVAL '30 days';

-- Which builders changed project status?
SELECT pr.name, COUNT(*) as changes
FROM project_changes pc
JOIN projects p  ON p.id = pc.project_id
JOIN promoters pr ON pr.id = p.promoter_id
WHERE pc.field_name = 'project_status'
GROUP BY pr.name
ORDER BY changes DESC;
```

---

## Versioning Roadmap

| Version | Feature | Status |
|---|---|---|
| **V1** | Data ingestion — crawl MahaRERA, normalize, store | ✅ Done |
| **V2** | Idempotent crawling — checksum comparison, field-level change detection | ✅ Done |
| **V3** | Scheduled crawling — nightly cron, `crawl_runs` tracking, retries | 🔜 Next |
| **V4** | Search API — filter by city, district, builder, status | 🔜 |
| **V5** | Notifications — email/webhook on status changes or new registrations | 🔜 |
| **V6** | Duplicate detection — trigram similarity + Levenshtein for project name dedup | 🔜 |

---

## Tech Stack

- **Go 1.22** — crawler, HTTP client, worker pool, change detection
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

Starts Postgres on port `5433` and auto-runs all migrations in `migrations/` on first boot.

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

MahaRERA's public site (`maharerait.maharashtra.gov.in`) is an Angular SPA backed by a REST API secured with Keycloak JWT. The site uses a hardcoded public-view account for unauthenticated browsing.

Project IDs are sequential integers. For each ID the crawler makes 7 parallel API calls:

| Endpoint | Data fetched |
|---|---|
| `getProjectGeneralDetailsByProjectId` | Name, RERA number, status, type, dates, `userProfileId` |
| `getProjectAndAssociatedPromoterDetails` | Promoter name (used as fallback when general details is empty) |
| `fetchPromoterGeneralDetails` | PAN, GSTIN, promoter type |
| `getProjectLandAddressDetails` | Project land address (plot, street, district, state) |
| `getPromoterAddressDetails` | Promoter office address |
| `getProjectProfessionalByType` | Architects, structural engineers |
| `getAgentByProjectId` | Registered real estate agents |

The Bearer token TTL is ~100 minutes and is auto-refreshed by the crawler.

---

## Migrations

Migration files follow Goose-style timestamp naming:

```
migrations/
├── 20260606135000_InfraLens_Create_Schema.sql      # Core tables
└── 20260606135001_InfraLens_Add_ProjectChanges.sql # Change detection table
```

---

## Project Structure

```
InfraLens/
├── cmd/
│   └── crawler/
│       └── main.go              # Entry point, reads env vars
├── internal/
│   ├── client/
│   │   └── maharera.go          # HTTP client, Keycloak auth, all API methods
│   ├── model/
│   │   └── project.go           # API response structs + DB models
│   ├── store/
│   │   └── postgres.go          # Upserts, snapshot reads, change inserts
│   └── crawler/
│       └── crawler.go           # Worker pool, orchestration, diff logic
├── migrations/                  # SQL migration files (timestamp-named)
└── docker-compose.yml
```
