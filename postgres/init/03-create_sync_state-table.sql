CREATE TABLE IF NOT EXISTS sync_state (
    id UUID PRIMARY KEY,
    competition_id INTEGER NOT NULL,
    provider INTEGER NOT NULL,
    last_synced_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(competition_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_sync_state_competition_provider ON sync_state(competition_id, provider);
