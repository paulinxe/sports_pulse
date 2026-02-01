package main

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"math/big"
	"relayer/config"
	"relayer/entity"
	"relayer/services"
	"relayer/testutil"
)

func Test_we_get_an_error_when_database_is_not_initialized(t *testing.T) {
	dbUser := os.Getenv("DB_USER")
	defer func() { _ = os.Setenv("DB_USER", dbUser) }()
	_ = os.Unsetenv("DB_USER")

	cfg, mockClient := buildTestBroadcasterConfig(t)
	errorCode := Run(mockClient, cfg)

	if errorCode != int(DB_INIT_ERROR) {
		t.Errorf("expected error code %d, got %d", DB_INIT_ERROR, errorCode)
	}
}

func Test_we_handle_no_matches_to_broadcast(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	cfg, mockClient := buildTestBroadcasterConfig(t)
	errorCode := Run(mockClient, cfg)

	if errorCode != int(SUCCESS) {
		t.Errorf("expected error code %d, got %d", SUCCESS, errorCode)
	}

	if mockClient.TimesCalled != 0 {
		t.Errorf("expected 0 times called, got %d", mockClient.TimesCalled)
	}
}

func Test_we_can_broadcast_matches(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	sigHex := "deadbeef"
	id1 := uuid.New()
	id2 := uuid.New()
	testutil.InsertSignedMatch(t, id1, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1, 10, 20, 2, 1, start, sigHex)
	testutil.InsertSignedMatch(t, id2, "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 1, 30, 40, 0, 0, start, sigHex)

	cfg, mockClient := buildTestBroadcasterConfig(t)
	mockClient.Receipt = &types.Receipt{Status: types.ReceiptStatusSuccessful}
	errorCode := Run(mockClient, cfg)

	if errorCode != int(SUCCESS) {
		t.Errorf("expected error code %d, got %d", SUCCESS, errorCode)
	}
	assertMatchStatus(t, id1, entity.BROADCASTED_STATUS)
	assertMatchStatus(t, id2, entity.BROADCASTED_STATUS)
}

func Test_we_update_status_to_broadcasted_when_match_already_submitted(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	testutil.InsertSignedMatch(t, id, "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 1, 30, 40, 0, 0, start, "deadbeef")

	cfg, mockClient := buildTestBroadcasterConfig(t)
	matchErr := cfg.ContractABI.Errors["MatchAlreadySubmitted"]
	selectorHex := "0x" + hex.EncodeToString(matchErr.ID[:4])
	mockClient.SendErr = &testutil.DataErrorForTests{Msg: "MatchAlreadySubmitted", Data: selectorHex}
	errorCode := Run(mockClient, cfg)

	if errorCode != int(SUCCESS) {
		t.Errorf("expected SUCCESS when match already submitted, got %d", errorCode)
	}

	assertMatchStatus(t, id, entity.BROADCASTED_STATUS)
}

func Test_match_remains_signed_when_client_fails_when_sending_transaction(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	testutil.InsertSignedMatch(t, id, "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", 1, 10, 20, 1, 1, start, "deadbeef")

	cfg, mockClient := buildTestBroadcasterConfig(t)
	mockClient.SendErr = errors.New("send failed")
	errorCode := Run(mockClient, cfg)

	if errorCode != int(BROADCAST_FAILURE) {
		t.Errorf("expected BROADCAST_FAILURE, got %d", errorCode)
	}

	assertMatchStatus(t, id, entity.SIGNED_STATUS)
}

func Test_match_remains_signed_when_transaction_reverted(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	testutil.InsertSignedMatch(t, id, "0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", 1, 10, 20, 0, 0, start, "deadbeef")

	cfg, mockClient := buildTestBroadcasterConfig(t)
	mockClient.Receipt = &types.Receipt{Status: types.ReceiptStatusFailed}
	errorCode := Run(mockClient, cfg)

	if errorCode != int(BROADCAST_FAILURE) {
		t.Errorf("expected BROADCAST_FAILURE when transaction reverted, got %d", errorCode)
	}

	assertMatchStatus(t, id, entity.SIGNED_STATUS)
}

// retryReceiptClient returns ethereum.NotFound for the first receiptCallsBeforeSuccess
// TransactionReceipt calls, then returns the embedded mock's Receipt. Used to test waitForReceipt retry behavior.
type retryReceiptClient struct {
	*testutil.MockChainClient
	receiptCallsBeforeSuccess int
	receiptCalls              int
}

func (client *retryReceiptClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	client.receiptCalls++
	if client.receiptCalls <= client.receiptCallsBeforeSuccess {
		return nil, ethereum.NotFound
	}

	return client.MockChainClient.TransactionReceipt(ctx, txHash)
}

func Test_broadcast_succeeds_after_waitForReceipt_retries_on_NotFound(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	testutil.InsertSignedMatch(t, id, "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", 1, 10, 20, 2, 1, start, "deadbeef")

	cfg, mockClient := buildTestBroadcasterConfig(t)
	mockClient.Receipt = &types.Receipt{Status: types.ReceiptStatusSuccessful}
	client := &retryReceiptClient{
		MockChainClient:           mockClient,
		receiptCallsBeforeSuccess: 2,
	}
	errorCode := Run(client, cfg)

	if errorCode != int(SUCCESS) {
		t.Errorf("expected SUCCESS after retries, got %d", errorCode)
	}

	assertMatchStatus(t, id, entity.BROADCASTED_STATUS)
	if client.receiptCalls != 3 {
		t.Errorf("expected 3 TransactionReceipt calls, got %d", client.receiptCalls)
	}
}

func assertMatchStatus(t *testing.T, matchID uuid.UUID, expectedStatus int) {
	t.Helper()
	var actualStatus int

	err := config.DB.QueryRow("SELECT status FROM matches WHERE id = $1", matchID).Scan(&actualStatus)
	testutil.AssertNoError(t, err)

	if actualStatus != expectedStatus {
		t.Errorf("match %s: expected status %d, got %d", matchID, expectedStatus, actualStatus)
	}
}

func buildTestBroadcasterConfig(t *testing.T) (services.BroadcasterConfig, *testutil.MockChainClient) {
	t.Helper()
	key, err := crypto.GenerateKey()
	testutil.AssertNoError(t, err)
	contractABI, err := abi.JSON(strings.NewReader(services.MATCH_REGISTRY_SUBMIT_MATCH_ABI))
	testutil.AssertNoError(t, err)
	cfg := services.BroadcasterConfig{
		RPCURL:          "",
		ContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
		PrivateKey:      key,
		ChainID:         big.NewInt(31337),
		ContractABI:     contractABI,
	}

	mockClient := &testutil.MockChainClient{
		Nonce:  0,
		TipCap: big.NewInt(1),
		FeeCap: big.NewInt(2),
	}

	return cfg, mockClient
}
