# InfraLens

A construction intelligence platform that crawls public RERA registries, normalizes project and promoter metadata, tracks field-level changes, fires notifications when high-value fields shift, and exposes a REST API for search and analytics.

Inspired by platforms like Biltrax — built to demonstrate real-world data engineering: web crawling, API reverse engineering, idempotent ingestion, change detection, notification pipelines, and layered API design.

---

## What it does

InfraLens crawls the **MahaRERA** (Maharashtra Real Estate Regulatory Authority) public API and stores structured data about:

- Real estate projects (name, type, status, registration details, completion dates)
- Promoters / developers (name, PAN, GSTIN, type)
- Project and promoter addresses
- Professional contacts (architects, engineers, agents)
- Per-crawl snapshots with MD5 checksums
- Field-level change history — what changed, from what value, to what value, when
- Notifications when high-value fields change (status shifts, deadline slips)
- Analytics — status distribution, top builders by project count, breakdown by district

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
    SCHED[Scheduler\nrobfig/cron] -- triggers --> CR[Create crawl_run record]
    CR --> C
    C -- processed / failed counts --> UPD[Update crawl_run status]
    UPD --> M
    L --> NOTIF[Notification Engine]
    NOTIF -- unnotified changes --> NL[Log to Terminal\nPhase 1]
    NOTIF -- Phase 2 --> EMAIL[Email]
    NOTIF -- Phase 2 --> WEBHOOK[Webhook]
    M --> N[REST API]
    N --> O[Search / Detail / Changes]
    N --> R[Crawl Run History]
    N --> S[Analytics]
```

### Notification Pipeline

```mermaid
flowchart LR
    A[Crawler] --> B[Snapshots]
    B --> C[Diff Engine]
    C --> D[project_changes]
    D --> E[Notification Engine]
    E --> F[Email]
    E --> G[Webhook]
    E --> H[Slack]
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
        timestamp notified_at
    }
    CRAWL_RUNS {
        bigint id PK
        timestamp started_at
        timestamp finished_at
        string status
        int start_id
        int end_id
        int processed
        int failed
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
| `project_changes` | Field-level change log — `field_name`, `old_value`, `new_value`, `detected_at`, `notified_at` |
| `crawl_runs` | Audit log of every scheduled crawl — status, duration, processed/failed counts |

---

## Scheduled Crawling

The API server embeds a cron scheduler (`robfig/cron`) that automatically triggers full crawls on a configurable schedule. Every run is recorded in the `crawl_runs` table and exposed via `GET /api/v1/crawls`.

### How it works

1. On server startup, the scheduler registers a cron job and authenticates with MahaRERA
2. On each tick: creates a `crawl_run` record (`status=running`), runs the full worker-pool crawler, then updates the record with final stats
3. Token auto-refresh is handled transparently — the crawler re-authenticates before expiry

### Scheduler log output

The scheduler fires automatically, authenticates with MahaRERA, and processes each project through the full change-detection pipeline:

![Scheduler terminal output](docs/screenshots/crawler_terminal.png)

### `GET /api/v1/crawls` — live proof

Every run is recorded in `crawl_runs` and accessible via the API. After running for a few minutes with `@every 1m`, the history shows all completed runs with timestamps and counts:

![GET /api/v1/crawls response](docs/screenshots/crawler_browser.png)

Each run processed 10 projects in ~3 seconds with zero failures. The `status` field is `completed_with_errors` if any projects failed, or `failed` if the run could not start.

### Configuring the schedule

```bash
# Default: every night at 2am
DATABASE_URL="..." PORT=8080 go run ./cmd/api/

# Custom: every hour
DATABASE_URL="..." CRAWL_SCHEDULE="0 * * * *" START_ID=1 END_ID=50000 PORT=8080 go run ./cmd/api/

# Testing: every minute
DATABASE_URL="..." CRAWL_SCHEDULE="@every 1m" START_ID=1 END_ID=10 PORT=8080 go run ./cmd/api/
```

---

## Notification Engine

After every crawl run, the notification engine scans `project_changes` for rows where `notified_at IS NULL` and `field_name` is one of the watched fields. It fans the change out to every registered adapter, then stamps `notified_at = NOW()` so the same change is never re-notified.

### Notification Pipeline

Four stages — from database row to delivered alert:

**Step 1 — Change detected & notification dispatched**

The scheduler detects an unnotified row for a watched field and fires the notifier. A `[NOTIFY]` block is printed to the terminal immediately:

![Change detected — notifier terminal output](docs/screenshots/notifier_terminal.png)

**Step 2 — Webhook received**

Simultaneously, the `WebhookAdapter` POSTs a JSON payload. The full payload arrives at the endpoint within milliseconds, with `InfraLens-Notifier/1.0` as the user-agent:

![Webhook received — webhook.site](docs/screenshots/notifier_webhook.png)

**Step 3 — Email delivered**

The `EmailAdapter` sends via SMTP (port 587 / STARTTLS). The email lands in the inbox with project name, field label, old → new values, and timestamp:

![Email delivered — Mailtrap inbox](docs/screenshots/notifier_email.png)

All three adapters run in the same dispatch loop. A failed adapter logs a warning but never blocks the others or causes a double-notify.

### Why not notify everything?

Most field changes are noise. Only three fields have direct business value:

| Field | Why it matters |
|---|---|
| `project_status` | A status shift (e.g. `New → Ongoing`) means a project has moved forward |
| `project_current_status` | Tracks internal approval milestones |
| `proposed_completion_date` | A date pushed forward means a builder is slipping |

### Configuring delivery adapters

```bash
# Email only (Gmail, Mailtrap, any SMTP)
SMTP_HOST=sandbox.smtp.mailtrap.io \
SMTP_PORT=587 \
SMTP_USER=<user> \
SMTP_PASS=<pass> \
NOTIFY_EMAIL_FROM=infralens@notify.com \
NOTIFY_EMAIL_TO=you@example.com \
go run ./cmd/api/

# Webhook only (Slack, n8n, custom endpoint)
NOTIFY_WEBHOOK_URL=https://hooks.slack.com/... go run ./cmd/api/

# Both at once
SMTP_HOST=... NOTIFY_WEBHOOK_URL=... go run ./cmd/api/
```

If neither is set, only the `LogAdapter` runs — terminal output as in Phase 1.

---

## Search & Discovery (V6)

Full-text search across project names, promoter names, and districts — powered by PostgreSQL's `pg_trgm` extension. Results are ranked by relevance score so the most likely matches always appear first.

### `GET /api/v1/projects?q=<term>` — relevance-ranked search

Add `?q=` to any project list query. The engine computes `word_similarity()` against project name, promoter name, and district simultaneously. Results include a `relevance` score (0–1) so you can see how well each result matched.

```bash
# Find all projects matching "pramukh" — fuzzy, case-insensitive
curl "localhost:8080/api/v1/projects?q=pramukh&limit=3"
```

```json
{
  "data": [
    {
      "id": 18,
      "project_name": "Pramukh Sneh Phase 2",
      "promoter_name": "Pramukh Realty",
      "district": "Dadra & Nagar Haveli",
      "project_status": "Ongoing",
      "relevance": 1
    },
    {
      "id": 28,
      "project_name": "PRAMUKH GARDENS PHASE 2",
      "promoter_name": "PRASHANT DEVELOPERS PVT LTD",
      "district": "Dadra & Nagar Haveli",
      "project_status": "Ongoing",
      "relevance": 1
    }
  ],
  "meta": { "page": 1, "limit": 3, "total": 3 }
}
```

Composable with all existing filters — `?q=prestige&district=pune&status=Ongoing` works exactly as expected.

![Search with ?q= — ranked results in browser](docs/screenshots/search_q_browser.png)

---

### `GET /api/v1/search/suggestions?q=<term>` — autocomplete

Returns ranked project names and promoter names for a partial query. Built for typeahead UI — minimum 2 characters, max 20 results. Each suggestion carries a `type` (`"project"` or `"promoter"`) and a `score`.

```bash
curl "localhost:8080/api/v1/search/suggestions?q=pramukh"
```

```json
{
  "data": [
    { "text": "Pramukh Realty",       "type": "promoter", "score": 1 },
    { "text": "Pramukh Sneh Phase 2", "type": "project",  "score": 1 },
    { "text": "PRAMUKH GARDENS PHASE 2", "type": "project", "score": 1 },
    { "text": "PRASHANT DEVELOPERS PVT LTD", "type": "promoter", "score": 0.375 }
  ]
}
```

![Suggestions endpoint — terminal + browser](docs/screenshots/search_suggestions_browser.png)

---

## Natural Language Search (V7.1)

Type plain English. Get structured results. No LLM required — the engine pattern-matches against known intent rules and converts the query into SQL filters on the fly.

Every response includes an `interpreted` field showing exactly which filters were extracted — so you always know what the query did.

### `GET /api/v1/search/nl?q=<query>`

**Show ongoing projects in Pune**

```bash
curl "localhost:8080/api/v1/search/nl?q=show+ongoing+projects+in+Pune"
```

```json
{
  "query": "show ongoing projects in Pune",
  "query_type": "projects",
  "interpreted": { "status": "Ongoing", "district": "Pune" },
  "data": [...],
  "meta": { "page": 1, "limit": 20, "total": 1 }
}
```

![NL search — ongoing projects in Pune](docs/screenshots/nl_ongoing_pune.png)

---

**Show projects delayed beyond original completion date**

```bash
curl "localhost:8080/api/v1/search/nl?q=show+projects+delayed+beyond+original+completion+date"
```

```json
{
  "query": "show projects delayed beyond original completion date",
  "query_type": "projects",
  "interpreted": { "filter": "proposed_completion_date > original_completion_date" },
  "data": [...],
  "meta": { "page": 1, "limit": 20, "total": 314 }
}
```

![NL search — delayed projects](docs/screenshots/nl_delayed.png)

---

**Show builders with more than 2 projects**

Builder queries switch `query_type` to `"builders"` and return a promoter list instead of a project list:

```bash
curl "localhost:8080/api/v1/search/nl?q=show+builders+with+more+than+2+projects"
```

```json
{
  "query": "show builders with more than 2 projects",
  "query_type": "builders",
  "interpreted": { "min_projects": "2" },
  "data": [
    { "promoter_name": "Prestige Group", "project_count": 45, "total_units": 12500 }
  ]
}
```

![NL search — builders query](docs/screenshots/nl_builders.png)

---

**Show residential projects in Thane**

```bash
curl "localhost:8080/api/v1/search/nl?q=show+residential+projects+in+Thane"
```

```json
{
  "query": "show residential projects in Thane",
  "query_type": "projects",
  "interpreted": { "type": "Residential", "district": "Thane" },
  "data": [...],
  "meta": { "page": 1, "limit": 20, "total": 188 }
}
```

![NL search — residential projects in Thane](docs/screenshots/nl_residential_thane.png)

---

### How it works — V7.1 vs V7.2

```mermaid
flowchart TD
    A["User Query\nshow residential projects in Thane"] --> B

    subgraph PARSER ["nlsearch.Parse()  —  the only thing that changes in V7.2"]
        B{"V7.1\nRule-Based"}
        C{"V7.2\nLLM-Powered\n🔜"}
        B --> D["Pattern Matching\nstatus · type · location · delayed · builder"]
        C --> E["Claude API\nnatural language → JSON filter"]
    end

    D --> F["ParsedQuery\n{ interpreted, Filter, QueryType }"]
    E --> F

    F --> G{QueryType?}

    G -- projects --> H["core.SearchFilter\nStatus · Type · District · Delayed · Q"]
    G -- builders --> I["BuildersWithMinProjects\nHAVING count ≥ N"]

    H --> J["repo.Search()\nSELECT … WHERE … ORDER BY relevance"]
    I --> K["repo.BuildersWithMinProjects()\nGROUP BY promoter HAVING count ≥ N"]

    J --> L["JSON Response\n{ query, interpreted, data, meta }"]
    K --> L
```

The parser is the **only swappable part**. The handler, `SearchFilter`, SQL, and response shape are identical in V7.1 and V7.2 — dropping in an LLM is a one-file change.

**Intents recognized in V7.1:**

| Pattern | Example phrase | SQL filter |
|---|---|---|
| Status | "ongoing", "lapsed", "under approval" | `WHERE project_status = ?` |
| Type | "residential", "commercial", "plotted" | `WHERE project_type ILIKE ?` |
| Location | "in Pune", "at Thane", "near Nagpur" | `WHERE district ILIKE ?` |
| Delayed | "delayed", "overdue", "beyond original" | `WHERE proposed > original` |
| Builder count | "builders with more than N projects" | `HAVING count(projects) >= N` |

---

## Analytics API

Three read-only aggregation endpoints over the projects dataset. No query parameters needed — results are pre-grouped and sorted by count descending.

### `GET /api/v1/analytics/status-distribution`

Project counts grouped by `project_status`:

```bash
curl "localhost:8080/api/v1/analytics/status-distribution"
```

```json
{
  "data": [
    { "status": "Ongoing", "count": 14 },
    { "status": "New",     "count": 6  }
  ]
}
```

### `GET /api/v1/analytics/top-builders?limit=5`

Top promoters ranked by project count, with total unit inventory:

```bash
curl "localhost:8080/api/v1/analytics/top-builders?limit=5"
```

```json
{
  "data": [
    { "promoter_name": "Prestige Group",   "project_count": 45, "total_units": 12500 },
    { "promoter_name": "Godrej Properties","project_count": 38, "total_units": 9800  }
  ]
}
```

### `GET /api/v1/analytics/by-district?limit=5`

Project counts by district — useful for market concentration analysis:

```bash
curl "localhost:8080/api/v1/analytics/by-district?limit=5"
```

```json
{
  "data": [
    { "district": "Thane",     "count": 47 },
    { "district": "Nagpur",    "count": 22 },
    { "district": "Aurangabad","count": 15 }
  ]
}
```

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
| `GET` | `/api/v1/projects` | Search projects with filters, pagination, and `?q=` relevance ranking |
| `GET` | `/api/v1/projects/{id}` | Full project detail — promoter, addresses, contacts |
| `GET` | `/api/v1/projects/{id}/changes` | Field-level change history for a project |
| `GET` | `/api/v1/crawls` | Scheduled crawl run history — status, counts, duration |
| `GET` | `/api/v1/analytics/status-distribution` | Project counts grouped by status |
| `GET` | `/api/v1/analytics/top-builders` | Top promoters by project count + total units |
| `GET` | `/api/v1/analytics/by-district` | Project counts grouped by district |
| `GET` | `/api/v1/search/suggestions` | Autocomplete — ranked project + promoter names for `?q=` |
| `GET` | `/api/v1/search/nl` | Natural language search — plain English → SQL filters |
| `GET` | `/health` | Health check |

### Query Parameters — `GET /api/v1/projects`

| Param | Type | Description |
|---|---|---|
| `q` | string | **Full-text search** — ranked by `word_similarity` across project name, promoter, and district. Results include a `relevance` score. |
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
# Full-text search — ranked by relevance
GET /api/v1/projects?q=pramukh

# Fuzzy search combined with a filter
GET /api/v1/projects?q=prestige&district=pune&status=Ongoing

# Autocomplete for a search input
GET /api/v1/search/suggestions?q=pram&limit=5

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
| **V4** | Scheduled crawling — nightly cron, `crawl_runs` tracking, analytics API | ✅ Done |

**V5 — Notification Platform**

| Feature | Status |
|---|---|
| Change subscriptions — `notified_at` tracking, watched fields, dedup | ✅ Done |
| Terminal notifications — formatted `[NOTIFY]` blocks after every crawl | ✅ Done |
| Email notifications — SMTP delivery via pluggable `EmailAdapter` | ✅ Done |
| Webhook notifications — HTTP POST JSON via pluggable `WebhookAdapter` | ✅ Done |

**V6 — Search & Discovery**

| Feature | Status |
|---|---|
| Full-text search — `?q=prestige nagpur` across project + promoter fields | ✅ Done |
| Relevance ranking — `pg_trgm` similarity scoring | ✅ Done |
| Query suggestions — autocomplete across project + promoter names | ✅ Done |

**V7 — AI Intelligence Layer**

| Feature | Status |
|---|---|
| V7.1 Rule-based NL search — status, type, location, delayed, builder count intents | ✅ Done |
| V7.2 LLM-powered query generation — natural language → SQL via Claude API | 🔜 |
| Builder risk insights — delay patterns, status change frequency | 🔜 |
| Completion delay prediction — based on historical change data | 🔜 |

---

## Tech Stack

- **Go 1.23** — crawler, HTTP client, worker pool, REST API
- **chi** — lightweight HTTP router
- **robfig/cron v3** — embedded cron scheduler
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
| `CRAWL_SCHEDULE` | `0 2 * * *` | Cron expression for scheduled crawls (daily at 2am). Use `@every 1m` for testing |
| `START_ID` | `1` | First MahaRERA project ID the scheduler will crawl |
| `END_ID` | `100000` | Last MahaRERA project ID the scheduler will crawl |
| `SMTP_HOST` | — | SMTP host (e.g. `sandbox.smtp.mailtrap.io`, `smtp.gmail.com`) |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USER` | — | SMTP auth username |
| `SMTP_PASS` | — | SMTP auth password / app password |
| `NOTIFY_EMAIL_FROM` | `SMTP_USER` | From address shown in the email |
| `NOTIFY_EMAIL_TO` | — | Comma-separated recipient list. Email adapter activates when set |
| `NOTIFY_WEBHOOK_URL` | — | HTTP POST endpoint for webhook notifications. Adapter activates when set |

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
├── 20260606135001_InfraLens_Add_ProjectChanges.sql # Change detection table
├── 20260607000000_InfraLens_Add_CrawlRuns.sql      # Scheduler audit log
└── 20260612000000_InfraLens_Add_NotifiedAt.sql     # Notification tracking
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
│   ├── scheduler/
│   │   └── scheduler.go         # robfig/cron wrapper, crawl_run lifecycle, calls notifier
│   ├── notifier/
│   │   └── notifier.go          # finds unnotified changes, logs [NOTIFY], marks notified
│   ├── repo/
│   │   └── project.go           # API read layer (search, detail, changes, crawl runs, analytics)
│   ├── core/
│   │   ├── project.go           # ProjectService — business logic + analytics methods
│   │   └── types.go             # DTOs, SearchFilter, CrawlRunItem, analytics types
│   └── server/
│       ├── server.go            # chi router, middleware, graceful shutdown
│       └── handler/
│           ├── project.go       # HTTP handlers — List, Get, Changes
│           ├── crawl.go         # HTTP handler — crawl run history
│           └── analytics.go     # HTTP handlers — status dist, top builders, by district
├── migrations/
└── docker-compose.yml
```
