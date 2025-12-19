package testutil

import (
	"bytes"
	"log/slog"
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
