package middleware

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"time"
)

// StreamLogger returns a grpc.StreamServerInterceptor that logs stream events.
func StreamLogger(logger *zap.Logger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return &loggerServerStream{
			ClientStream: stream,
			logger:       logger,
		}, nil
	}
}

// loggerServerStream wraps a grpc.ServerStream and logs received and sent messages.
type loggerServerStream struct {
	grpc.ClientStream
	logger *zap.Logger
}

// RecvMsg logs the received message and the time taken to handle it.
func (s *loggerServerStream) RecvMsg(m any) error {
	start := time.Now()

	s.logMessage("Handling RecvMsg", m)

	err := s.ClientStream.RecvMsg(m)

	duration := time.Since(start)
	s.logMessage("Finished handling RecvMsg", m, zap.Duration("duration", duration), zap.Error(err))

	return err
}

// SendMsg logs the sent message and the time taken to handle it.
func (s *loggerServerStream) SendMsg(m any) error {
	start := time.Now()

	s.logMessage("Handling SendMsg", m)

	err := s.ClientStream.SendMsg(m)

	duration := time.Since(start)
	s.logMessage("Finished handling SendMsg", m, zap.Duration("duration", duration), zap.Error(err))

	return err
}

// logMessage logs a message with additional fields.
func (s *loggerServerStream) logMessage(msg string, m any, additionalFields ...zap.Field) {
	additionalFields = append(additionalFields, s.getMessageField(m))
	s.logger.Info(msg, additionalFields...)
}

// getMessageField returns a zap.Field for logging the message.
func (s *loggerServerStream) getMessageField(m any) zap.Field {
	if marshaler, ok := m.(zapcore.ObjectMarshaler); ok {
		return zap.Object("message", marshaler)
	}

	return zap.Any("message", m)
}
