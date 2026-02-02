package testutil

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MockChainClient is a services.ChainClient implementation for tests.
// Set Nonce, TipCap, FeeCap, Receipt, SendErr, EstimatedGas, or EstimateGasErr to control behavior.
type MockChainClient struct {
	Nonce          uint64
	TipCap         *big.Int
	FeeCap         *big.Int
	Receipt        *types.Receipt
	SendErr        error
	EstimatedGas   uint64
	EstimateGasErr error
	TimesCalled    int
	LastSentTx     *types.Transaction // last transaction passed to SendTransaction
}

func (mock *MockChainClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return mock.Nonce, nil
}

func (mock *MockChainClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	if mock.TipCap != nil {
		return mock.TipCap, nil
	}

	return big.NewInt(1), nil
}

func (mock *MockChainClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if mock.FeeCap != nil {
		return mock.FeeCap, nil
	}

	return big.NewInt(2), nil
}

func (mock *MockChainClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if mock.EstimateGasErr != nil {
		return 0, mock.EstimateGasErr
	}

	if mock.EstimatedGas != 0 {
		return mock.EstimatedGas, nil
	}
	// Default success value when neither is set. Fallback logic lives in broadcaster, not the client.
	return 100_000, nil
}

func (mock *MockChainClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	mock.TimesCalled++
	mock.LastSentTx = tx
	return mock.SendErr
}

func (mock *MockChainClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if mock.Receipt != nil {
		return mock.Receipt, nil
	}

	return nil, ethereum.NotFound
}


// DataErrorForTests implements rpc.DataError so tests can simulate MatchAlreadySubmitted.
// Use as SendErr to make isMatchAlreadySubmitted return true when Data is the selector hex (e.g. "0x...").
type DataErrorForTests struct {
	Msg  string
	Data interface{}
}

func (dataError *DataErrorForTests) Error() string        { return dataError.Msg }
func (dataError *DataErrorForTests) ErrorData() interface{} { return dataError.Data }
