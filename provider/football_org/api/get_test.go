package api

import (
	"context"
	"net/http"
	"provider/testutil"
	"strings"
	"testing"
	"time"
)

func Test_we_can_handle_unauthorized_response(t *testing.T) {
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusForbidden).
		WithResponseBody("").
		Build()
	defer mockServer.Close()

	_, err := GetOne(context.Background(), "/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "403 Forbidden") {
		t.Errorf("Expected '403 Forbidden' in error message, but got: %s", errMsg)
	}
}

func Test_we_can_handle_too_many_requests_response(t *testing.T) {
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusTooManyRequests).
		WithResponseBody("").
		Build()
	defer mockServer.Close()

	_, err := GetOne(context.Background(), "/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "429 Too Many Requests") {
		t.Errorf("Expected '429 Too Many Requests' in error message, but got: %s", errMsg)
	}
}

func Test_we_can_handle_internal_server_error_response(t *testing.T) {
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusInternalServerError).
		WithResponseBody("").
		Build()
	defer mockServer.Close()

	_, err := GetOne(context.Background(), "/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "500 Internal Server Error") {
		t.Errorf("Expected '500 Internal Server Error' in error message, but got: %s", errMsg)
	}
}

func Test_we_can_handle_invalid_json_response(t *testing.T) {
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody("invalid json").
		Build()
	defer mockServer.Close()

	_, err := GetOne(context.Background(), "")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Failed to parse JSON response") {
		t.Errorf("Expected 'Failed to parse JSON response' in error message, but got: %s", errMsg)
	}
}

func Test_GetOne_handles_context_cancellation(t *testing.T) {
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(`{"id": 1}`).
		Build()
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := GetOne(ctx, "/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Request canceled") {
		t.Errorf("Expected 'Request canceled' in error message, but got: %s", errMsg)
	}
}

func Test_GetOne_handles_context_timeout(t *testing.T) {
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(`{"id": 1}`).
		WithDelay(2 * time.Second). // Delay longer than context timeout
		Build()
	defer mockServer.Close()

	// Create context with 500ms timeout - shorter than server delay
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := GetOne(ctx, "/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Context timeout") {
		t.Errorf("Expected 'Context timeout' in error message, but got: %s", errMsg)
	}
}
