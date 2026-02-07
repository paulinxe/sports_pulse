CREATE TABLE matches (
    id UUID PRIMARY KEY,
    canonical_id VARCHAR(255) NOT NULL,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    start TIMESTAMP NOT NULL,
    "end" TIMESTAMP NOT NULL,
    status INTEGER NOT NULL,
    home_team_score INTEGER DEFAULT 0,
    away_team_score INTEGER DEFAULT 0,
    provider_match_id VARCHAR(255) NOT NULL,
    competition_id INTEGER NOT NULL,
    provider INTEGER NOT NULL,
    transaction_hash VARCHAR(255),
    signature VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_matches_status ON matches(status);
CREATE UNIQUE INDEX idx_matches_provider_match ON matches(provider, provider_match_id);
