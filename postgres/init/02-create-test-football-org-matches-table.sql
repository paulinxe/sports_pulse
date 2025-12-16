CREATE TABLE test_football_org_matches (
    id BIGSERIAL PRIMARY KEY,
    competition_id INTEGER NOT NULL,
    utc_date TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_team_score INTEGER NOT NULL DEFAULT 0,
    away_team_score INTEGER NOT NULL DEFAULT 0,
    matchday INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_test_football_org_matches_date_range ON test_football_org_matches(competition_id, utc_date);
