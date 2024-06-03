package grpc

import (
	"context"
	"fmt"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/gRPC/internal/middleware"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"

	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Server is responsible for handling gRPC connections.
type Server struct {
	_ struct{}

	v1.UnimplementedEchoSphereTransmissionServiceServer

	listener   net.Listener
	gRPCServer *grpc.Server

	logger      *zap.Logger
	multiplexer core.RelayRouter
	useCase     UseCase
	serving     bool
	servingMux  sync.Mutex
}

type Config struct {
	Listener net.Listener
	Router   core.RelayRouter
	Logger   *zap.Logger
	UseCases UseCase
}

type UseCase interface {
	RegisterHandler(ctx context.Context, cmd usecase.RegisterCMD) error
	RelayHandler(ctx context.Context, cmd usecase.RelayCMD) error
	AckHandler(ctx context.Context, cmd usecase.AckCMD) error
	UnregisterHandler(ctx context.Context, cmd usecase.UnregisterCMD) error
}

// NewServer creates a new gRPC server with the provided listener.
func NewServer(cfg Config) *Server {
	s := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			middleware.StreamIdentifier(),
			middleware.StreamLogger(cfg.Logger),
			middleware.StreamMetric(),
			middleware.StreamTracing(),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	srv := &Server{
		listener:    cfg.Listener,
		gRPCServer:  s,
		multiplexer: cfg.Router,
		useCase:     cfg.UseCases,
		logger:      cfg.Logger,
		serving:     false,
	}

	v1.RegisterEchoSphereTransmissionServiceServer(s, srv)

	// Register server reflection service on your gRPC server
	reflection.Register(s)

	return srv
}

// Run starts the gRPC server and handles graceful shutdown.
func (s *Server) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		s.serving = true

		return s.gRPCServer.Serve(s.listener)
	})

	g.Go(func() error {
		<-ctx.Done()

		s.gRPCServer.Stop()

		s.servingMux.Lock()
		defer s.servingMux.Unlock()

		s.serving = false

		return nil
	})

	return g.Wait()
}

func (s *Server) Shutdown() {
	s.gRPCServer.GracefulStop()
}

func (s *Server) HealthCheck() error {
	s.servingMux.Lock()
	defer s.servingMux.Unlock()

	if !s.serving {
		return fmt.Errorf("not serving")
	}

	return nil
}
