package middleware

import (
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"time"
)

func StreamLogger(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Execute the next handler in the chain
		return handler(srv, &loggerServerStream{ServerStream: stream, logger: logger})
	}
}

type loggerServerStream struct {
	_ struct{}
	grpc.ServerStream
	logger *zap.Logger
}

func (s *loggerServerStream) RecvMsg(m any) error {
	start := time.Now()

	if !lo.IsEmpty(m.(*v1.EchoSphereTransmissionServiceTransmitRequest).IncomingData) {
		s.logMessage("Recived Msg", m)
	}

	err := s.ServerStream.RecvMsg(m)

	s.logMessage("Finished handling RecvMsg", m, zap.Duration("duration", time.Since(start)), zap.Error(err))

	return err
}

func (s *loggerServerStream) SendMsg(m any) error {
	start := time.Now()

	s.logMessage("Handling SendMsg", m)

	err := s.ServerStream.SendMsg(m)

	s.logMessage("Finished handling SendMsg", m, zap.Duration("duration", time.Since(start)), zap.Error(err))

	return err
}

func (s *loggerServerStream) logMessage(msg string, m any, additionalFields ...zap.Field) {
	additionalFields = append(additionalFields, s.getMessageField(m))
	s.logger.Info(msg, additionalFields...)
}

func (s *loggerServerStream) getMessageField(m any) zap.Field {
	if marshaler, ok := m.(zapcore.ObjectMarshaler); ok {
		return zap.Object("message", marshaler)
	}

	return zap.Any("message", m)
}
