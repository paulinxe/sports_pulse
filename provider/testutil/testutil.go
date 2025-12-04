package testutil

import (
    "bytes"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "os"
    "provider/db"
)

func CreateServer(statusCode int, responseBody string) *httptest.Server {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(statusCode)
        w.Write([]byte(responseBody))
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

func InitDatabase() error {
    return db.Init()
}

func CloseDatabase() {
    if err := db.Close(); err != nil {
        slog.Error("Failed to close database", "error", err)
        os.Exit(1)
    }
}
