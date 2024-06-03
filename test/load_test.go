package load_test

import (
	"context"
	"github.com/google/uuid"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/build/server"
	esc "github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/io/gRPC"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync/atomic"
	"testing"
	"time"
)

type loadTestSuite struct {
	suite.Suite
	cancel context.CancelFunc
}

func (l *loadTestSuite) TearDownSuite() {
	l.cancel()
}

func (l *loadTestSuite) SetupSuite() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	l.cancel = cancelFunc

	go func() {
		err := server.Run(ctx, server.Config{
			Server: server.SrvCfg{
				Port: 8080,
			},
			SideCar: server.SideCarCfg{
				Enabled: true,
				Port:    9090,
			},
		})

		l.ErrorIs(err, context.Canceled)
	}()
}

func (l *loadTestSuite) TestLoad100k() {
	g, ctx := errgroup.WithContext(context.Background())

	rand.Seed(uint64(time.Now().UnixNano()))

	failedCons := atomic.Int64{}

	const load = 100000

	for _, _ = range [load]struct{}{} {
		g.Go(func() error {
			conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
			l.NoError(err)

			client := v1.NewEchoSphereTransmissionServiceClient(conn)

			stream, err := client.Transmit(ctx)
			if err != nil {
				failedCons.Add(1)
				return nil
			}

			time.Sleep(time.Duration(rand.Intn(60)) * time.Second)

			message := &v1.Message{
				From:    uuid.Must(uuid.NewV7()).String(),
				Content: esc.GenerateMessage(),
			}

			err = stream.Send(&v1.EchoSphereTransmissionServiceTransmitRequest{
				IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
					Message: message,
				},
			})
			l.NoError(err)

			for {
				recv, err := stream.Recv()
				l.NoError(err)

				if rcvMsg := recv.GetMessage(); rcvMsg != nil && rcvMsg == message {
					continue
				}

				if rcvMsg := recv.GetMessage(); rcvMsg != nil {
					time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

					err = stream.Send(&v1.EchoSphereTransmissionServiceTransmitRequest{
						IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Ack{
							Ack: &v1.Ack{
								From:    message.GetFrom(),
								To:      rcvMsg.GetFrom(),
								Content: rcvMsg.GetContent(),
							},
						},
					})
					l.NoError(err)
				}

				if rcvMsg := recv.GetAck(); rcvMsg != nil {
					if rcvMsg.GetTo() == message.GetFrom() && rcvMsg.GetContent() == message.GetContent() {
						break
					}

					continue
				}

			}

			return stream.CloseSend()
		})
	}

	l.Require().Empty(g.Wait())
	failed := failedCons.Load()
	l.T().Logf("Successful Connections: %d\n", load-failed)
	l.T().Logf("Failed Connections: %d\n", failed)
}

func TestLoad(t *testing.T) {
	suite.Run(t, new(loadTestSuite))
}
