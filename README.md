# InfraLens

A construction intelligence platform that crawls public RERA registries, normalizes project and promoter metadata, maintains field-level change history, and exposes a REST API for search and analysis.

Inspired by platforms like Biltrax — built to demonstrate real-world data engineering: web crawling, API reverse engineering, idempotent ingestion, change detection, and layered API design.

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

See [System Architecture](#system-architecture) diagram below for the full flow.

---

## System Architecture

```mermaid
flowchart TD
    A[MahaRERA Public API] --> B[Auth Layer\nKeycloak JWT]
    B --> C[Crawler]
    C --> D[Worker Pool\n5 Goroutines]
    D --> E[Rate Limiter\n300ms / worker]
    E --> F[Fetch Project Data\n7 parallel API calls]
    F --> G[Generate MD5 Checksum]
    G --> H{Snapshot Exists?}
    H -- Same Checksum --> I[Skip]
    H -- No Snapshot --> J[Insert Project]
    H -- Different Checksum --> K[Detect Changes]
    K --> L[Write project_changes]
    J --> M[(PostgreSQL)]
    L --> M
    M --> N[REST API]
    N --> O[Search Projects]
    N --> P[Project Details]
    N --> Q[Change History]
```

---

## Database Schema

InfraLens stores normalized project, promoter, contact and address data while maintaining historical snapshots and field-level change tracking for incremental synchronization.

```mermaid
erDiagram
    PROMOTERS {
        bigint id PK
        bigint user_profile_id UK
        string name
        string pan
        string gstin
        string promoter_type
        timestamp created_at
    }
    PROJECTS {
        bigint id PK
        bigint maha_id UK
        bigint promoter_id FK
        string project_name
        string project_type
        string project_status
        string project_current_status
        string rera_registration_no
        date rera_registration_date
        date proposed_completion_date
        int total_units
        int total_sold_units
        timestamp created_at
        timestamp updated_at
    }
    CONTACTS {
        bigint id PK
        bigint project_id FK
        string role
        string name
        string phone
        string email
        timestamp created_at
    }
    PROJECT_SNAPSHOTS {
        bigint id PK
        bigint project_id FK
        string checksum
        string raw_json
        timestamp fetched_at
    }
    PROJECT_CHANGES {
        bigint id PK
        bigint project_id FK
        string field_name
        string old_value
        string new_value
        timestamp detected_at
    }
    ADDRESSES {
        bigint id PK
        string entity_type
        bigint entity_id
        string line1
        string city
        string district
        string state
        string pincode
        timestamp created_at
    }

    PROMOTERS ||--o{ PROJECTS : "owns"
    PROJECTS ||--o{ CONTACTS : "has"
    PROJECTS ||--o{ PROJECT_SNAPSHOTS : "stores"
    PROJECTS ||--o{ PROJECT_CHANGES : "generates"
    PROJECTS ||--o| ADDRESSES : "project address"
    PROMOTERS ||--o| ADDRESSES : "promoter address"
```

| Table | Description |
|---|---|
| `projects` | Core project data, deduplicated by `maha_id` (MahaRERA internal ID) |
| `promoters` | Builder details, deduplicated by `user_profile_id` |
| `addresses` | Shared table for project + promoter addresses via `entity_type` / `entity_id` |
| `contacts` | Architects, engineers, agents linked to a project |
| `project_snapshots` | Full raw JSON + MD5 checksum of every crawl per project |
| `project_changes` | Field-level change log — `field_name`, `old_value`, `new_value`, `detected_at` |

---

## API in Action

### Search Projects

Filter by any combination of city, district, state, promoter, status, or type. All string filters use case-insensitive partial matching.

```bash
curl "localhost:8080/api/v1/projects?district=Nagpur&limit=2"
```

```json
{
  "data": [
    {
      "id": 26,
      "maha_id": 19,
      "rera_registration_no": "P50500001314",
      "project_name": "Diamond One",
      "project_type": "Others",
      "project_status": "New",
      "project_current_status": "Certificate Signed",
      "rera_registration_date": "2017-07-27T00:00:00Z",
      "proposed_completion_date": "2019-12-31T00:00:00Z",
      "total_units": 0,
      "total_sold_units": 0,
      "promoter_name": "Diamond Estate Builders & Developers",
      "city": "",
      "district": "Nagpur",
      "state": "MAHARASHTRA",
      "pincode": "440010"
    },
    {
      "id": 25,
      "maha_id": 16,
      "rera_registration_no": "P50500000348",
      "project_name": "ROYAL RESIDENCY",
      "project_type": "Residential / Group Housing",
      "project_status": "New",
      "project_current_status": "Certificate Signed",
      "rera_registration_date": "2017-07-15T00:00:00Z",
      "proposed_completion_date": "2020-11-30T00:00:00Z",
      "total_units": 0,
      "total_sold_units": 0,
      "promoter_name": "DESIGN DEVELOPERS PRIVATE LIMITED",
      "city": "",
      "district": "Nagpur",
      "state": "MAHARASHTRA",
      "pincode": "440015"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 2,
    "total": 8
  }
}
```

---

### Change History — The Hero Feature

Every time the crawler detects a field value has changed since the last crawl, it writes a row to `project_changes`. This is what turns a crawler into an intelligence product.

```bash
curl "localhost:8080/api/v1/projects/2/changes"
```

```json
{
  "data": [
    {
      "field_name": "project_status",
      "old_value": "Under Approval",
      "new_value": "Ongoing",
      "detected_at": "2026-06-06T09:33:55Z"
    },
    {
      "field_name": "proposed_completion_date",
      "old_value": "2024-12-31",
      "new_value": "2025-06-30",
      "detected_at": "2026-05-01T02:10:00Z"
    }
  ]
}
```

Now you can answer questions like:
- *"Which projects changed status in the last 30 days?"*
- *"Which builders keep extending completion dates?"*
- *"How many projects moved from Under Approval to Ongoing this month?"*

This is the data Biltrax sells.

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
        ├── Same checksum → [SAME] Skip, nothing to do
        │
        └── Different?   → [DIFF] Decode old snapshot JSON
                                  Compare 7 tracked fields
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
JOIN projects p   ON p.id  = pc.project_id
JOIN promoters pr ON pr.id = p.promoter_id
WHERE pc.field_name = 'project_status'
GROUP BY pr.name
ORDER BY changes DESC;
```

---

## Search API

### Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/projects` | Search projects with filters and pagination |
| `GET` | `/api/v1/projects/{id}` | Full project detail — promoter, addresses, contacts |
| `GET` | `/api/v1/projects/{id}/changes` | Field-level change history for a project |
| `GET` | `/health` | Health check |

### Query Parameters — `GET /api/v1/projects`

| Param | Type | Description |
|---|---|---|
| `city` | string | Partial match on project city |
| `district` | string | Partial match on district |
| `state` | string | Partial match on state |
| `promoter` | string | Partial match on promoter/builder name |
| `status` | string | Exact match on project status |
| `type` | string | Partial match on project type |
| `page` | int | Page number (default: 1) |
| `limit` | int | Results per page (default: 20, max: 100) |

### Example Requests

```bash
# All projects in Nagpur
GET /api/v1/projects?district=Nagpur

# Ongoing residential projects, page 2
GET /api/v1/projects?status=Ongoing&type=Residential&page=2&limit=20

# Projects by a specific builder
GET /api/v1/projects?promoter=Prestige

# Full project detail
GET /api/v1/projects/42

# What changed on project 42?
GET /api/v1/projects/42/changes
```

### Example Response — `GET /api/v1/projects`

```json
{
  "data": [
    {
      "id": 1,
      "maha_id": 5,
      "rera_registration_no": "P51700002065",
      "project_name": "MERIDIAN MYSTIC",
      "project_type": "Residential / Group Housing",
      "project_status": "Ongoing",
      "project_current_status": "Certificate Signed",
      "rera_registration_date": "2017-05-20T00:00:00Z",
      "proposed_completion_date": "2021-12-31T00:00:00Z",
      "total_units": 0,
      "total_sold_units": 0,
      "promoter_name": "MANOJ AWASTHI",
      "city": "Nerul Navi Mumbai",
      "district": "Thane",
      "state": "MAHARASHTRA",
      "pincode": "400706"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1547
  }
}
```

### Layered Architecture

The API follows a strict three-layer design:

```
Request
   │
   ▼
server/handler    — parse params, write JSON, nothing else
   │
   ▼
core              — business logic, pagination defaults, DTOs
   │              — defines ProjectRepo interface (repo is swappable)
   ▼
repo              — SQL only, returns typed structs
   │
   ▼
PostgreSQL
```

| Layer | Package | Responsibility |
|---|---|---|
| HTTP | `internal/server` | Routing (chi), middleware, request/response |
| Business | `internal/core` | Validation, defaults, orchestration, DTOs |
| Data | `internal/repo` | SQL queries, DB connection |

---

## Versioning Roadmap

| Version | Feature | Status |
|---|---|---|
| **V1** | Data ingestion — crawl MahaRERA, normalize, store | ✅ Done |
| **V2** | Idempotent crawling — checksum comparison, field-level change detection | ✅ Done |
| **V3** | Search API — layered REST API with filters, pagination, change history | ✅ Done |
| **V4** | Scheduled crawling — nightly cron, `crawl_runs` tracking, retries | 🔜 Next |
| **V5** | Notifications — email/webhook on status changes or new registrations | 🔜 |
| **V6** | Duplicate detection — trigram similarity + Levenshtein for project name dedup | 🔜 |

---

## Tech Stack

- **Go 1.23** — crawler, HTTP client, worker pool, REST API
- **chi** — lightweight HTTP router
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

### 3. Run the API server

```bash
DATABASE_URL="postgres://infralens:infralens@127.0.0.1:5433/infralens?sslmode=disable" \
  PORT=8080 \
  go run ./cmd/api/
```

### Environment Variables

**Crawler (`cmd/crawler`)**

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://infralens:infralens@localhost:5433/infralens?sslmode=disable` | Postgres DSN |
| `START_ID` | `1` | First MahaRERA project ID to crawl |
| `END_ID` | `100000` | Last MahaRERA project ID to crawl |

**API Server (`cmd/api`)**

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://infralens:infralens@localhost:5433/infralens?sslmode=disable` | Postgres DSN |
| `PORT` | `8080` | HTTP port |

---

## How the MahaRERA API works

MahaRERA's public site (`maharerait.maharashtra.gov.in`) is an Angular SPA backed by a REST API secured with Keycloak JWT. The site uses a hardcoded public-view account for unauthenticated browsing.

Project IDs are sequential integers. For each ID the crawler makes 7 parallel API calls:

| Endpoint | Data fetched |
|---|---|
| `getProjectGeneralDetailsByProjectId` | Name, RERA number, status, type, dates, `userProfileId` |
| `getProjectAndAssociatedPromoterDetails` | Promoter name (fallback when general details is empty) |
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
│   ├── api/
│   │   └── main.go              # API server entry point
│   └── crawler/
│       └── main.go              # Crawler entry point
├── internal/
│   ├── client/
│   │   └── maharera.go          # HTTP client, Keycloak auth, all API methods
│   ├── model/
│   │   └── project.go           # API response structs + DB models
│   ├── store/
│   │   └── postgres.go          # Crawler write layer (upserts, snapshots, changes)
│   ├── crawler/
│   │   └── crawler.go           # Worker pool, orchestration, diff logic
│   ├── repo/
│   │   └── project.go           # API read layer (search, detail, changes)
│   ├── core/
│   │   ├── project.go           # ProjectService — business logic
│   │   └── types.go             # DTOs, SearchFilter, ListResponse
│   └── server/
│       ├── server.go            # chi router, middleware, graceful shutdown
│       └── handler/
│           └── project.go       # HTTP handlers
├── migrations/
└── docker-compose.yml
```
