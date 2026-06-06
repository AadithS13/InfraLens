CREATE TABLE IF NOT EXISTS project_changes (
    id           SERIAL PRIMARY KEY,
    project_id   INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    field_name   TEXT NOT NULL,
    old_value    TEXT,
    new_value    TEXT,
    detected_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_changes_project_id  ON project_changes(project_id);
CREATE INDEX IF NOT EXISTS idx_project_changes_detected_at ON project_changes(detected_at);
CREATE INDEX IF NOT EXISTS idx_project_changes_field_name  ON project_changes(field_name);
