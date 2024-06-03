package middleware

import (
	"context"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"log"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// MetricsRecorder records metrics for command execution.
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
func createInt64Counter(meter metric.Meter, name string) metric.Int64Counter { //nolint:ireturn
	counter, err := meter.Int64Counter(name)
	if err != nil {
		log.Fatalf("failed to create counter %s: %v", name, err)
	}

	return counter
}

// createFloat64Histogram creates a new Float64Histogram.
func createFloat64Histogram(meter metric.Meter, name string) metric.Float64Histogram { //nolint:ireturn
	histogram, err := meter.Float64Histogram(name)
	if err != nil {
		log.Fatalf("failed to create histogram %s: %v", name, err)
	}

	return histogram
}

// recordMetrics records the metrics for a command execution.
func (mr *MetricsRecorder) recordMetrics(ctx context.Context, err error, executionTime float64) {
	mr.requestsCounter.Add(ctx, 1)

	if err != nil {
		mr.failureCounter.Add(ctx, 1)
	} else {
		mr.successCounter.Add(ctx, 1)
	}

	mr.executionTimeHistogram.Record(ctx, executionTime)
}

func StreamMetric() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &metricServerStream{ServerStream: stream})
	}
}

type metricServerStream struct {
	_ struct{}
	grpc.ServerStream
}

func (s *metricServerStream) RecvMsg(m any) error {
	startTime := time.Now()

	err := s.ServerStream.RecvMsg(m)

	newMetricsRecorder(otel.GetMeterProvider().Meter("echosphere.io/receiver")).
		recordMetrics(
			context.Background(),
			err,
			time.Since(startTime).Seconds(),
		)

	return err
}

func (s *metricServerStream) SendMsg(m any) error {
	startTime := time.Now()

	err := s.ServerStream.SendMsg(m)

	newMetricsRecorder(otel.GetMeterProvider().Meter("echosphere.io/sender")).
		recordMetrics(
			context.Background(),
			err,
			time.Since(startTime).Seconds(),
		)

	return err
}
