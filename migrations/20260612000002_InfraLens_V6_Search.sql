-- V6: Search Intelligence
-- Enable pg_trgm and add GIN indexes for word_similarity() queries.
-- These indexes make ?q= full-text search and /search/suggestions fast
-- even as the dataset grows into hundreds of thousands of rows.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Project name search (primary search surface)
CREATE INDEX IF NOT EXISTS idx_projects_name_trgm
    ON projects USING GIN (project_name gin_trgm_ops);

-- Promoter name search (builder name queries like ?q=prestige)
CREATE INDEX IF NOT EXISTS idx_promoters_name_trgm
    ON promoters USING GIN (name gin_trgm_ops);

-- District search (location queries like ?q=nagpur)
CREATE INDEX IF NOT EXISTS idx_addresses_district_trgm
    ON addresses USING GIN (district gin_trgm_ops);
