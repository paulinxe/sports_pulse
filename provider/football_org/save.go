package football_org

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/repository"
)

// SaveMatches saves entity.Match objects to the database, filtering to only save finished matches.
func (p *Provider) SaveMatches(ctx context.Context, tx *sql.Tx, matches []entity.Match) error {
	for _, match := range matches {
		// Only save matches that are in finished status
		if match.Status != entity.Finished {
			slog.Debug("Skipping match that is not in final status", "match_id", match.ProviderMatchID, "status", match.Status)
			continue
		}

		// As a match may be rescheduled, we need to delete the existing match in case it already exists.
		// TODO: we should not delete matches that are already signed.
		// This could happen if we reprocess old dates or if we already stored a match during the day.
		if err := repository.DeleteByCanonicalID(ctx, tx, match.CanonicalID, p.GetProviderEntity()); err != nil {
			slog.Error("Failed to delete match", "error", err, "match", match)
		}

		if err := repository.Save(ctx, tx, match); err != nil {
			// TODO: a single match fail should not fail the entire sync.
			// We should have a "dead-letter" queue for failed matches so we reconcile each of them individually later.
			return fmt.Errorf("Failed to insert match: %w", err)
		}
	}

	return nil
}


