package football_org

import (
	"context"
	"strings"
	"testing"
	"time"

	"provider/db"
	"provider/entity"
	"provider/testutil"
)

// This test is not part of sync_football_org_test.go because for this specific scenario, its easier to "unit" test it here.
func Test_SaveMatches_continues_when_save_fails_but_reconciliation_succeeds(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	provider := &Provider{}
	ctx := context.Background()

	// Create a match
	startTime, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
	match, err := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		"test_match_123",
		entity.Girona,
		entity.RayoVallecano,
		1,
		2,
		entity.LaLiga,
		entity.Finished,
	)
	testutil.AssertNoError(t, err)

	// Insert the match directly into the database with SQL to bypass normal flow
	// Use a different canonical_id so DeleteByCanonicalID won't find it
	// This ensures the match exists with this ID, causing a primary key violation
	_, err = db.DB.Exec(`
		INSERT INTO matches (
			id, canonical_id, home_team_id, away_team_id, start, "end", status,
			home_team_score, away_team_score, provider_match_id, competition_id, provider
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		match.ID,
		"different_canonical_id_to_avoid_delete", // Different canonical_id so DeleteByCanonicalID won't find it
		match.HomeTeamID,
		match.AwayTeamID,
		match.Start,
		match.End,
		match.Status,
		match.HomeTeamScore,
		match.AwayTeamScore,
		"different_provider_match_id",
		match.CompetitionID,
		match.Provider,
	)
	testutil.AssertNoError(t, err)

	// Now try to save a match with the same ID but different canonical_id
	// DeleteByCanonicalID won't find it (different canonical_id), but Save will fail due to primary key violation
	match2, err := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		"test_match_123",
		entity.Girona,
		entity.RayoVallecano,
		1,
		2,
		entity.LaLiga,
		entity.Finished,
	)
	testutil.AssertNoError(t, err)
	// Use the same ID to cause primary key violation
	match2.ID = match.ID
	
	provider.SaveMatches(ctx, []entity.Match{match2})
	
	// Verify match was added to reconciliation queue
	if !testutil.ReconciliationEntryExists(t, "test_match_123", int(entity.FootballOrg)) {
		t.Errorf("Expected match to be in reconciliation queue, but it is not")
	}
	
	// Verify warning was logged
	outputStr := logger.String()
	if !strings.Contains(outputStr, "Match save failed, added to reconciliation queue") {
		t.Errorf("Expected warning log 'Match save failed, added to reconciliation queue', but got: %s", outputStr)
	}
}