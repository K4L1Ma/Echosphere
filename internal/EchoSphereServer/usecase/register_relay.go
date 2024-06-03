package usecase

import (
	"context"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
)

type RegisterCMD struct {
	OwnerID      string
	StreamSender core.Messager
}

func (uc *UC) RegisterHandler(ctx context.Context, cmd RegisterCMD) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	uc.router.Register(ctx, cmd.OwnerID, cmd.StreamSender)

	return nil
}
