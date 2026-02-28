package apifootball

import (
	"provider/internal/entity"
	"testing"
)

// I normally prefer functional tests instead of unit tests.
// However, for this specific scenario of validation, given that logic is shared in sync.go and reconcile.go,
// it makes more sense to test it here.
func Test_we_get_an_error_when_missing_ids_for_a_match(t *testing.T) {
	competition := entity.Championship
	validMatch := apifootballEvent{
		MatchID:            "123",
		MatchHometeamID:    "3432",
		MatchAwayteamID:    "3096",
		LeagueID:           LeagueIDChampionship,
		MatchDate:          "2025-01-15",
		MatchTime:          "20:00",
		MatchStatus:        "Finished",
		MatchHometeamScore: "1",
		MatchAwayteamScore: "0",
	}

	t.Run("missing match ID", func(t *testing.T) {
		match := validMatch
		match.MatchID = ""
		_, err := eventToEntityMatch(match, competition)
		if err == nil {
			t.Fatal("expected error for missing match ID")
		}

		if err.Error() != "missing home or away team ID or match ID" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing home team ID", func(t *testing.T) {
		match := validMatch
		match.MatchHometeamID = ""
		_, err := eventToEntityMatch(match, competition)
		if err == nil {
			t.Fatal("expected error for missing home team ID")
		}

		if err.Error() != "missing home or away team ID or match ID" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing away team ID", func(t *testing.T) {
		match := validMatch
		match.MatchAwayteamID = ""
		_, err := eventToEntityMatch(match, competition)
		if err == nil {
			t.Fatal("expected error for missing away team ID")
		}

		if err.Error() != "missing home or away team ID or match ID" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func Test_we_get_an_error_when_missing_scores_for_finished_match(t *testing.T) {
	competition := entity.Championship
	validMatch := apifootballEvent{
		MatchID:            "124",
		MatchHometeamID:    "3432",
		MatchAwayteamID:    "3096",
		LeagueID:           LeagueIDChampionship,
		MatchDate:          "2025-01-15",
		MatchTime:          "20:00",
		MatchStatus:        "Finished",
		MatchHometeamScore: "1",
		MatchAwayteamScore: "0",
	}

	t.Run("missing home score", func(t *testing.T) {
		match := validMatch
		match.MatchHometeamScore = ""
		_, err := eventToEntityMatch(match, competition)
		if err == nil {
			t.Fatal("expected error for missing home score on finished match")
		}

		if err.Error() != "missing home or away team score for finished match" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing away score", func(t *testing.T) {
		match := validMatch
		match.MatchAwayteamScore = ""
		_, err := eventToEntityMatch(match, competition)
		if err == nil {
			t.Fatal("expected error for missing away score on finished match")
		}

		if err.Error() != "missing home or away team score for finished match" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
