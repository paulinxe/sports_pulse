package testutil

import (
	"context"
	"relayer/entity"
)

type MockBroadcaster struct {
	TimesCalled int
}

func (m *MockBroadcaster) Broadcast(ctx context.Context, match entity.Match) error {
	m.TimesCalled++
	return nil
}