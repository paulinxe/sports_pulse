package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"provider/internal/entity"
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

type ReconciliationRepository struct {
	db *sql.DB
}

func NewReconciliationRepository(db *sql.DB) (*ReconciliationRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	return &ReconciliationRepository{db: db}, nil
}

func (r *ReconciliationRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

func (r *ReconciliationRepository) SaveToReconciliationQueue(ctx context.Context, providerMatchID string, provider entity.Provider) error {
	query := `
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, NULL, 0)
		ON CONFLICT (provider_match_id, provider) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New().String(), providerMatchID, provider)
	if err != nil {
		return fmt.Errorf("failed to insert into reconciliation queue: %w", err)
	}

	return nil
}

func (r *ReconciliationRepository) GetPendingReconciliations(ctx context.Context, limit, maxTries int) ([]ReconciliationEntry, error) {
	query := `
		SELECT id, provider_match_id, provider, reconciled_at, tries
		FROM match_reconciliation
		WHERE tries < $1
		ORDER BY id
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, maxTries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending reconciliations: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var entries []ReconciliationEntry
	for rows.Next() {
		var entry ReconciliationEntry
		var reconciledAt sql.NullTime
		err := rows.Scan(&entry.ID, &entry.ProviderMatchID, &entry.Provider, &reconciledAt, &entry.Tries)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reconciliation entry: %w", err)
		}
		if reconciledAt.Valid {
			entry.ReconciledAt = &reconciledAt.Time
		}
		entries = append(entries, entry)
	}


	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reconciliation entries: %w", err)
	}

	return entries, nil
}

func (r *ReconciliationRepository) MarkReconciled(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `DELETE FROM match_reconciliation WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove from reconciliation queue: %w", err)
	}

	return nil
}

// IncrementTries updates the tries count for an entry. Use when not inside a transaction.
func (r *ReconciliationRepository) IncrementTries(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE match_reconciliation
		SET tries = tries + 1, reconciled_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment tries: %w", err)
	}

	return nil
}

// IncrementTriesInTx updates the tries count within the given transaction. Use when part of a larger transaction.
func (r *ReconciliationRepository) IncrementTriesInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `
		UPDATE match_reconciliation
		SET tries = tries + 1, reconciled_at = NOW()
		WHERE id = $1
	`
	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment tries: %w", err)
	}
	return nil
}

func (r *ReconciliationRepository) FindByProviderMatchID(ctx context.Context, providerMatchID string, provider entity.Provider) (*ReconciliationEntry, error) {
	query := `
		SELECT id, provider_match_id, provider, reconciled_at, tries
		FROM match_reconciliation
		WHERE provider_match_id = $1 AND provider = $2
	`

	var entry ReconciliationEntry
	var reconciledAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, providerMatchID, provider).Scan(
		&entry.ID,
		&entry.ProviderMatchID,
		&entry.Provider,
		&reconciledAt,
		&entry.Tries,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find reconciliation entry: %w", err)
	}

	if reconciledAt.Valid {
		entry.ReconciledAt = &reconciledAt.Time
	}

	return &entry, nil
}
