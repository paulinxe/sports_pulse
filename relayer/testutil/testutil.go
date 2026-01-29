package testutil

import (
	"testing"
)

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}
