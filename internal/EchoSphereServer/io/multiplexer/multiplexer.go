// Package multiplexer provides a mechanism to manage and route messages through different relayers.
package multiplexer

import (
	"context"
	"fmt"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
	"github.com/samber/lo"
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

const noOpID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

// NoopRelayer is a relayer that performs no operations.
type NoopRelayer struct{}

// SendMsg implements the core.Messager interface but performs no action.
func (n NoopRelayer) SendMsg(context.Context, any) error {
	return nil
}

// Relayers manages a collection of relayers with thread-safe access.
type Relayers struct {
	mu       sync.Mutex
	relayers map[string]core.Messager
}

// NewRelayers creates a new instance of Relayers.
func NewRelayers() *Relayers {
	return &Relayers{
		relayers: make(map[string]core.Messager),
	}
}

// Multiplexer manages relayers and routes messages between them.
type Multiplexer struct {
	relayers *Relayers
}

// New creates a new instance of Multiplexer.
func New() *Multiplexer {
	return &Multiplexer{
		relayers: NewRelayers(),
	}
}

// AcquireRelayer acquires a relayer associated with the given ownerID, removing it from the collection.
func (r *Relayers) acquire(ownerID string) (core.Messager, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	relayer, ok := r.relayers[ownerID]
	if !ok {
		return nil, fmt.Errorf("%w: relayer not found", core.ErrFailedToGetRelayer)
	}

	delete(r.relayers, ownerID)

	return relayer, nil
}

// AcquireRandom acquires a random relayer, excluding the specified relayer, from the collection.
func (r *Relayers) acquireRandom(excludeRelayer string) (string, core.Messager, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.relayers) <= 1 {
		return noOpID, NoopRelayer{}, fmt.Errorf("%w: %s", core.ErrFailedToGetRelayer, "no relayers available")
	}

	keys := lo.Filter(
		lo.Keys(r.relayers),
		func(ownerID string, _ int) bool {
			return ownerID != excludeRelayer
		},
	)

	rand.Seed(uint64(time.Now().UnixNano()))
	randomKey := keys[rand.Intn(len(keys))]

	relayer := r.relayers[randomKey]

	delete(r.relayers, randomKey)

	return randomKey, relayer, nil
}

// Release adds a relayer back to the collection under the given ownerID.
func (r *Relayers) release(ownerID string, relayer core.Messager) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ownerID == noOpID {
		return
	}

	r.relayers[ownerID] = relayer
}

// Register adds a new relayer to the collection under the given ownerID.
func (r *Relayers) register(ownerID string, relayer core.Messager) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.relayers[ownerID] = relayer
}

// AcquireRelayer acquires a relayer associated with the given ownerID from the Multiplexer.
func (m *Multiplexer) AcquireRelayer(_ context.Context, ownerID string) (core.Messager, error) { //nolint:ireturn
	return m.relayers.acquire(ownerID)
}

// AcquireRandomRelayer acquires a random relayer from the Multiplexer, excluding the specified relayer.
func (m *Multiplexer) AcquireRandomRelayer(_ context.Context, excludeRelayer string) (string, core.Messager, error) { //nolint:ireturn
	return m.relayers.acquireRandom(excludeRelayer)
}

// ReleaseRelayer releases a relayer back to the Multiplexer under the given ownerID.
func (m *Multiplexer) ReleaseRelayer(_ context.Context, ownerID string, relayer core.Messager) {
	m.relayers.release(ownerID, relayer)
}

// Register registers a new relayer with the Multiplexer under the given ownerID.
func (m *Multiplexer) Register(_ context.Context, ownerID string, relayer core.Messager) {
	m.relayers.register(ownerID, relayer)
}
