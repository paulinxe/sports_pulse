package main

import (
	"os"
	"relayer/testutil"
	"testing"
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
