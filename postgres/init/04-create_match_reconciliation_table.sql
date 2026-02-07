-- Create match_reconciliation table for dead-letter queue
-- This table stores matches that need reconciliation (e.g., stale in-progress matches)

CREATE TABLE IF NOT EXISTS match_reconciliation (
    id UUID PRIMARY KEY,
    provider_match_id TEXT NOT NULL,
    provider INTEGER NOT NULL,
    reconciled_at TIMESTAMP NULL,
    tries INTEGER NOT NULL DEFAULT 0,
    UNIQUE (provider_match_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_match_reconciliation_pending ON match_reconciliation (reconciled_at, tries);
