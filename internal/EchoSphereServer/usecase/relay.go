// Package multiplexer provides a mechanism to manage and route messages through different relayers.
package usecase

import (
	"context"
	"errors"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
)

// RelayCMD represents a command to relay a message.
type RelayCMD struct {
	From    string
	Content string
}

// RelayHandler handles the relay of a message from one owner to a random relayer and back.
func (uc *UC) RelayHandler(ctx context.Context, cmd RelayCMD) error {
	// this needs to be an ACID transaction we don't want a relayer being stole while we do actions
	uc.mu.Lock()
	defer uc.mu.Unlock()

	randomOwnerID, randomRelayer, err := uc.router.AcquireRandomRelayer(ctx, cmd.From)
	if err != nil && !errors.Is(err, core.ErrFailedToGetRelayer) {
		return err
	}
	defer uc.router.ReleaseRelayer(ctx, randomOwnerID, randomRelayer)

	if err := sendRelayMessage(ctx, randomRelayer, &v1.Message{From: cmd.From, Content: cmd.Content}); err != nil {
		return err
	}

	ownerRelayer, err := uc.router.AcquireRelayer(ctx, cmd.From)
	if err != nil {
		return err
	}
	defer uc.router.ReleaseRelayer(ctx, cmd.From, ownerRelayer)

	return sendRelayMessage(ctx, ownerRelayer, &v1.Message{From: cmd.From, Content: cmd.Content})
}
