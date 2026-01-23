package football_org

import (
	"context"
	"log/slog"
	"provider/entity"
	"provider/repository"
)

// SaveMatches saves entity.Match objects to the database, filtering to only save finished matches.
// A single match failure does not stop processing of other matches.
func (p *Provider) SaveMatches(ctx context.Context, matches []entity.Match) {
	for _, match := range matches {
		// Only save matches that are in finished status
		if match.Status != entity.Finished {
			slog.Debug("Skipping match that is not in final status", "match_id", match.ProviderMatchID, "status", match.Status)
			continue
		}

		// As a match may be rescheduled, we need to delete the existing match in case it already exists.
		// This could happen if we reprocess old dates or if we already stored a match during the day
		if err := repository.DeleteByCanonicalID(ctx, match.CanonicalID, p.GetProviderEntity()); err != nil {
			slog.Error("Failed to delete match", "error", err, "match", match)
		}

		if err := repository.Save(ctx, match); err != nil {
			// A single match failure should not fail the entire sync.
			// Add to reconciliation queue for later reconciliation.
			if saveErr := repository.SaveToReconciliationQueue(ctx, match.ProviderMatchID, p.GetProviderEntity()); saveErr != nil {
				slog.Error("MANUAL INTERVENTION REQUIRED: Failed to insert match and failed to add to reconciliation queue",
					"provider_match_id", match.ProviderMatchID,
					"save_error", err,
					"reconciliation_error", saveErr)

				continue
			}

			slog.Warn("Match save failed, added to reconciliation queue",
				"provider_match_id", match.ProviderMatchID,
				"error", err)

			continue
		}
	}
}
