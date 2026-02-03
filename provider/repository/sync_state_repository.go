package repository

import (
	"context"
	"database/sql"
	"fmt"
	"provider/config"
	"provider/entity"
	"time"

	"github.com/google/uuid"
)

func GetLastSyncedDate(ctx context.Context, competition entity.Competition, provider entity.Provider) (*time.Time, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT last_synced_date
		FROM sync_state
		WHERE competition_id = $1 AND provider = $2
	`
	var lastSyncedDate time.Time
	err := config.DB.QueryRowContext(ctx, query, competition, provider).Scan(&lastSyncedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last synced date: %w", err)
	}

	return &lastSyncedDate, nil
}

func UpdateLastSyncedDate(
	ctx context.Context,
	competition entity.Competition,
	provider entity.Provider,
	date time.Time,
) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO sync_state (id, competition_id, provider, last_synced_date, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (competition_id, provider)
		DO UPDATE SET last_synced_date = $4, updated_at = NOW()
	`
	_, err := config.DB.ExecContext(ctx, query, uuid.New().String(), competition, provider, date)
	if err != nil {
		return fmt.Errorf("failed to update last synced date: %w", err)
	}

	return nil
}
