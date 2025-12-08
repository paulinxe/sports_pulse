package testutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

func ExpectNumberOfRequests(t *testing.T, server *ServerWithRequestCapture, numberOfRequests uint) {
    if server.RequestsCount != numberOfRequests {
        t.Errorf("Expected %d requests, but got %d", numberOfRequests, server.RequestsCount)
    }
}