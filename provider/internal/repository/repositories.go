package repository

import "database/sql"

type Repositories struct {
	Match          *MatchRepository
	SyncState      *SyncStateRepository
	Reconciliation *ReconciliationRepository
}

func InitRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Match:          NewMatchRepository(db),
		SyncState:      NewSyncStateRepository(db),
		Reconciliation: NewReconciliationRepository(db),
	}
}
