package repository

import "database/sql"

type Repositories struct {
	Match          *MatchRepository
	SyncState      *SyncStateRepository
	Reconciliation *ReconciliationRepository
}

func InitRepositories(db *sql.DB) (*Repositories, error) {
	matchRepo, err := NewMatchRepository(db)
	if err != nil {
		return nil, err
	}

	syncStateRepo, err := NewSyncStateRepository(db)
	if err != nil {
		return nil, err
	}

	reconciliationRepo, err := NewReconciliationRepository(db)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		Match:          matchRepo,
		SyncState:      syncStateRepo,
		Reconciliation: reconciliationRepo,
	}, nil
}
