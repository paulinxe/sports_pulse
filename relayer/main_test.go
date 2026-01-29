package main

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"relayer/config"
	"relayer/testutil"
)

func Test_we_get_an_error_when_database_is_not_initialized(t *testing.T) {
	dbUser := os.Getenv("DB_USER")
	defer func() { _ = os.Setenv("DB_USER", dbUser) }()
	_ = os.Unsetenv("DB_USER")

	broadcaster := &testutil.MockBroadcaster{}
	errorCode := Run(broadcaster)

	if errorCode != int(DB_INIT_ERROR) {
		t.Errorf("expected error code %d, got %d", DB_INIT_ERROR, errorCode)
	}
}

func Test_we_handle_no_matches_to_broadcast(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	broadcaster := &testutil.MockBroadcaster{}
	errorCode := Run(broadcaster)

	if errorCode != int(SUCCESS) {
		t.Errorf("expected error code %d, got %d", SUCCESS, errorCode)
	}

	if broadcaster.TimesCalled != 0 {
		t.Errorf("broadcaster expected 0 times called, got %d", broadcaster.TimesCalled)
	}
}

func Test_we_can_broadcast_matches(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	// Insert two signed matches so Run will pick them up
	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	sigHex := "deadbeef"
	id1 := uuid.New()
	id2 := uuid.New()
	testutil.InsertSignedMatch(t, id1, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1, 10, 20, 2, 1, start, sigHex)
	testutil.InsertSignedMatch(t, id2, "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 1, 30, 40, 0, 0, start, sigHex)

	broadcaster := &testutil.MockBroadcaster{}
	errorCode := Run(broadcaster)

	if errorCode != int(SUCCESS) {
		t.Errorf("expected error code %d, got %d", SUCCESS, errorCode)
	}

	if broadcaster.TimesCalled != 2 {
		t.Errorf("broadcaster expected 2 times called, got %d", broadcaster.TimesCalled)
	}

	assertMatchStatus(t, id1, 5) // TODO: use the constant
	assertMatchStatus(t, id2, 5) // TODO: use the constant
}

func assertMatchStatus(t *testing.T, matchID uuid.UUID, expectedStatus int) {
	t.Helper()
	var actualStatus int

	err := config.DB.QueryRow("SELECT status FROM matches WHERE id = $1", matchID).Scan(&actualStatus)
	testutil.AssertNoError(t, err)

	if actualStatus != expectedStatus {
		t.Errorf("match %s: expected status %d, got %d", matchID, expectedStatus, actualStatus)
	}
}
