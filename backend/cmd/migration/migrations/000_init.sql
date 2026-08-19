-- Init creates the migrations table
-- v.000
-- 08/18/2026

CREATE TABLE IF NOT EXISTS migrations (
    version VARCHAR(15) PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
