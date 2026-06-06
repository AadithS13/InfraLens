CREATE TABLE IF NOT EXISTS promoters (
    id               SERIAL PRIMARY KEY,
    user_profile_id  INTEGER UNIQUE NOT NULL,
    name             TEXT,
    pan              TEXT,
    gstin            TEXT,
    promoter_type    TEXT,
    raw_json         JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS addresses (
    id           SERIAL PRIMARY KEY,
    entity_type  TEXT NOT NULL, -- 'project' | 'promoter'
    entity_id    INTEGER NOT NULL,
    line1        TEXT,
    line2        TEXT,
    city         TEXT,
    district     TEXT,
    state        TEXT,
    pincode      TEXT,
    raw_json     JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS projects (
    id                       SERIAL PRIMARY KEY,
    maha_id                  INTEGER UNIQUE NOT NULL,
    rera_registration_no     TEXT UNIQUE,
    acknowledgement_number   TEXT,
    project_name             TEXT,
    project_type             TEXT,
    project_status           TEXT,
    project_current_status   TEXT,
    rera_registration_date   DATE,
    proposed_completion_date DATE,
    original_completion_date DATE,
    total_units              INTEGER,
    total_sold_units         INTEGER,
    user_profile_id          INTEGER,
    promoter_id              INTEGER REFERENCES promoters(id),
    is_migrated              BOOLEAN,
    raw_json                 JSONB,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contacts (
    id          SERIAL PRIMARY KEY,
    project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role        TEXT,   -- 'Architect' | 'Engineer' | 'Agent' | 'AuthorizedPerson'
    name        TEXT,
    phone       TEXT,
    email       TEXT,
    raw_json    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- foundation for V2 change detection, costs nothing now
CREATE TABLE IF NOT EXISTS project_snapshots (
    id          SERIAL PRIMARY KEY,
    project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checksum    TEXT NOT NULL,
    raw_json    JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_maha_id          ON projects(maha_id);
CREATE INDEX IF NOT EXISTS idx_projects_rera_no          ON projects(rera_registration_no);
CREATE INDEX IF NOT EXISTS idx_snapshots_project_id      ON project_snapshots(project_id);
CREATE INDEX IF NOT EXISTS idx_addresses_entity          ON addresses(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_contacts_project_id       ON contacts(project_id);
