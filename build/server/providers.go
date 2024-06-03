package server

import (
	"fmt"
	grpc "github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/gRPC"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/multiplexer"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
	"net"
)

func ProvideListener(i do.Injector) (net.Listener, error) {
	cfg := do.MustInvoke[Config](i)

	return net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
}

func ProvideUseCaseHandler(i do.Injector) (*usecase.UC, error) {
	multiplexer := do.MustInvoke[*multiplexer.Multiplexer](i)

	return usecase.New(multiplexer), nil
}

func ProvideGRPCServer(i do.Injector) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Listener: do.MustInvoke[net.Listener](i),
		Router:   do.MustInvoke[*multiplexer.Multiplexer](i),
		Logger:   do.MustInvoke[*zap.Logger](i),
		UseCases: do.MustInvoke[*usecase.UC](i),
	}), nil
}
