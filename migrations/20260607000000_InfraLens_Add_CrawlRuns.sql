CREATE TABLE IF NOT EXISTS crawl_runs (
    id           SERIAL PRIMARY KEY,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at  TIMESTAMPTZ,
    status       TEXT NOT NULL DEFAULT 'running', -- 'running' | 'completed' | 'completed_with_errors' | 'failed'
    start_id     INTEGER NOT NULL,
    end_id       INTEGER NOT NULL,
    processed    INTEGER NOT NULL DEFAULT 0,
    failed       INTEGER NOT NULL DEFAULT 0,
    error        TEXT
);

CREATE INDEX IF NOT EXISTS idx_crawl_runs_started_at ON crawl_runs(started_at DESC);
