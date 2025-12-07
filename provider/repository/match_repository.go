package repository

import (
    "database/sql"
    "fmt"
    "log/slog"
    "provider/db"
    "provider/entity"
    "time"
)

func Save(match entity.Match) error {
    if db.DB == nil {
        slog.Warn("Database connection not initialized, skipping insert")
        return nil
    }

    query := `
        INSERT INTO matches (
            id, home_team_id, away_team_id, start, "end", status,
            home_team_score, away_team_score, provider_match_id, competition_id,
            provider
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

    result, err := db.DB.Exec(query,
        match.ID,
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
            "match_id", match.ID,
            "error", err)
        return fmt.Errorf("failed to insert match %s: %v", match.ID, err)
    }

    rowsAffected, _ := result.RowsAffected()
    slog.Debug("Inserted match",
        "match_id", match.ID,
        "rows_affected", rowsAffected)

    return nil
}

func FindById(id string) (*entity.Match, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }
    query := `
        SELECT
            id,
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
        WHERE id = $1
    `

    var (
        match entity.Match
    )

    err := db.DB.QueryRow(query, id).Scan(
        &match.ID,
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
        return nil, fmt.Errorf("failed to find match by id %s: %v", id, err)
    }

    return &match, nil
}

func FindMostRecentTimestamp(competition entity.Competition) (*time.Time, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }

    query := `
        SELECT start
        FROM matches
        WHERE competition_id = $1
        ORDER BY start DESC
        LIMIT 1
    `
    var timestamp time.Time
    err := db.DB.QueryRow(query, competition).Scan(&timestamp)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find most recent timestamp: %v", err)
    }
    return &timestamp, nil
}
