package api

import (
	"net/http"
	"provider/testutil"
	"strings"
	"testing"
)

func Test_we_can_handle_unauthorized_response(t *testing.T) {
	mockServer := testutil.CreateServer(http.StatusForbidden, "")
	defer mockServer.Close()

	_, err := GetOne("/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "403 Forbidden") {
		t.Errorf("Expected '403 Forbidden' in error message, but got: %s", errMsg)
	}
}

func Test_we_can_handle_too_many_requests_response(t *testing.T) {
	mockServer := testutil.CreateServer(http.StatusTooManyRequests, "")
	defer mockServer.Close()

	_, err := GetOne("/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "429 Too Many Requests") {
		t.Errorf("Expected '429 Too Many Requests' in error message, but got: %s", errMsg)
	}
}

func Test_we_can_handle_internal_server_error_response(t *testing.T) {
	mockServer := testutil.CreateServer(http.StatusInternalServerError, "")
	defer mockServer.Close()

	_, err := GetOne("/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "500 Internal Server Error") {
		t.Errorf("Expected '500 Internal Server Error' in error message, but got: %s", errMsg)
	}
}

func Test_we_can_handle_invalid_json_response(t *testing.T) {
	mockServer := testutil.CreateServer(http.StatusOK, "invalid json")
	defer mockServer.Close()

	_, err := GetOne("")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Failed to parse JSON response") {
		t.Errorf("Expected 'Failed to parse JSON response' in error message, but got: %s", errMsg)
	}
}

func Test_GetList_we_can_handle_unauthorized_response(t *testing.T) {
	mockServer := testutil.CreateServer(http.StatusForbidden, "")
	defer mockServer.Close()

	_, err := GetList("/")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "403 Forbidden") {
		t.Errorf("Expected '403 Forbidden' in error message, but got: %s", errMsg)
	}
}

func Test_GetList_we_can_handle_invalid_json_response(t *testing.T) {
	mockServer := testutil.CreateServer(http.StatusOK, "invalid json")
	defer mockServer.Close()

	_, err := GetList("")
	if err == nil {
		t.Error("Expected error but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, errMsg) {
		t.Errorf("Expected '%s' in error message, but got: %s", errMsg, errMsg)
	}
}