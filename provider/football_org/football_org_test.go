package football_org

import (
	"net/http"
	"provider/testutil"
	"strings"
	"testing"
)

func Test_we_can_handle_unauthorized_response(t *testing.T) {
	logger := testutil.GetLogger()
	mockServer := testutil.CreateServer(http.StatusForbidden)
	defer mockServer.Close()

	err := Sync()
	if err == nil {
		t.Error("Expected error but got nil")
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "403 Forbidden") {
		t.Errorf("Expected '403 Forbidden' in output, but got: %s", outputStr)
	}
}

func Test_we_can_handle_too_many_requests_response(t *testing.T) {
	logger := testutil.GetLogger()
	mockServer := testutil.CreateServer(http.StatusTooManyRequests)
	defer mockServer.Close()

	err := Sync()
	if err == nil {
		t.Error("Expected error but got nil")
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "429 Too Many Requests") {
		t.Errorf("Expected '429 Too Many Requests' in output, but got: %s", outputStr)
	}
}

func Test_we_can_handle_internal_server_error_response(t *testing.T) {
	logger := testutil.GetLogger()
	mockServer := testutil.CreateServer(http.StatusInternalServerError)
	defer mockServer.Close()

	err := Sync()
	if err == nil {
		t.Error("Expected error but got nil")
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "500 Internal Server Error") {
		t.Errorf("Expected '500 Internal Server Error' in output, but got: %s", outputStr)
	}
}
