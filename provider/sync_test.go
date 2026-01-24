package main

import (
	"testing"
)

func Test_Sync_returns_error_for_unknown_competition(t *testing.T) {
	err := Sync("football_org", "invalid_competition", systemClock{})
	if err == nil {
		t.Error("Expected error for unknown competition, but got nil")
	}

	expectedError := "Unknown competition: invalid_competition"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got: %s", expectedError, err.Error())
	}
}

func Test_Sync_returns_error_for_unknown_provider(t *testing.T) {
	err := Sync("invalid_provider", "la_liga", systemClock{})
	if err == nil {
		t.Error("Expected error for unknown provider, but got nil")
	}

	expectedError := "Unknown provider: invalid_provider"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got: %s", expectedError, err.Error())
	}
}
