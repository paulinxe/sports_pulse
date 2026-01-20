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

	_, err := GetList(context.Background(), 1, time.Now(), time.Now())
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

	_, err := GetList(context.Background(), 1, time.Now(), time.Now())
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

	_, err := GetList(context.Background(), 1, time.Now(), time.Now())
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

	_, err := GetList(context.Background(), 1, time.Now(), time.Now())
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Failed to parse JSON response") {
		t.Errorf("Expected 'Failed to parse JSON response' in error message, but got: %s", errMsg)
	}
}
