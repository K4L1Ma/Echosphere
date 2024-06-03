// Package grpc provides functionality for interacting with gRPC services in the EchoSphere application.
package grpc

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereClient/io/gRPC/internal/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ErrDone represents a done error.
var ErrDone = fmt.Errorf("done")

// EchoSphereClient is a client for interacting with the EchoSphereTransmissionService gRPC service.
type EchoSphereClient struct {
	cli      v1.EchoSphereTransmissionServiceClient
	logger   *zap.Logger
	clientID string
	message  *v1.EchoSphereTransmissionServiceTransmitRequest
	deadline time.Duration
}

// Config represents the configuration for creating a new EchoSphereClient.
type Config struct {
	Logger   *zap.Logger
	Target   string
	DialOpts []grpc.DialOption
	Deadline time.Duration
}

// NewEchoSphereClient creates a new EchoSphereClient instance.
func NewEchoSphereClient(cfg Config) (*EchoSphereClient, error) {
	cfg.DialOpts = append(
		cfg.DialOpts,
		grpc.WithChainStreamInterceptor(
			middleware.StreamLogger(cfg.Logger),
			middleware.StreamMetric(),
			middleware.StreamTracing(),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(cfg.Target, cfg.DialOpts...)
	if err != nil {
		return nil, err
	}

	client := v1.NewEchoSphereTransmissionServiceClient(conn)

	cliID := generateClientID()
	message := newMessage(cliID)

	return &EchoSphereClient{
		cli:      client,
		logger:   cfg.Logger,
		clientID: cliID,
		deadline: cfg.Deadline,
		message:  message,
	}, nil
}

// newMessage creates a new message to be transmitted.
func newMessage(clientID string) *v1.EchoSphereTransmissionServiceTransmitRequest {
	return &v1.EchoSphereTransmissionServiceTransmitRequest{
		IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Message{
			Message: &v1.Message{
				From:    clientID,
				Content: GenerateMessage(),
			},
		},
	}
}

// sendMessage sends a message via the gRPC stream.
func (esc *EchoSphereClient) sendMessage(stream v1.EchoSphereTransmissionService_TransmitClient, msg *v1.EchoSphereTransmissionServiceTransmitRequest) error {
	err := stream.SendMsg(msg)
	if err != nil {
		return err
	}

	return nil
}

// generateClientID generates a unique client ID.
func generateClientID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// GenerateMessage generates a random message.
func GenerateMessage() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rand.Seed(uint64(time.Now().UnixNano()))

	message := make([]byte, 12)

	for i := range message {
		randomIndex := rand.Intn(len(charset))
		message[i] = charset[randomIndex]
	}

	return string(message)
}
