package apifootball

import (
	"context"
	"log/slog"
	"provider/internal/entity"
)

// SaveMatches saves entity.Match objects to the database, filtering to only save finished matches.
// A single match failure does not stop processing; failed saves are enqueued for reconciliation.
// TODO: this is the same as the football_org/save.go file, we should refactor this to a common function.
func (p *Provider) SaveMatches(ctx context.Context, matches []entity.Match) {
	for _, match := range matches {
		if match.Status != entity.Finished {
			slog.Debug("Skipping match that is not in final status", "match_id", match.ProviderMatchID, "status", match.Status)
			continue
		}

		if err := p.matchRepository.Save(ctx, match); err != nil {
			if saveErr := p.reconciliationRepository.SaveToReconciliationQueue(ctx, match.ProviderMatchID, p.GetProviderEntity()); saveErr != nil {
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
