package main

import (
	"fmt"
	"provider/football_org"
	"strings"
	"time"
)

const RECONCILE_CONTEXT_TIMEOUT = 15 * time.Second

func Reconcile(provider string) error {
	switch strings.ToLower(provider) {
	case "football_org":
		if err := football_org.Reconcile(RECONCILE_CONTEXT_TIMEOUT); err != nil {
			return fmt.Errorf("Failed to reconcile Football Data API: %w", err)
		}
	default:
		return fmt.Errorf("Unknown provider: %s", provider)
	}

	return nil
}
