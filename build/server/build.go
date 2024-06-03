package server

import (
	"context"
	"github.com/k4l1ma/EchoSphere/build/common"
	grpc "github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/gRPC"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/multiplexer"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
)

const Name = "server.echosphere.io"

func Run(ctx context.Context, cfg Config) error {
	cleanUp := common.InitTracing(ctx, Name)
	defer cleanUp()

	diContainer := do.New()

	do.ProvideValue[Config](diContainer, cfg)
	do.Provide[net.Listener](diContainer, ProvideListener)
	do.ProvideValue[*zap.Logger](diContainer, zap.Must(zap.NewProduction()).Named(Name))
	do.ProvideValue[*multiplexer.Multiplexer](diContainer, multiplexer.New())
	do.Provide[*usecase.UC](diContainer, ProvideUseCaseHandler)
	do.Provide[common.HTTPSideCarServer](diContainer, common.ProvideHTTPSideCar[Config])
	do.Provide[*grpc.Server](diContainer, ProvideGRPCServer)

	gRPCServer := do.MustInvoke[*grpc.Server](diContainer)
	httpSideCar := do.MustInvoke[common.HTTPSideCarServer](diContainer)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return gRPCServer.Run(ctx) })
	g.Go(func() error { return httpSideCar.Run(ctx) })

	return g.Wait()
}
