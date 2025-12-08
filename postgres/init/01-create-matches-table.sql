-- Create status enum type
CREATE TYPE match_status AS ENUM ('pending', 'processing', 'finished', 'signed', 'submitted', 'stored');

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create matches table
CREATE TABLE matches (
    id UUID PRIMARY KEY,
    canonical_id VARCHAR(255) NOT NULL,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    start TIMESTAMP NOT NULL,
    "end" TIMESTAMP NOT NULL,
    status match_status NOT NULL,
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

-- Create index on status field
CREATE INDEX idx_matches_status ON matches(status);

-- Create unique index on canonical_id and competition_id
CREATE UNIQUE INDEX idx_matches_canonical_competition ON matches(canonical_id, competition_id);

