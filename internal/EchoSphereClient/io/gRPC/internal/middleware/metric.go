package middleware

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
)

// MetricsRecorder records metrics for gRPC streaming.
type MetricsRecorder struct {
	requestsCounter        metric.Int64Counter
	successCounter         metric.Int64Counter
	failureCounter         metric.Int64Counter
	executionTimeHistogram metric.Float64Histogram
}

// newMetricsRecorder creates a new MetricsRecorder.
func newMetricsRecorder(meter metric.Meter) *MetricsRecorder {
	return &MetricsRecorder{
		requestsCounter:        createInt64Counter(meter, "requests_total"),
		successCounter:         createInt64Counter(meter, "success_total"),
		failureCounter:         createInt64Counter(meter, "failure_total"),
		executionTimeHistogram: createFloat64Histogram(meter, "execution_time_seconds"),
	}
}

// createInt64Counter creates a new Int64Counter.
func createInt64Counter(meter metric.Meter, name string) metric.Int64Counter {
	counter, err := meter.Int64Counter(name)
	if err != nil {
		log.Fatalf("failed to create counter %s: %v", name, err)
	}

	return counter
}

// createFloat64Histogram creates a new Float64Histogram.
func createFloat64Histogram(meter metric.Meter, name string) metric.Float64Histogram {
	histogram, err := meter.Float64Histogram(name)
	if err != nil {
		log.Fatalf("failed to create histogram %s: %v", name, err)
	}

	return histogram
}

// StreamMetric returns a gRPC client stream interceptor that records metrics.
func StreamMetric() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		meter := otel.GetMeterProvider().Meter("grpc_client_metrics")

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return &metricClientStream{
			ClientStream: clientStream,
			metrics:      newMetricsRecorder(meter),
			method:       method,
		}, nil
	}
}

type metricClientStream struct {
	grpc.ClientStream
	metrics *MetricsRecorder
	method  string
}

func (s *metricClientStream) RecvMsg(m interface{}) error {
	start := time.Now()

	err := s.ClientStream.RecvMsg(m)

	defer s.metrics.recordMetrics(context.Background(), err, time.Since(start).Seconds()) //nolint:govet

	return nil
}

func (s *metricClientStream) SendMsg(m interface{}) error {
	start := time.Now()

	err := s.ClientStream.SendMsg(m)

	defer s.metrics.recordMetrics(context.Background(), err, time.Since(start).Seconds()) //nolint:govet

	return err
}

// recordMetrics records the metrics for a gRPC streaming call.
func (mr *MetricsRecorder) recordMetrics(ctx context.Context, err error, executionTime float64) {
	mr.requestsCounter.Add(ctx, 1)

	if err != nil {
		mr.failureCounter.Add(ctx, 1)
	} else {
		mr.successCounter.Add(ctx, 1)
	}

	mr.executionTimeHistogram.Record(ctx, executionTime)
}
