package client

import (
	grpc "github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/io/gRPC"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

func ProvideEchoSphereClient(i do.Injector) (*grpc.EchoSphereClient, error) {
	cfg := do.MustInvoke[Config](i)

	clientCfg := grpc.Config{
		Logger:   do.MustInvoke[*zap.Logger](i),
		Target:   cfg.Target,
		Deadline: cfg.DeadLine,
	}

	return grpc.NewEchoSphereClient(clientCfg)
}
