package testutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
)

// ServerWithRequestCapture wraps a test server and captures the request URL and query parameters
type ServerWithRequestCapture struct {
	Server        *httptest.Server
	RequestURL    *url.URL
	mu            sync.Mutex
	RequestsCount uint
}

// ServerBuilder provides a fluent interface for building test servers
type ServerBuilder struct {
	statusCode   int
	responseBody string
	delay        time.Duration
}

func CreateServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		statusCode:   http.StatusOK,
		responseBody: "",
		delay:        0,
	}
}

func (b *ServerBuilder) WithStatusCode(statusCode int) *ServerBuilder {
	b.statusCode = statusCode
	return b
}

func (b *ServerBuilder) WithResponseBody(responseBody string) *ServerBuilder {
	b.responseBody = responseBody
	return b
}

func (b *ServerBuilder) WithDelay(delay time.Duration) *ServerBuilder {
	b.delay = delay
	return b
}

func (b *ServerBuilder) Build() *ServerWithRequestCapture {
	capture := &ServerWithRequestCapture{}

	capture.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.mu.Lock()
		capture.RequestURL = r.URL
		capture.RequestsCount += 1
		capture.mu.Unlock()

		if b.delay > 0 {
			time.Sleep(b.delay)
		}

		w.WriteHeader(b.statusCode)
		w.Write([]byte(b.responseBody))
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
