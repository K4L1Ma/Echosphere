package grpc

import (
	"context"
	"errors"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"golang.org/x/sync/errgroup"
	"time"
)

// Run starts the EchoSphereClient and initiates the communication with the gRPC service.
// It sends a message, starts receiving and processing responses, and handles retries.
func (esc *EchoSphereClient) Run(ctx context.Context) error {
	stream, err := esc.cli.Transmit(ctx)
	if err != nil {
		return err
	}

	if err := esc.sendMessage(stream, esc.message); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	resChan := make(chan *v1.EchoSphereTransmissionServiceTransmitResponse, 1)

	g.Go(func() error {
		return esc.receiveMessages(ctx, stream, resChan)
	})

	g.Go(func() error {
		return esc.sendMessages(ctx, stream, resChan)
	})

	if err = g.Wait(); err != nil && !errors.Is(err, ErrDone) {
		return err
	}

	return ctx.Err()
}

// receiveMessages continuously receives messages from the gRPC stream and sends them to the result channel.
// It returns an error if the context is canceled or if there's an error receiving messages.
func (esc *EchoSphereClient) receiveMessages(ctx context.Context, stream v1.EchoSphereTransmissionService_TransmitClient, resChan chan *v1.EchoSphereTransmissionServiceTransmitResponse) error {
	defer func() { esc.logger.Info("Closing ReceiveMessages") }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			recv, err := stream.Recv()
			if err != nil {
				return err
			}
			resChan <- recv
		}
	}
}

// sendMessages continuously sends messages via the gRPC stream, handles retries, and processes received messages.
// It returns an error if the context is canceled or if there's an error processing received
func (esc *EchoSphereClient) sendMessages(ctx context.Context, stream v1.EchoSphereTransmissionService_TransmitClient, resChan chan *v1.EchoSphereTransmissionServiceTransmitResponse) error {
	defer func() { esc.logger.Info("Closing SendMessage") }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(esc.deadline):
			err := esc.sendMessage(stream, esc.message)
			if err != nil {
				return err
			}
		case recv := <-resChan:
			if err := esc.processReceivedMessage(stream, recv); err != nil {
				return err
			}
		}
	}
}

// processReceivedMessage handles the received message, manages ack, and sends ack for received messages.
// It returns ErrDone if the received ack matches the message sent, or if there's an error sending an ack.
func (esc *EchoSphereClient) processReceivedMessage(stream v1.EchoSphereTransmissionService_TransmitClient, recv *v1.EchoSphereTransmissionServiceTransmitResponse) error {
	if ack := recv.GetAck(); ack != nil {
		if ack.GetTo() == esc.clientID && ack.GetContent() == esc.message.GetMessage().GetContent() {
			_ = stream.CloseSend() //nolint:errcheck

			return ErrDone
		}
	}

	if recvMessage := recv.GetMessage(); recvMessage != nil {
		if recvMessage.GetFrom() == esc.message.GetMessage().GetFrom() &&
			recvMessage.GetContent() == esc.message.GetMessage().GetContent() {
			return nil
		}

		err := esc.sendMessage(
			stream,
			&v1.EchoSphereTransmissionServiceTransmitRequest{IncomingData: &v1.EchoSphereTransmissionServiceTransmitRequest_Ack{Ack: &v1.Ack{
				From:    esc.clientID,
				To:      recvMessage.GetFrom(),
				Content: recvMessage.GetContent(),
			}}},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
