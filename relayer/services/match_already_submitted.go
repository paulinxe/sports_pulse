package services

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/rpc"
)

func isMatchAlreadySubmitted(err error, contractABI abi.ABI) bool {
	var dataErr rpc.DataError
	if !errors.As(err, &dataErr) {
		return false
	}

	data := dataErr.ErrorData()
	if data == nil {
		return false
	}

	errorSelector, ok := data.(string)
	if !ok {
		return false
	}

	errorSelector = strings.TrimPrefix(errorSelector, "0x")
	errorSelectorBytes, decodeErr := hex.DecodeString(errorSelector)
	if decodeErr != nil || len(errorSelectorBytes) < 4 {
		return false
	}

	matchSubmittedError, ok := contractABI.Errors["MatchAlreadySubmitted"]
	if !ok {
		return false
	}

	return string(errorSelectorBytes[:4]) == string(matchSubmittedError.ID[:4])
}
