package grpc_test

import (
	"context"
	"github.com/google/uuid"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	essGRPC "github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/gRPC"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/gRPC/internal/mocks"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/multiplexer"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -destination=./internal/mocks/usecase.mock.go -package=mocks -source=server.go UseCase

type grpcIntegrationSuite struct {
	suite.Suite
	shutdownServer context.CancelFunc

	SUT     *essGRPC.Server
	client1 v1.EchoSphereTransmissionServiceClient
	client2 v1.EchoSphereTransmissionServiceClient

	useCase     *mocks.MockUseCase
	multiplexer *multiplexer.Multiplexer
}

func (g *grpcIntegrationSuite) TearDownSuite() {
	g.shutdownServer()
}

func (g *grpcIntegrationSuite) SetupSuite() {
	ctrl := gomock.NewController(g.T())

	g.T().Cleanup(func() { ctrl.Finish() })

	listen := bufconn.Listen(1024 * 1024)

	g.useCase = mocks.NewMockUseCase(ctrl)

	g.multiplexer = multiplexer.New()

	g.SUT = essGRPC.NewServer(
		essGRPC.Config{
			Listener: listen,
			Router:   g.multiplexer,
			Logger:   zap.NewNop(),
			UseCases: g.useCase,
		})

	ctx, cancelFunc := context.WithCancel(context.Background())
	g.shutdownServer = cancelFunc

	g.Require().Error(g.SUT.HealthCheck())

	go func() { g.NoError(g.SUT.Run(ctx)) }()

	g.Require().Eventually(
		func() bool { return g.SUT.HealthCheck() == nil },
		time.Second,
		time.Millisecond,
	)

	resolver.SetDefaultScheme("passthrough")

	conn, err := grpc.NewClient(
		"mock://server.echosphere.io",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listen.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	g.Require().NoError(err)

	g.client1 = v1.NewEchoSphereTransmissionServiceClient(conn)
	g.client2 = v1.NewEchoSphereTransmissionServiceClient(conn)
}

func (g *grpcIntegrationSuite) TestTransmitMessage() {
	clientID, err := uuid.NewV7()
	g.Require().NoError(err)

	transmit, err := g.client1.Transmit(context.Background())
	g.Require().NoError(err)

	x := make(chan usecase.RelayCMD, 1)

	closeChan := make(chan struct{})
	g.useCase.EXPECT().RegisterHandler(gomock.Any(), gomock.Any()).Times(1)
	g.useCase.EXPECT().UnregisterHandler(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(func(any, any) error {
		close(closeChan)
		return nil
	})

	g.useCase.EXPECT().RelayHandler(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, cmd usecase.RelayCMD) error {
		x <- cmd

		return nil
	})

	msg := &v1.EchoSphereTransmissionServiceTransmitRequest{IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
		Message: &v1.Message{
			From:    clientID.String(),
			Content: "x",
		},
	}}
	err = transmit.Send(msg)
	g.Require().NoError(err)

	cmd := <-x
	g.Require().Equal(msg.GetMessage().GetFrom(), cmd.From)
	g.Require().Equal(msg.GetMessage().GetContent(), cmd.Content)

	err = transmit.CloseSend()
	g.Require().NoError(err)
	g.Eventually(func() bool {
		select {
		case <-closeChan:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond, "did not receive cancel signal")
}

func (g *grpcIntegrationSuite) TestTransmitAck() {
	clientID, err := uuid.NewV7()
	g.Require().NoError(err)

	clientID2, err := uuid.NewV7()
	g.Require().NoError(err)

	transmit, err := g.client1.Transmit(context.Background())
	g.Require().NoError(err)

	x := make(chan usecase.AckCMD, 1)

	g.useCase.EXPECT().UnregisterHandler(gomock.Any(), gomock.Any()).Times(1)

	g.useCase.EXPECT().AckHandler(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, cmd usecase.AckCMD) error {
		x <- cmd

		return nil
	})

	msg := &v1.EchoSphereTransmissionServiceTransmitRequest{IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Ack{
		Ack: &v1.Ack{
			From:    clientID2.String(),
			To:      clientID.String(),
			Content: "x",
		},
	}}
	err = transmit.Send(msg)
	g.Require().NoError(err)

	cmd := <-x
	g.Require().Equal(msg.GetAck().GetFrom(), cmd.From)
	g.Require().Equal(msg.GetAck().GetContent(), cmd.Content)

	err = transmit.CloseSend()
	g.Require().NoError(err)
}

func TestGRPCLayer(t *testing.T) {
	suite.Run(t, new(grpcIntegrationSuite))
}
