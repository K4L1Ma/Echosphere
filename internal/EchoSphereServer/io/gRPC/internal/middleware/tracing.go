package middleware

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func StreamTracing() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &tracingServerStream{ServerStream: stream})
	}
}

type tracingServerStream struct {
	_ struct{}
	grpc.ServerStream
}

func (s *tracingServerStream) RecvMsg(m any) error {
	tracer := otel.GetTracerProvider().Tracer("server.echosphere.io/grpc/reciver")

	bytes, err := protojson.Marshal(m.(proto.Message))
	if err != nil {
		return err
	}

	_, span := tracer.Start(s.Context(), "RecvMsg", trace.WithAttributes(attribute.String("Message", string(bytes))))
	defer span.End()

	err = s.ServerStream.RecvMsg(m)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (s *tracingServerStream) SendMsg(m any) error {
	tracer := otel.GetTracerProvider().Tracer("server.echosphere.io/grpc/sender")

	bytes, err := protojson.Marshal(m.(proto.Message))
	if err != nil {
		return err
	}

	_, span := tracer.Start(s.Context(), "SendMsg", trace.WithAttributes(attribute.String("Message", string(bytes))))
	defer span.End()

	err = s.ServerStream.SendMsg(m)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
