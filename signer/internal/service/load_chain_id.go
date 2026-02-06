package service

import (
	"fmt"
	"os"
	"strconv"
)

func LoadChainId() (uint, error) {
	chainIdStr := os.Getenv("CHAIN_ID")
	if chainIdStr == "" {
		return 0, fmt.Errorf("CHAIN_ID environment variable is not set")
	}

	chainId, err := strconv.ParseUint(chainIdStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse chain ID: %s: %w", chainIdStr, err)
	}

	return uint(chainId), nil
}
