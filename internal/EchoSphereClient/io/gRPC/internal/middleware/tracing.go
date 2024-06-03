package middleware

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// StreamTracing returns a gRPC client stream interceptor that traces streaming RPC calls.
func StreamTracing() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		tracer := otel.GetTracerProvider().Tracer("client.echosphere.io/grpc")

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return &tracingClientStream{
			ClientStream: clientStream,
			tracer:       tracer,
			method:       method,
		}, nil
	}
}

type tracingClientStream struct {
	grpc.ClientStream
	tracer trace.Tracer
	method string
}

func (s *tracingClientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)

	_, span := s.tracer.Start(s.Context(), "RecvMsg", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	// Marshal the message to bytes for tracing
	bytes, err := protojson.Marshal(m.(proto.Message))
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return err
	}

	span.SetAttributes(attribute.String("Message", string(bytes)))

	return err
}

func (s *tracingClientStream) SendMsg(m interface{}) error {
	err := s.ClientStream.SendMsg(m)

	_, span := s.tracer.Start(s.Context(), "SendMsg", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	// Marshal the message to bytes for tracing
	bytes, err := protojson.Marshal(m.(proto.Message))
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return err
	}

	span.SetAttributes(attribute.String("Message", string(bytes)))

	return err
}
