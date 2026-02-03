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

type ReconciliationEntry struct {
	ID              uuid.UUID
	ProviderMatchID string
	Provider        entity.Provider
	ReconciledAt    *time.Time
	Tries           int
}

func SaveToReconciliationQueue(ctx context.Context, providerMatchID string, provider entity.Provider) error {
	if config.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, NULL, 0)
		ON CONFLICT (provider_match_id, provider) DO NOTHING
	`
	_, err := config.DB.ExecContext(ctx, query, uuid.New().String(), providerMatchID, provider)
	if err != nil {
		return fmt.Errorf("failed to insert into reconciliation queue: %w", err)
	}

	return nil
}

// This will be useful when writing the reconciliation logic.
// func GetPendingReconciliations(ctx context.Context, limit int, maxTries int) ([]ReconciliationEntry, error) {
// 	if config.DB == nil {
// 		return nil, fmt.Errorf("database connection not initialized")
// 	}

// 	query := `
// 		SELECT id, provider_match_id, provider, reconciled_at, tries
// 		FROM match_reconciliation
// 		WHERE reconciled_at IS NULL AND tries < $1
// 		ORDER BY id
// 		LIMIT $2
// 	`

// 	rows, err := config.DB.QueryContext(ctx, query, maxTries, limit)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query pending reconciliations: %w", err)
// 	}
// 	defer rows.Close()

// 	var entries []ReconciliationEntry
// 	for rows.Next() {
// 		var entry ReconciliationEntry
// 		var reconciledAt sql.NullTime
// 		err := rows.Scan(&entry.ID, &entry.ProviderMatchID, &entry.Provider, &reconciledAt, &entry.Tries)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan reconciliation entry: %w", err)
// 		}
// 		if reconciledAt.Valid {
// 			entry.ReconciledAt = &reconciledAt.Time
// 		}
// 		entries = append(entries, entry)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error iterating reconciliation entries: %w", err)
// 	}

// 	return entries, nil
// }

// This will be useful when writing the reconciliation logic.
// func MarkReconciled(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
// 	query := `
// 		UPDATE match_reconciliation
// 		SET reconciled_at = NOW()
// 		WHERE id = $1
// 	`
// 	_, err := tx.ExecContext(ctx, query, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to mark as reconciled: %w", err)
// 	}
// 	return nil
// }

func IncrementTries(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `
		UPDATE match_reconciliation
		SET tries = tries + 1
		WHERE id = $1
	`
	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment tries: %w", err)
	}

	return nil
}

func FindByProviderMatchID(ctx context.Context, providerMatchID string, provider entity.Provider) (*ReconciliationEntry, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, provider_match_id, provider, reconciled_at, tries
		FROM match_reconciliation
		WHERE provider_match_id = $1 AND provider = $2
	`

	var entry ReconciliationEntry
	var reconciledAt sql.NullTime
	err := config.DB.QueryRowContext(ctx, query, providerMatchID, provider).Scan(
		&entry.ID,
		&entry.ProviderMatchID,
		&entry.Provider,
		&reconciledAt,
		&entry.Tries,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find reconciliation entry: %w", err)
	}

	if reconciledAt.Valid {
		entry.ReconciledAt = &reconciledAt.Time
	}

	return &entry, nil
}
