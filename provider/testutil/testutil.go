package testutil

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func GetLogger() *bytes.Buffer {
	var logBuf bytes.Buffer

	slog.SetDefault(
		slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	)

	return &logBuf
}

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

func AssertMessageGotLogged(t *testing.T, logBuf *bytes.Buffer, message string) {
	if !strings.Contains(logBuf.String(), message) {
		t.Errorf("Expected message '%s' to be logged, but got: %s", message, logBuf.String())
	}
}