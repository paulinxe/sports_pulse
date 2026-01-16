package testutil

import (
	"bytes"
	"log/slog"
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
