package multiplexer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type MockRelayer struct {
	Messages []any
	mu       sync.Mutex
}

const ownerID = "owner1"

func (m *MockRelayer) SendMsg(_ context.Context, msg any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, msg)

	return nil
}

func TestNewRelayers(t *testing.T) {
	relayers := NewRelayers()
	assert.NotNil(t, relayers)
	assert.Empty(t, relayers.relayers)
}

func TestNewMultiplexer(t *testing.T) {
	mux := New()
	assert.NotNil(t, mux)
	assert.NotNil(t, mux.relayers)
}

func TestRelayers_Acquire(t *testing.T) {
	relayers := NewRelayers()
	relayer := &MockRelayer{}
	relayers.relayers[ownerID] = relayer

	acquiredRelayer, err := relayers.acquire(ownerID)
	require.NoError(t, err)
	assert.Equal(t, relayer, acquiredRelayer)
	assert.Empty(t, relayers.relayers)
}

func TestRelayers_Acquire_NotFound(t *testing.T) {
	relayers := NewRelayers()
	ownerID := "nonexistent"

	acquiredRelayer, err := relayers.acquire(ownerID)
	require.Error(t, err)
	assert.Nil(t, acquiredRelayer)
}

func TestRelayers_AcquireRandom(t *testing.T) {
	relayers := NewRelayers()
	relayer1 := &MockRelayer{}
	relayer2 := &MockRelayer{}
	relayers.relayers["owner1"] = relayer1
	relayers.relayers["owner2"] = relayer2

	ownerID, acquiredRelayer, err := relayers.acquireRandom("owner1")
	require.NoError(t, err)
	assert.Equal(t, "owner2", ownerID)
	assert.Equal(t, relayer2, acquiredRelayer)
}

func TestRelayers_AcquireRandom_NoRelayers(t *testing.T) {
	relayers := NewRelayers()
	ownerID, acquiredRelayer, err := relayers.acquireRandom("owner1")
	require.Error(t, err)
	assert.Equal(t, noOpID, ownerID)
	assert.IsType(t, NoopRelayer{}, acquiredRelayer)
}

func TestRelayers_Release(t *testing.T) {
	relayers := NewRelayers()
	relayer := &MockRelayer{}

	relayers.release(ownerID, relayer)
	assert.Equal(t, relayer, relayers.relayers[ownerID])
}

func TestRelayers_Register(t *testing.T) {
	relayers := NewRelayers()
	relayer := &MockRelayer{}

	relayers.register(ownerID, relayer)
	assert.Equal(t, relayer, relayers.relayers[ownerID])
}

func TestMultiplexer_AcquireRelayer(t *testing.T) {
	mux := New()
	relayer := &MockRelayer{}
	mux.Register(context.Background(), ownerID, relayer)

	acquiredRelayer, err := mux.AcquireRelayer(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Equal(t, relayer, acquiredRelayer)
}

func TestMultiplexer_AcquireRandomRelayer(t *testing.T) {
	mux := New()
	relayer1 := &MockRelayer{}
	relayer2 := &MockRelayer{}

	mux.Register(context.Background(), "owner1", relayer1)
	mux.Register(context.Background(), "owner2", relayer2)

	ownerID, acquiredRelayer, err := mux.AcquireRandomRelayer(context.Background(), "owner1")
	require.NoError(t, err)
	assert.Equal(t, "owner2", ownerID)
	assert.Equal(t, relayer2, acquiredRelayer)
}

func TestMultiplexer_ReleaseRelayer(t *testing.T) {
	mux := New()
	relayer := &MockRelayer{}
	mux.Register(context.Background(), ownerID, relayer)

	mux.ReleaseRelayer(context.Background(), ownerID, relayer)
	assert.Equal(t, relayer, mux.relayers.relayers[ownerID])
}

func TestMultiplexer_Register(t *testing.T) {
	mux := New()
	relayer := &MockRelayer{}

	mux.Register(context.Background(), ownerID, relayer)
	assert.Equal(t, relayer, mux.relayers.relayers[ownerID])
}
