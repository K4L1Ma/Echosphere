// Package usecase provides a mechanism to manage and route messages through different relayers.
package usecase

import (
	"context"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
)

// AckCMD represents a command to acknowledge a message.
type AckCMD struct {
	From    string
	To      string
	Content string
}

// AckHandler handles the acknowledgment of a message for a specific owner.
func (uc *UC) AckHandler(ctx context.Context, cmd AckCMD) error {
	// this needs to be an ACID transaction we don't want a relayer being stole while we do actions
	uc.mu.Lock()
	defer uc.mu.Unlock()

	relayer, err := uc.router.AcquireRelayer(ctx, cmd.To)
	if err != nil {
		return err
	}
	defer uc.router.ReleaseRelayer(ctx, cmd.To, relayer)

	return sendRelayMessage(ctx, relayer, &v1.Ack{From: cmd.From, To: cmd.To, Content: cmd.Content})
}
