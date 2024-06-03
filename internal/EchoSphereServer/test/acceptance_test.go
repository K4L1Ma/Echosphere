package test

import (
	"context"
	"github.com/google/uuid"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/build/server"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

type ServerAcceptanceSuite struct {
	suite.Suite
	cancel   context.CancelFunc
	errGroup errgroup.Group
}

func (s *ServerAcceptanceSuite) TearDownSuite() {
	s.cancel()
	s.NoError(s.errGroup.Wait())
}

func (s *ServerAcceptanceSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.errGroup = errgroup.Group{}
	s.errGroup.Go(func() error {
		err := server.Run(ctx, server.Config{
			Server: server.SrvCfg{
				Port: 8080,
			},
			SideCar: server.SideCarCfg{
				Enabled: false,
				Port:    9090,
			},
		})

		return err
	})
}

// TestServerForwardsRandom:
//
//	Scenario: Server receives message X and forwards it to a randomly selected connection
//	Given a server is running and listening for connections
//	And a client is connected to the server
//	When the client sends "message X"
//	Then the server should forward "message X" to a randomly selected connection
func (s *ServerAcceptanceSuite) Test1ServerForwardsRandom() {
	g := errgroup.Group{}

	// First client sends a message and verifies the server's behavior
	g.Go(func() error {
		conn, err := grpc.NewClient(
			"localhost:8080",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		s.Require().NoError(err)

		client1 := v1.NewEchoSphereTransmissionServiceClient(conn)
		stream, err := client1.Transmit(context.Background())
		s.Require().NoError(err)

		err = stream.Send(
			&v1.EchoSphereTransmissionServiceTransmitRequest{
				IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
					Message: &v1.Message{
						From:    uuid.Must(uuid.NewV7()).String(),
						Content: "1",
					},
				},
			},
		)
		s.Require().NoError(err)

		recv, err := stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)

		recv, err = stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)
		s.Require().Equal("2", recv.GetMessage().GetContent())

		s.Require().NoError(stream.CloseSend())

		return nil
	})

	// Small delay to ensure the first client's message is processed
	time.Sleep(125 * time.Millisecond)

	// Second client sends a message and verifies the server's behavior
	// Having only two clients narrow it down the random as it can not ever be the same client
	g.Go(func() error {
		conn, err := grpc.NewClient(
			"localhost:8080",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		s.Require().NoError(err)

		client1 := v1.NewEchoSphereTransmissionServiceClient(conn)
		stream, err := client1.Transmit(context.Background())
		s.Require().NoError(err)

		err = stream.Send(
			&v1.EchoSphereTransmissionServiceTransmitRequest{
				IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
					Message: &v1.Message{
						From:    uuid.Must(uuid.NewV7()).String(),
						Content: "2",
					},
				},
			},
		)
		s.Require().NoError(err)

		recv, err := stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)

		s.Require().NoError(stream.CloseSend())

		return nil
	})

	s.Require().NoError(g.Wait())
}

// TestServerClientGetAck:
//
//	Scenario: Server receives ok X and forwards it to the original sender
//	  Given a server is running and listening for connections
//	  And a client is connected to the server
//	  And the client has previously sent "message X"
//	  When the server receives "ok X"
//	  Then the server should forward "ok X" to the original sender of "message X"
func (s *ServerAcceptanceSuite) TestServerClientGetAck() {
	g := errgroup.Group{}

	// First client sends a message and verifies the server's behavior
	g.Go(func() error {
		conn, err := grpc.NewClient(
			"localhost:8080",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		s.Require().NoError(err)

		client1 := v1.NewEchoSphereTransmissionServiceClient(conn)
		stream, err := client1.Transmit(context.Background())
		s.Require().NoError(err)

		clientID := uuid.Must(uuid.NewV7()).String()

		err = stream.Send(
			&v1.EchoSphereTransmissionServiceTransmitRequest{
				IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
					Message: &v1.Message{
						From:    clientID,
						Content: "1",
					},
				},
			},
		)
		s.Require().NoError(err)

		// This message Register the Client
		recv, err := stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)

		// This is the Peer Message
		recv, err = stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)
		s.Require().Equal("2", recv.GetMessage().GetContent())

		// Now is going to Ack
		err = stream.Send(
			&v1.EchoSphereTransmissionServiceTransmitRequest{
				IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Ack{
					Ack: &v1.Ack{
						From:    clientID,
						To:      recv.GetMessage().GetFrom(),
						Content: recv.GetMessage().GetContent(),
					},
				},
			},
		)
		s.Require().NoError(err)

		s.Require().NoError(stream.CloseSend())

		return nil
	})

	// Small delay to ensure the first client's message is processed
	time.Sleep(125 * time.Millisecond)

	// Second client sends a message and verifies the server's behavior
	// Having only two clients narrow it down the random as it can not ever be the same client
	g.Go(func() error {
		conn, err := grpc.NewClient(
			"localhost:8080",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		s.Require().NoError(err)

		client1 := v1.NewEchoSphereTransmissionServiceClient(conn)
		stream, err := client1.Transmit(context.Background())
		s.Require().NoError(err)

		err = stream.Send(
			&v1.EchoSphereTransmissionServiceTransmitRequest{
				IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
					Message: &v1.Message{
						From:    uuid.Must(uuid.NewV7()).String(),
						Content: "2",
					},
				},
			},
		)
		s.Require().NoError(err)

		recv, err := stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)

		// This is the Original Message Ack
		recv, err = stream.Recv()
		s.Require().NoError(err)
		s.Require().NotEmpty(recv)
		s.Require().NotEmpty(recv.GetAck())

		s.Require().NoError(stream.CloseSend())

		return nil
	})

	s.Require().NoError(g.Wait())
}

// TestServerRemoveClosedConnections:
//
//	Scenario: Server removes closed connection from active connections
//	Given a server is running and listening for connections
//	And a client is connected to the server
//	When the client disconnects
//	Then the server should remove the connection from the active connections
func (s *ServerAcceptanceSuite) TestServerRemoveClosedConnections() {
	// TODO: Needs testing strategy. Functionality seems to be working based on log observations.
}

func TestServerAcceptance(t *testing.T) {
	suite.Run(t, new(ServerAcceptanceSuite))
}
