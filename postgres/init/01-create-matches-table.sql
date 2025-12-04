-- Create status enum type
CREATE TYPE match_status AS ENUM ('pending', 'processing', 'finished', 'signed', 'submitted', 'stored');

-- Create matches table
CREATE TABLE matches (
    id VARCHAR(255) PRIMARY KEY,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    start TIMESTAMP NOT NULL,
    "end" TIMESTAMP NOT NULL,
    status match_status NOT NULL,
    home_team_score INTEGER DEFAULT 0,
    away_team_score INTEGER DEFAULT 0,
    provider_match_id VARCHAR(255) NOT NULL,
    provider INTEGER NOT NULL,
    transaction_hash VARCHAR(255),
    signature VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Create index on status field
CREATE INDEX idx_matches_status ON matches(status);

