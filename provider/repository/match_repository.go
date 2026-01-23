package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"provider/db"
	"provider/entity"
	"time"

	"github.com/google/uuid"
)

func Save(ctx context.Context, match entity.Match) error {
	if db.DB == nil {
		slog.Warn("Database connection not initialized, skipping insert")
		return nil
	}

	// At the moment, if we have a conflict (same match provided by different providers), we skip the insert.
	// On future versions we would need some kind of consensus mechanism to handle this.
	query := `
        INSERT INTO matches (
            id, canonical_id, home_team_id, away_team_id, start, "end", status,
            home_team_score, away_team_score, provider_match_id, competition_id,
            provider
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (canonical_id, competition_id) DO NOTHING
    `

	_, err := db.DB.ExecContext(ctx, query,
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
		return fmt.Errorf("failed to insert match %s: %w", match.ProviderMatchID, err)
	}

	slog.Debug("Inserted match (or skipped due to conflict)",
		"id", match.ID,
		"canonical_id", match.CanonicalID)

	return nil
}

func FindByCanonicalID(ctx context.Context, canonicalID string, provider entity.Provider) (*entity.Match, error) {
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

	err := db.DB.QueryRowContext(ctx, query, canonicalID, provider).Scan(
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
		return nil, fmt.Errorf("failed to find match by canonical_id %s and provider %v: %w", canonicalID, provider, err)
	}

	return &match, nil
}

func FindMostRecentTimestamp(ctx context.Context, competition entity.Competition, provider entity.Provider) (*time.Time, error) {
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
	err := db.DB.QueryRowContext(ctx, query, competition, provider).Scan(&timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find most recent timestamp: %w", err)
	}

	return &timestamp, nil
}

// We only delete matches that are in Pending or InProgress status to avoid deleting matches
// that are already signed or that have been broadcast to the blockchain.
func DeleteByCanonicalID(ctx context.Context, canonicalID string, provider entity.Provider) error {
	if db.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
        DELETE FROM matches
        WHERE canonical_id = $1 AND provider = $2
        AND status IN ($3, $4)
    `
	_, err := db.DB.ExecContext(ctx, query, canonicalID, provider, entity.Pending, entity.InProgress)
	return err
}

func FinishMatch(ctx context.Context, matchID uuid.UUID, homeTeamScore uint, awayTeamScore uint) error {
	if db.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
        UPDATE matches SET status = $1, home_team_score = $2, away_team_score = $3, updated_at = now() WHERE id = $4
    `
	_, err := db.DB.ExecContext(ctx, query, entity.Finished, homeTeamScore, awayTeamScore, matchID)
	return err
}
