package save

import (
	"context"
	"log/slog"
	"provider/internal/entity"
	"provider/internal/repository"
)

// Saver saves matches to the database, filtering to finished only and enqueueing failures for reconciliation.
type Saver struct {
	matchRepo          *repository.MatchRepository
	reconciliationRepo *repository.ReconciliationRepository
}

func NewSaver(matchRepo *repository.MatchRepository, reconciliationRepo *repository.ReconciliationRepository) *Saver {
	return &Saver{
		matchRepo:          matchRepo,
		reconciliationRepo: reconciliationRepo,
	}
}

// SaveMatches saves entity.Match objects to the database, filtering to only save finished matches.
// A single match failure does not stop processing; failed saves are enqueued for reconciliation.
func (s *Saver) SaveMatches(ctx context.Context, matches []entity.Match, provider entity.Provider) {
	for _, match := range matches {
		if match.Status != entity.Finished {
			slog.Debug("Skipping match that is not in final status", "match_id", match.ProviderMatchID, "status", match.Status)
			continue
		}

		if err := s.matchRepo.Save(ctx, match); err != nil {
			if saveErr := s.reconciliationRepo.SaveToReconciliationQueue(ctx, match.ProviderMatchID, provider); saveErr != nil {
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
