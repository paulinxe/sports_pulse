package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"provider/internal/entity"
	"time"

	"github.com/google/uuid"
)

// MatchRepository handles match persistence.
type MatchRepository struct {
	db *sql.DB
}

func NewMatchRepository(db *sql.DB) (*MatchRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	return &MatchRepository{db: db}, nil
}

// TODO: find a way to unify this with the SaveInTx function.
func (r *MatchRepository) Save(ctx context.Context, match entity.Match) error {
	query := `
        INSERT INTO matches (
            id, canonical_id, home_team_id, away_team_id, start, "end", status,
            home_team_score, away_team_score, provider_match_id, competition_id,
            provider
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (provider, provider_match_id) DO UPDATE SET
            status = EXCLUDED.status,
            home_team_score = EXCLUDED.home_team_score,
            away_team_score = EXCLUDED.away_team_score,
            updated_at = now()
    `

	_, err := r.db.ExecContext(ctx, query,
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

// TODO: find a way to unify this with the Save function.
func (r *MatchRepository) SaveInTx(ctx context.Context, tx *sql.Tx, match entity.Match) error {
	query := `
        INSERT INTO matches (
            id, canonical_id, home_team_id, away_team_id, start, "end", status,
            home_team_score, away_team_score, provider_match_id, competition_id,
            provider
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (provider, provider_match_id) DO UPDATE SET
            status = EXCLUDED.status,
            home_team_score = EXCLUDED.home_team_score,
            away_team_score = EXCLUDED.away_team_score,
            updated_at = now()
    `

	_, err := tx.ExecContext(ctx, query,
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

func (r *MatchRepository) FindByCanonicalID(ctx context.Context, canonicalID string, provider entity.Provider) (*entity.Match, error) {
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

	err := r.db.QueryRowContext(ctx, query, canonicalID, provider).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find match by canonical_id %s and provider %v: %w", canonicalID, provider, err)
	}

	return &match, nil
}

func (r *MatchRepository) FindMostRecentTimestamp(ctx context.Context, competition entity.Competition, provider entity.Provider) (*time.Time, error) {
	query := `
        SELECT start
        FROM matches
        WHERE competition_id = $1 AND provider = $2
        ORDER BY start DESC
        LIMIT 1
    `
	var timestamp time.Time
	err := r.db.QueryRowContext(ctx, query, competition, provider).Scan(&timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find most recent timestamp: %w", err)
	}

	return &timestamp, nil
}

func (r *MatchRepository) FinishMatch(ctx context.Context, matchID uuid.UUID, homeTeamScore uint, awayTeamScore uint) error {
	query := `
        UPDATE matches SET status = $1, home_team_score = $2, away_team_score = $3, updated_at = now() WHERE id = $4
    `
	_, err := r.db.ExecContext(ctx, query, entity.Finished, homeTeamScore, awayTeamScore, matchID)
	return err
}