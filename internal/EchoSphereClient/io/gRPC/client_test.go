package grpc_test

import (
	"context"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	esc "github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/io/gRPC"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/io/gRPC/internal/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -destination=./internal/mocks/server.mock.go -package=mocks -source=../../../../api/v1/echosphere_grpc.pb.go EchoSphereTransmissionService_TransmitServer

type grpcIntegrationSuite struct {
	suite.Suite

	SUT     *esc.EchoSphereClient
	srvCtrl *mocks.MockEchoSphereTransmissionServiceServer
}

func (g *grpcIntegrationSuite) SetupSuite() {
	ctrl := gomock.NewController(g.T())
	g.srvCtrl = mocks.NewMockEchoSphereTransmissionServiceServer(ctrl)
	listener := bufconn.Listen(1024 * 1024)

	resolver.SetDefaultScheme("passthrough")

	server := grpc.NewServer()
	v1.RegisterEchoSphereTransmissionServiceServer(server, g.srvCtrl)

	go func() { g.NoError(server.Serve(listener)) }()
	g.T().Cleanup(func() { server.Stop() })

	// Configure the client
	clientConfig := esc.Config{
		Logger: zap.NewNop(),
		Target: "mock://server.echosphere.io",
		DialOpts: []grpc.DialOption{
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return listener.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
		Deadline: 1 * time.Second,
	}

	// Create a new EchoSphere client
	client, err := esc.NewEchoSphereClient(clientConfig)
	g.Require().NoError(err)

	// Set the System Under Test (SUT)
	g.SUT = client
}

func (g *grpcIntegrationSuite) TestRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var streamChan = make(chan v1.EchoSphereTransmissionService_TransmitServer, 1)

	g.srvCtrl.
		EXPECT().
		Transmit(gomock.Any()).
		Times(1).
		DoAndReturn(func(stream v1.EchoSphereTransmissionService_TransmitServer) error {
			streamChan <- stream

			<-ctx.Done()

			return nil
		})

	go g.SUT.Run(ctx) //nolint:errcheck

	x := <-streamChan
	recv, err := x.Recv()
	g.Require().NoError(err)
	g.Require().NotEmpty(recv)

	time.Sleep(1 * time.Second) // sleep 1 seconds to force retry

	recv, err = x.Recv()
	g.Require().NoError(err)
	g.Require().NotEmpty(recv)

	messageToAck := recv.GetMessage()

	// send a different message
	err = x.Send(
		&v1.EchoSphereTransmissionServiceTransmitResponse{
			OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{
				Message: &v1.Message{
					From:    "OtherClientID",
					Content: "Different Message",
				},
			},
		},
	)
	g.Require().NoError(err)

	// wait for an ACK
	recv, err = x.Recv()
	g.Require().NoError(err)
	g.Require().NotEmpty(recv.GetAck())
	g.Require().Equal("OtherClientID", recv.GetAck().GetTo())
	g.Require().Equal("Different Message", recv.GetAck().GetContent())

	// finally we ack client message
	err = x.Send(
		&v1.EchoSphereTransmissionServiceTransmitResponse{
			OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Ack{
				Ack: &v1.Ack{
					From:    "OtherClientID",
					To:      messageToAck.GetFrom(),
					Content: messageToAck.GetContent(),
				},
			},
		},
	)
	g.Require().NoError(err)

	// Expect an EOF
	_, err = x.Recv()
	g.Require().Error(err)
}

func TestGRPCLayer(t *testing.T) {
	suite.Run(t, new(grpcIntegrationSuite))
}
