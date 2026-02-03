package repository

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"relayer/config"
	"relayer/entity"
	"log/slog"
)

func FindSignedMatches(ctx context.Context) ([]entity.Match, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, canonical_id, competition_id, home_team_id, away_team_id,
		       home_team_score, away_team_score, start, signature
		FROM matches
		WHERE status = $1
	`
	rows, err := config.DB.QueryContext(ctx, query, entity.SIGNED_STATUS)
	if err != nil {
		return nil, fmt.Errorf("query signed matches: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var matches []entity.Match
	for rows.Next() {
		var match entity.Match
		var start time.Time
		var sigHex string

		err := rows.Scan(
			&match.ID,
			&match.CanonicalID,
			&match.CompetitionID,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.HomeTeamScore,
			&match.AwayTeamScore,
			&start,
			&sigHex,
		)
		if err != nil {
			slog.Error("error parsing match. skipping", "match_id", match.ID, "error", err)
			continue
		}

		// This is to convert a date into a an integer following the format YYYYMMDD
		match.Start = uint32(start.Year()*10000 + int(start.Month())*100 + start.Day())
		sigHex = strings.TrimPrefix(sigHex, "0x")
		match.Signature, err = hex.DecodeString(sigHex)
		if err != nil {
			slog.Error("error parsing signature. skipping", "match_id", match.ID, "error", err)
			continue
		}

		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

func BroadcastMatch(ctx context.Context, matchID uuid.UUID) error {
	if config.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	result, err := config.DB.ExecContext(ctx, `UPDATE matches SET status = $1 WHERE id = $2`, entity.BROADCASTED_STATUS, matchID)
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("match not found: %s", matchID)
	}

	return nil
}
