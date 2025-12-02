package testutil

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
)

func CreateServer(expectedStatusCode int) *httptest.Server {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatusCode)
		w.Write([]byte(http.StatusText(expectedStatusCode)))
	}))

	// Set up environment variable to point to mock server
	os.Setenv("FOOTBALL_ORG_API_ENDPOINT", mockServer.URL)

	return mockServer
}

func GetLogger() *bytes.Buffer {
	var logBuf bytes.Buffer

	slog.SetDefault(
		slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	)

	return &logBuf
}

