ALTER TABLE project_changes ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ;

-- Partial index — only indexes unnotified rows, keeps it fast as the table grows
CREATE INDEX IF NOT EXISTS idx_project_changes_unnotified
    ON project_changes(detected_at)
    WHERE notified_at IS NULL;
