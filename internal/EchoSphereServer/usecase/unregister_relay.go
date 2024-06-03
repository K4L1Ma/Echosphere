package usecase

import (
	"context"
)

type UnregisterCMD struct {
	OwnerID string
}

func (uc *UC) UnregisterHandler(ctx context.Context, cmd UnregisterCMD) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	_, err := uc.router.AcquireRelayer(ctx, cmd.OwnerID)

	return err
}
