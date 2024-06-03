package usecase

import (
	"context"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
	"sync"
)

type UC struct {
	router core.RelayRouter
	mu     *sync.Mutex
}

func New(router core.RelayRouter) *UC {
	return &UC{router: router, mu: &sync.Mutex{}}
}

// sendRelayMessage sends a message through the provided relayer.
func sendRelayMessage[T *v1.Message | *v1.Ack](ctx context.Context, relayer core.Messager, msg T) error {
	switch x := any(msg).(type) {
	case *v1.Message:
		return relayer.SendMsg(ctx, &v1.EchoSphereTransmissionServiceTransmitResponse{OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{Message: x}})
	case *v1.Ack:
		return relayer.SendMsg(ctx, &v1.EchoSphereTransmissionServiceTransmitResponse{OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Ack{Ack: x}})
	}

	return nil
}
