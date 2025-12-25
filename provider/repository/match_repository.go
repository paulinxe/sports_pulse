package repository

import (
    "database/sql"
    "fmt"
    "log/slog"
    "provider/db"
    "provider/entity"
    "time"
    "github.com/google/uuid"
)

func Save(transaction *sql.Tx, match entity.Match) error {
    if db.DB == nil {
        slog.Warn("Database connection not initialized, skipping insert")
        return nil
    }

    query := `
        INSERT INTO matches (
            id, canonical_id, home_team_id, away_team_id, start, "end", status,
            home_team_score, away_team_score, provider_match_id, competition_id,
            provider
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

    _, err := transaction.Exec(query,
        match.ID,
        match.CanonicalID,
        match.HomeTeamID,
        match.AwayTeamID,
        match.Start,
        match.End,
        match.Status,
        match.HomeTeamScore,
        match.AwayTeamScore,
        match.ProviderMatchID,
        match.CompetitionID,
        match.Provider,
    )

    if err != nil {
        slog.Error("Failed to insert match",
            "id", match.ID,
            "provider_match_id", match.ProviderMatchID,
            "error", err)
        return fmt.Errorf("failed to insert match %s: %v", match.ProviderMatchID, err)
    }

    slog.Debug("Inserted match",
        "id", match.ID,
        "canonical_id", match.CanonicalID)

    return nil
}

func FindByCanonicalID(canonicalID string, provider entity.Provider) (*entity.Match, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }
    query := `
        SELECT
            id,
            canonical_id,
            start,
            "end",
            status,
            provider,
            provider_match_id,
            competition_id,
            home_team_id,
            away_team_id,
            home_team_score,
            away_team_score
        FROM matches 
        WHERE canonical_id = $1 AND provider = $2
    `

    var (
        match entity.Match
    )

    err := db.DB.QueryRow(query, canonicalID, provider).Scan(
        &match.ID,
        &match.CanonicalID,
        &match.Start,
        &match.End,
        &match.Status,
        &match.Provider,
        &match.ProviderMatchID,
        &match.CompetitionID,
        &match.HomeTeamID,
        &match.AwayTeamID,
        &match.HomeTeamScore,
        &match.AwayTeamScore,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find match by canonical_id %s and provider %v: %v", canonicalID, provider, err)
    }

    return &match, nil
}

func FindMostRecentTimestamp(competition entity.Competition, provider entity.Provider) (*time.Time, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }

    query := `
        SELECT start
        FROM matches
        WHERE competition_id = $1 AND provider = $2
        ORDER BY start DESC
        LIMIT 1
    `
    var timestamp time.Time
    err := db.DB.QueryRow(query, competition, provider).Scan(&timestamp)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }

        return nil, fmt.Errorf("failed to find most recent timestamp: %v", err)
    }

    return &timestamp, nil
}

func DeleteByCanonicalID(canonicalID string, provider entity.Provider) error {
    if db.DB == nil {
        return fmt.Errorf("database connection not initialized")
    }

    query := `
        DELETE FROM matches WHERE canonical_id = $1 AND provider = $2 AND status = $3
    `
    _, err := db.DB.Exec(query, canonicalID, provider, entity.Pending)
    return err
}

func FindMatchesToReconcile(provider entity.Provider, start time.Time, end time.Time) ([]entity.ReconciliableMatch, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }

    query := `
        SELECT id, provider_match_id, home_team_score, away_team_score
        FROM matches 
        WHERE provider = $1 AND status = $2 AND "end" >= $3 AND "end" <= $4
    `
    rows, err := db.DB.Query(query, provider, entity.Pending, start, end)
    if err != nil {
        return nil, fmt.Errorf("failed to find matches to reconcile: %v", err)
    }
    defer rows.Close()

    var matches []entity.ReconciliableMatch
    for rows.Next() {
        var match entity.ReconciliableMatch
        err := rows.Scan(&match.ID, &match.ProviderMatchID, &match.HomeTeamScore, &match.AwayTeamScore)
        if err != nil {
            slog.Error("Failed to load match", "error", err)
            continue
        }

        matches = append(matches, match)
    }

    return matches, nil
}

func FinishMatch(matchID uuid.UUID, homeTeamScore uint, awayTeamScore uint) error {
    if db.DB == nil {
        return fmt.Errorf("database connection not initialized")
    }

    query := `
        UPDATE matches SET status = $1, home_team_score = $2, away_team_score = $3, updated_at = now() WHERE id = $4
    `
    _, err := db.DB.Exec(query, entity.Finished, homeTeamScore, awayTeamScore, matchID)
    return err
}