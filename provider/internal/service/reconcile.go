package service

import (
	"context"
	"fmt"
	"log/slog"
	"provider/internal/entity"
	"provider/internal/football_org"
	"provider/internal/repository"
	"time"

	"github.com/google/uuid"
)

const (
	RECONCILE_BATCH_SIZE      = 10
	RECONCILE_MAX_TRIES       = 5
	RECONCILE_REQUEST_DELAY   = 7 * time.Second // Throttle requests to avoid rate limiting
	RECONCILE_CONTEXT_TIMEOUT = 2 * time.Minute
)

// ReconcileProvider extends SyncProvider with the ability to fetch a single match by ID (for reconciliation)
type ReconcileProvider interface {
	SyncProvider
	FetchMatchByID(ctx context.Context, providerMatchID string) (*entity.Match, error)
}

// Reconcile processes all pending matches in the reconciliation queue.
// It fetches 10 items at a time, processes them sequentially, then fetches the next batch
// until the queue is empty. A 2-minute context timeout helps catch API timeouts.
// It is provider-agnostic: each entry's provider field determines which API to call.
func Reconcile(repositories *repository.Repositories) error {
	ctx, cancel := context.WithTimeout(context.Background(), RECONCILE_CONTEXT_TIMEOUT)
	defer cancel()

	for {
		entries, err := repositories.Reconciliation.GetPendingReconciliations(ctx, RECONCILE_BATCH_SIZE, RECONCILE_MAX_TRIES)
		if err != nil {
			slog.Error("Failed to get pending reconciliations", "error", err)
			return fmt.Errorf("failed to get pending reconciliations: %w", err)
		}

		if len(entries) == 0 {
			slog.Debug("No pending reconciliations")
			return nil
		}

		slog.Info("Processing reconciliation batch", "count", len(entries))

		for i, entry := range entries {
			// As this function is not context aware, we need to check if the context is done before each request
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Throttle: sleep before each request except the first (to stay under 10 req/min)
			if i > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(RECONCILE_REQUEST_DELAY):
				}
			}

			reconcileProvider, err := getProviderForReconcile(entry.Provider, repositories)
			if err != nil {
				slog.Error("unable to get provider for reconciliation. manual intervention required.",
					"provider_match_id", entry.ProviderMatchID,
					"provider", entry.Provider,
					"error", err)

				incrementTries(ctx, repositories, &entry)
				continue
			}

			match, err := reconcileProvider.FetchMatchByID(ctx, entry.ProviderMatchID)
			if err != nil {
				slog.Warn("failed to fetch match for reconciliation",
					"provider_match_id", entry.ProviderMatchID,
					"error", err)
				incrementTries(ctx, repositories, &entry)
				continue
			}

			if match.Status != entity.Finished {
				slog.Debug("match not yet finished, will retry later",
					"provider_match_id", entry.ProviderMatchID,
					"status", match.Status)

				// TODO: add a delay to the next reconciliation

				incrementTries(ctx, repositories, &entry)
				continue
			}

			// Save match and remove from queue atomically. Data is valid at this point; failures are DB-related.
			if err := saveAndRemoveFromQueue(ctx, repositories, *match, entry.ID); err != nil {
				slog.Error("failed to save reconciled match",
					"provider_match_id", entry.ProviderMatchID,
					"error", err)
				continue
			}

			slog.Info("reconciled match",
				"provider_match_id", entry.ProviderMatchID,
				"provider", entry.Provider)
		}
	}
}

func getProviderForReconcile(provider entity.Provider, repositories *repository.Repositories) (ReconcileProvider, error) {
	switch provider {
	case entity.FootballOrg:
		return football_org.NewProvider(repositories.Match, repositories.Reconciliation), nil
	default:
		return nil, fmt.Errorf("unknown provider: %v", provider)
	}
}

func incrementTries(ctx context.Context, repositories *repository.Repositories, entry *repository.ReconciliationEntry) {
	if incErr := repositories.Reconciliation.IncrementTries(ctx, entry.ID); incErr != nil {
		slog.Error("failed to increment tries for reconciliation entry",
			"entry_id", entry.ID,
			"provider_match_id", entry.ProviderMatchID,
			"error", incErr)
	}
}

func saveAndRemoveFromQueue(ctx context.Context, repositories *repository.Repositories, match entity.Match, entryID uuid.UUID) error {
	tx, err := repositories.Reconciliation.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction for reconciliation",
			"entry_id", entryID,
			"error", err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := repositories.Match.SaveInTx(ctx, tx, match); err != nil {
		slog.Error("failed to save match in reconciliation transaction",
			"entry_id", entryID,
			"provider_match_id", match.ProviderMatchID,
			"error", err)
		return err
	}

	if err := repositories.Reconciliation.MarkReconciled(ctx, tx, entryID); err != nil {
		slog.Error("failed to remove entry from reconciliation queue",
			"entry_id", entryID,
			"error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit reconciliation transaction",
			"entry_id", entryID,
			"error", err)
		return err
	}

	return nil
}
