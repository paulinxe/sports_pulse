package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"signer/internal/entity"
)

type MatchRepository struct {
	db *sql.DB
}

func NewMatchRepository(db *sql.DB) (*MatchRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection must not be nil")
	}

	return &MatchRepository{db: db}, nil
}

func (r *MatchRepository) FindMatchesToSign(ctx context.Context) ([]entity.Match, error) {
	query := `
        SELECT id, canonical_id, home_team_score, away_team_score
        FROM matches 
        WHERE status = $1
    `
	rows, err := r.db.QueryContext(ctx, query, entity.Finished)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matches: %w", err)
	}

	return matches, nil
}

func (r *MatchRepository) StoreSignature(ctx context.Context, match entity.Match, signature string) error {
	query := `
        UPDATE matches
        SET signature = $1, status = $2
        WHERE id = $3
    `
	_, err := r.db.ExecContext(ctx, query, signature, entity.Signed, match.ID)
	if err != nil {
		return fmt.Errorf("failed to store signature: %w", err)
	}

	return nil
}
