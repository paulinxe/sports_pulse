package main

import (
	"net/http"
	"strings"
	"testing"
)

func Test_we_can_handle_unauthorized_response(t *testing.T) {
	logger := getLogger()
    mockServer := createServer(http.StatusForbidden)
    defer mockServer.Close()

	args := []string{"provider", "football_org"}

	exitCode := run(args)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, but got %d", exitCode)
	}

    outputStr := logger.String()
	if !strings.Contains(outputStr, "403 Forbidden") {
		t.Errorf("Expected '403 Forbidden' in output, but got: %s", outputStr)
	}
}

func Test_we_can_handle_too_many_requests_response(t *testing.T) {
	logger := getLogger()
    mockServer := createServer(http.StatusTooManyRequests)
    defer mockServer.Close()

    args := []string{"provider", "football_org"}
    exitCode := run(args)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, but got %d", exitCode)
	}

    outputStr := logger.String()
	if !strings.Contains(outputStr, "429 Too Many Requests") {
		t.Errorf("Expected '429 Too Many Requests' in output, but got: %s", outputStr)
	}
}

func Test_we_can_handle_internal_server_error_response(t *testing.T) {
	logger := getLogger()
    mockServer := createServer(http.StatusInternalServerError)
    defer mockServer.Close()

    args := []string{"provider", "football_org"}
    exitCode := run(args)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, but got %d", exitCode)
	}

    outputStr := logger.String()
	if !strings.Contains(outputStr, "500 Internal Server Error") {
		t.Errorf("Expected '500 Internal Server Error' in output, but got: %s", outputStr)
	}
}