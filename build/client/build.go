package client

import (
	"context"
	"github.com/k4l1ma/EchoSphere/build/common"
	grpc "github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/io/gRPC"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const Name = "client.echosphre.io"

func Run(ctx context.Context, cfg Config) error {
	cleanUp := common.InitTracing(ctx, Name)
	defer cleanUp()

	diContainer := do.New()

	do.ProvideValue[Config](diContainer, cfg)
	do.ProvideValue[*zap.Logger](diContainer, zap.Must(zap.NewProduction()).Named(Name))
	do.Provide[*grpc.EchoSphereClient](diContainer, ProvideEchoSphereClient)
	do.Provide[common.HTTPSideCarServer](diContainer, common.ProvideHTTPSideCar[Config])

	echoSphereClient := do.MustInvoke[*grpc.EchoSphereClient](diContainer)
	httpSideCar := do.MustInvoke[common.HTTPSideCarServer](diContainer)
	logger := do.MustInvoke[*zap.Logger](diContainer)

	g, ctx := errgroup.WithContext(ctx)

	logger.Info("Starting EchoSphere Client")

	g.Go(func() error { return echoSphereClient.Run(ctx) })

	if cfg.SideCar.Enabled {
		g.Go(func() error { return httpSideCar.Run(ctx) })
	}

	err := g.Wait()

	logger.Info("Closing EchoSphere Client")

	return err
}
