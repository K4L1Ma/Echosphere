package grpc

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/io/gRPC/internal/middleware"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
)

// Transmit handles incoming stream messages and relays them.
func (s *Server) Transmit(stream v1.EchoSphereTransmissionService_TransmitServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			req, err := stream.Recv()
			if err != nil {
				s.unregister(stream.Context()) //nolint:errcheck

				return lo.Ternary(!errors.Is(err, io.EOF), err, nil)
			}

			err = s.handleRequest(stream.Context(), req, stream)
			if err != nil {
				s.unregister(stream.Context()) //nolint:errcheck

				return err
			}
		}
	}
}

func (s *Server) handleRequest(ctx context.Context, req *v1.EchoSphereTransmissionServiceTransmitRequest, stream grpc.ServerStream) error {
	if message := req.GetMessage(); message != nil {
		if err := s.handleMessage(ctx, stream, message); err != nil {
			return err
		}
	}

	if ok := req.GetAck(); ok != nil {
		if err := s.handleAck(ctx, ok); err != nil && !errors.Is(err, core.ErrFailedToGetRelayer) {
			return err
		}
	}

	return nil
}

func (s *Server) handleMessage(ctx context.Context, stream grpc.ServerStream, message *v1.Message) error {
	err := s.useCase.RegisterHandler(ctx, usecase.RegisterCMD{
		OwnerID:      message.GetFrom(),
		StreamSender: StreamSender{stream: stream},
	})
	if err != nil {
		return err
	}

	s.logger.Info("Registered client successfully", zap.String("client-id", message.GetFrom()))

	err = s.useCase.RelayHandler(
		ctx,
		usecase.RelayCMD{
			From:    message.GetFrom(),
			Content: message.GetContent(),
		},
	)
	if err != nil {
		return fmt.Errorf("error handling message: %w", err)
	}

	return nil
}

func (s *Server) handleAck(ctx context.Context, ackMessage *v1.Ack) error {
	err := s.useCase.AckHandler(
		ctx,
		usecase.AckCMD{
			From:    ackMessage.GetFrom(),
			To:      ackMessage.To, //nolint:protogetter
			Content: ackMessage.GetContent(),
		},
	)
	if err != nil {
		return fmt.Errorf("error handling ack message: %w", err)
	}

	return nil
}

func (s *Server) unregister(ctx context.Context) error {
	ownerID, ok := ctx.Value(middleware.ClientSourceCtxKey).(string)
	if !ok {
		return errors.New("missing client source in context")
	}

	err := s.useCase.UnregisterHandler(ctx, usecase.UnregisterCMD{OwnerID: ownerID})
	if err != nil {
		s.logger.Error("Error unregistering:", zap.Error(err))

		return err
	}

	s.logger.Info("Unregister process completed successfully", zap.String("clientID", ownerID))

	return nil
}

// StreamSender provides a method to send messages via gRPC stream.
type StreamSender struct {
	stream grpc.ServerStream
}

// SendMsg sends a message via the gRPC stream.
func (s StreamSender) SendMsg(_ context.Context, m any) error {
	return s.stream.SendMsg(m)
}
