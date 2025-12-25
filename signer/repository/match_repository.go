package repository

import (
    "fmt"
    "signer/db"
    "signer/entity"
    "log/slog"
)

func FindMatchesToSign() ([]entity.Match, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }

    query := `
        SELECT id, canonical_id, home_team_score, away_team_score
        FROM matches 
        WHERE status = $1
    `
    rows, err := db.DB.Query(query, entity.Finished)
    if err != nil {
        return nil, fmt.Errorf("failed to find matches to sign: %v", err)
    }
    defer rows.Close()

    var matches []entity.Match
    for rows.Next() {
        var match entity.Match
        err := rows.Scan(&match.ID, &match.CanonicalID, &match.HomeTeamScore, &match.AwayTeamScore)
        if err != nil {
            slog.Error("Failed to load match", "error", err)
            continue
        }

        matches = append(matches, match)
    }

    return matches, nil
}

func StoreSignature(match entity.Match, signature string) error {
    if db.DB == nil {
        return fmt.Errorf("database connection not initialized")
    }

    query := `
        UPDATE matches
        SET signature = $1, status = $2
        WHERE id = $3
    `
    _, err := db.DB.Exec(query, signature, entity.Signed, match.ID)
    if err != nil {
        return fmt.Errorf("failed to store signature: %v", err)
    }

    return err
}