package core

import "context"

type Messager interface {
	SendMsg(ctx context.Context, m any) error
}

type RelayRouter interface {
	Register(ctx context.Context, ownerID string, relayer Messager)
	AcquireRelayer(ctx context.Context, ownerID string) (Messager, error)
	AcquireRandomRelayer(ctx context.Context, excludeRelayer string) (OwnerID string, Relayer Messager, err error)
	ReleaseRelayer(ctx context.Context, ownerID string, relayer Messager)
}
