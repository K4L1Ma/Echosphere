package middleware

import (
	"context"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"google.golang.org/grpc"
)

type CtxKey string

const ClientSourceCtxKey CtxKey = "client-source"

func StreamIdentifier() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &identifierServerStream{ServerStream: stream, ctx: stream.Context()})
	}
}

type identifierServerStream struct {
	_ struct{}
	grpc.ServerStream

	ctx context.Context
}

func (i *identifierServerStream) Context() context.Context {
	return i.ctx
}

func (i *identifierServerStream) RecvMsg(m any) error {
	err := i.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}

	request, ok := m.(*v1.EchoSphereTransmissionServiceTransmitRequest)
	if !ok {
		return nil
	}

	var From string

	if request.GetMessage() != nil {
		From = request.GetMessage().GetFrom()
	}

	if request.GetAck() != nil {
		From = request.GetAck().GetFrom()
	}

	i.ctx = context.WithValue(i.Context(), ClientSourceCtxKey, From)

	return nil
}
