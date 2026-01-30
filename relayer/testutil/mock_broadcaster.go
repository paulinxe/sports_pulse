package testutil

import "context"

type MockBroadcaster struct {
	TimesCalled int
}

func (m *MockBroadcaster) Broadcast(ctx context.Context, calldata []byte) error {
	m.TimesCalled++
	return nil
}