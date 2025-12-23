package services

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common/math"
)

func LoadChainId() (*math.HexOrDecimal256, error) {
	chainIdStr := os.Getenv("CHAIN_ID")
	if chainIdStr == "" {
		return nil, fmt.Errorf("CHAIN_ID environment variable is not set")
	}

	chainIdBig, ok := new(big.Int).SetString(chainIdStr, 10)
	if !ok {
		return nil, fmt.Errorf("unable to parse chain ID: %s", chainIdStr)
	}

	return (*math.HexOrDecimal256)(chainIdBig), nil
}
