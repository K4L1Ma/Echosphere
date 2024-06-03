package test

import (
	"context"
	"github.com/google/uuid"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/build/client"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/test/internal/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -destination=./internal/mocks/server.mock.go -package=mocks -source=../../../api/v1/echosphere_grpc.pb.go EchoSphereTransmissionService_TransmitServer

type ClientAcceptanceSuite struct {
	suite.Suite
	srvCtrl *mocks.MockEchoSphereTransmissionServiceServer
	server  *grpc.Server
}

func (s *ClientAcceptanceSuite) TearDownSuite() {
	s.server.Stop()
}

func (s *ClientAcceptanceSuite) SetupSuite() {
	// Configure Mock gRPC Stream Server
	ctrl := gomock.NewController(s.T())
	s.srvCtrl = mocks.NewMockEchoSphereTransmissionServiceServer(ctrl)

	listener, err := net.Listen("tcp", "localhost:8080")
	s.Require().NoError(err)

	_ = listener

	s.server = grpc.NewServer()
	v1.RegisterEchoSphereTransmissionServiceServer(s.server, s.srvCtrl)

	go func() { s.NoError(s.server.Serve(listener)) }()
}

// TestClientSendAMessage:
//
//	Scenario: Client connects to the server and sends message X
//	  Given a server is running and listening for connections
//	  When the client connects to the server
//	  And the client sends "message X"
//	  Then the client should listen for responses from the server
func (s *ClientAcceptanceSuite) TestClientSendAMessage() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var streamChan = make(chan v1.EchoSphereTransmissionService_TransmitServer, 1)

	s.srvCtrl.EXPECT().Transmit(gomock.Any()).DoAndReturn(func(iStream v1.EchoSphereTransmissionService_TransmitServer) error {
		streamChan <- iStream

		<-ctx.Done()

		return nil
	})

	go func() {
		err := client.Run(ctx, client.Config{
			Target:   "localhost:8080",
			DeadLine: 30 * time.Second,
			SideCar: client.SideCar{
				Enabled: true,
				Port:    9091,
			},
		})

		s.ErrorIs(err, context.Canceled)
	}()

	stream := <-streamChan

	// Client sends it registering message
	recv, err := stream.Recv()
	s.Require().NoError(err)
	s.Require().NotEmpty(recv)
	s.Require().NotEmpty(recv.GetMessage())

	// Server replies
	err = stream.Send(&v1.EchoSphereTransmissionServiceTransmitResponse{
		OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{
			Message: recv.GetMessage(),
		},
	})
	s.Require().NoError(err)
}

// TestOnAckClientCloses:
//
//	Scenario: Client receives ok X and closes the connection
//	  Given a client is connected to the server
//	  And the client has sent "message X"
//	  When the client receives "ok X"
//	  Then the client should close the connection
func (s *ClientAcceptanceSuite) TestOnAckClientCloses() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	var streamChan = make(chan v1.EchoSphereTransmissionService_TransmitServer, 1)

	s.srvCtrl.EXPECT().Transmit(gomock.Any()).DoAndReturn(func(iStream v1.EchoSphereTransmissionService_TransmitServer) error {
		streamChan <- iStream

		<-ctx.Done()

		return nil
	})

	closeChan := make(chan struct{}, 1)

	go func() {
		err := client.Run(ctx, client.Config{
			Target:   "localhost:8080",
			DeadLine: 30 * time.Second,
			SideCar: client.SideCar{
				Enabled: true,
				Port:    9091,
			},
		})

		s.ErrorIs(err, context.Canceled)

		close(closeChan)
	}()

	stream := <-streamChan

	// Client sends it registering message
	recv, err := stream.Recv()
	s.Require().NoError(err)
	s.Require().NotEmpty(recv)
	s.Require().NotEmpty(recv.GetMessage())

	// Someone ACK
	err = stream.Send(&v1.EchoSphereTransmissionServiceTransmitResponse{
		OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Ack{
			Ack: &v1.Ack{
				From:    uuid.Must(uuid.NewV7()).String(),
				To:      recv.GetMessage().GetFrom(),
				Content: recv.GetMessage().GetContent(),
			},
		},
	})

	s.Require().NoError(err)

	cancelFunc()

	s.Eventually(func() bool {
		select {
		case <-closeChan:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond, "did not receive cancel signal")
}

// TestClientAckMessage:
//
//	Scenario: Client receives message Y and replies with ok Y
//	  Given a client is connected to the server
//	  When the client receives "message Y"
//	  Then the client should reply with "ok Y"
func (s *ClientAcceptanceSuite) TestClientAckMessage() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var streamChan = make(chan v1.EchoSphereTransmissionService_TransmitServer, 1)

	s.srvCtrl.EXPECT().Transmit(gomock.Any()).DoAndReturn(func(iStream v1.EchoSphereTransmissionService_TransmitServer) error {
		streamChan <- iStream

		<-ctx.Done()

		return nil
	})

	go func() {
		err := client.Run(ctx, client.Config{
			Target:   "localhost:8080",
			DeadLine: 30 * time.Second,
			SideCar: client.SideCar{
				Enabled: true,
				Port:    9091,
			},
		})

		s.ErrorIs(err, context.Canceled)
	}()

	stream := <-streamChan

	// Client sends it registering message
	recv, err := stream.Recv()
	s.Require().NoError(err)
	s.Require().NotEmpty(recv)
	s.Require().NotEmpty(recv.GetMessage())

	message := &v1.Message{
		From:    uuid.Must(uuid.NewV7()).String(),
		Content: "1",
	}

	// Someone Sends a message
	err = stream.Send(&v1.EchoSphereTransmissionServiceTransmitResponse{
		OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{
			Message: message,
		},
	})
	s.Require().NoError(err)

	// Client Ack it
	recv, err = stream.Recv()
	s.Require().NoError(err)
	s.Require().Equal(message.GetFrom(), recv.GetAck().GetTo())
	s.Require().Equal(message.GetContent(), recv.GetAck().GetContent())
}

// TestClientResendMessage:
//
//	Scenario: Client resends message X if not acknowledged within timeout
//	  Given a client is connected to the server
//	  And the client has sent "message X"
//	  When the client does not receive "ok X" within the timeout period
//	  Then the client should resend "message X"
//	  And this process should repeat until "ok X" is received
func (s *ClientAcceptanceSuite) TestClientResendMessage() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var streamChan = make(chan v1.EchoSphereTransmissionService_TransmitServer, 1)

	s.srvCtrl.EXPECT().Transmit(gomock.Any()).DoAndReturn(func(iStream v1.EchoSphereTransmissionService_TransmitServer) error {
		streamChan <- iStream

		<-ctx.Done()

		return nil
	})

	go func() {
		err := client.Run(ctx, client.Config{
			Target:   "localhost:8080",
			DeadLine: 1 * time.Second,
			SideCar: client.SideCar{
				Enabled: true,
				Port:    9091,
			},
		})

		s.ErrorIs(err, context.Canceled)
	}()

	stream := <-streamChan

	// Client sends it registering message
	recv, err := stream.Recv()
	s.Require().NoError(err)
	s.Require().NotEmpty(recv)
	s.Require().NotEmpty(recv.GetMessage())

	time.Sleep(time.Second)

	recv2, err := stream.Recv()
	s.Require().NoError(err)
	s.Require().NotEmpty(recv2)
	s.Require().NotEmpty(recv2.GetMessage())
	s.Require().Equal(recv.GetMessage(), recv2.GetMessage())
}

func TestClientAcceptance(t *testing.T) {
	suite.Run(t, new(ClientAcceptanceSuite))
}
