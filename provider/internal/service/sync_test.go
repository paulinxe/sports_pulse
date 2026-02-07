package service

import (
	"testing"

	"provider/testutil"
)

func Test_Sync_returns_error_for_unknown_competition(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	err := Sync(repositories, "football_org", "invalid_competition", SystemClock{})
	if err == nil {
		t.Error("Expected error for unknown competition, but got nil")
	}

	expectedError := "unknown competition: invalid_competition"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got: %s", expectedError, err.Error())
	}
}

func Test_Sync_returns_error_for_unknown_provider(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	err := Sync(repositories, "invalid_provider", "la_liga", SystemClock{})
	if err == nil {
		t.Error("Expected error for unknown provider, but got nil")
	}

	expectedError := "unknown provider: invalid_provider"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got: %s", expectedError, err.Error())
	}
}
