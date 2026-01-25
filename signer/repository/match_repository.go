package repository

import (
	"context"
	"fmt"
	"log/slog"
	"signer/db"
	"signer/entity"
)

func FindMatchesToSign(ctx context.Context) ([]entity.Match, error) {
    if db.DB == nil {
        return nil, fmt.Errorf("database connection not initialized")
    }

    query := `
        SELECT id, canonical_id, home_team_score, away_team_score
        FROM matches 
        WHERE status = $1
    `
    rows, err := db.DB.QueryContext(ctx, query, entity.Finished)
    if err != nil {
        return nil, fmt.Errorf("failed to find matches to sign: %w", err)
    }
    defer func() { _ = rows.Close() }()

    var matches []entity.Match
    for rows.Next() {
        var match entity.Match
        err := rows.Scan(&match.ID, &match.CanonicalID, &match.HomeTeamScore, &match.AwayTeamScore)
        if err != nil {
            slog.Error("failed to load match", "error", err)
            continue
        }

        matches = append(matches, match)
    }

    return matches, nil
}

func StoreSignature(ctx context.Context, match entity.Match, signature string) error {
    if db.DB == nil {
        return fmt.Errorf("database connection not initialized")
    }

    query := `
        UPDATE matches
        SET signature = $1, status = $2
        WHERE id = $3
    `
    _, err := db.DB.ExecContext(ctx, query, signature, entity.Signed, match.ID)
    if err != nil {
        return fmt.Errorf("failed to store signature: %w", err)
    }

    return err
}