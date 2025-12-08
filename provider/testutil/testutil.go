package testutil

import (
    "bytes"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "net/url"
    "os"
    "provider/db"
    "sync"
    "testing"
)

// ServerWithRequestCapture wraps a test server and captures the request URL and query parameters
type ServerWithRequestCapture struct {
    Server     *httptest.Server
    RequestURL *url.URL
    mu         sync.Mutex
	RequestsCount uint
}

func CreateServer(statusCode int, responseBody string) *ServerWithRequestCapture {
    capture := &ServerWithRequestCapture{}

    capture.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        capture.mu.Lock()
        capture.RequestURL = r.URL
		capture.RequestsCount += 1
        capture.mu.Unlock()

        w.WriteHeader(statusCode)
        w.Write([]byte(responseBody))
    }))

    // Set up environment variable to point to mock server
    os.Setenv("FOOTBALL_ORG_API_ENDPOINT", capture.Server.URL)

    return capture
}

// Close closes the underlying test server
func (s *ServerWithRequestCapture) Close() {
    s.Server.Close()
}

// GetQueryParam returns the value of a query parameter from the captured request
func (s *ServerWithRequestCapture) GetQueryParam(key string) string {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.RequestURL == nil {
        return ""
    }
    return s.RequestURL.Query().Get(key)
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

func InitDatabase(t *testing.T) {
    err := db.Init()
    if err != nil {
        t.Fatalf("Failed to initialize database: %v", err)
    }

    // Verify database connection is ready
    if db.DB == nil {
        t.Fatalf("Database connection is nil after initialization")
    }

    // Clean up before the test
    // TODO: we need a better way to clean up the database.
    _, _ = db.DB.Exec("DELETE FROM matches")
}

func CloseDatabase() {
    if err := db.Close(); err != nil {
        slog.Error("Failed to close database", "error", err)
        os.Exit(1)
    }
}

func MatchExists(t *testing.T, matchID string) bool {
    var count int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE id = $1", matchID).Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query database: %v", err)
    }
    return count > 0
}
