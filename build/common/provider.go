package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"time"
)

type HTTPSideCarServer struct {
	*http.Server
}

func (h HTTPSideCarServer) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := h.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error { <-ctx.Done(); return h.Shutdown(context.Background()) })

	return g.Wait()
}

func InitTracing(ctx context.Context, name string) func() {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
		)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)

	return func() {
		_ = tracerProvider.Shutdown(context.Background()) //nolint:errcheck
	}
}

type GenericConfig interface {
	GetSideCar() struct {
		Enabled bool
		Port    int
	}
}

func ProvideHTTPSideCar[T GenericConfig](i do.Injector) (HTTPSideCarServer, error) {
	cfg := do.MustInvoke[T](i)

	mux := http.NewServeMux()
	attachPprof(mux)
	attachPrometheus(mux)
	attachHealz(mux, i)

	return HTTPSideCarServer{
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.GetSideCar().Port),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

// attachPprof creates a pprof endpoints
func attachPprof(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func attachPrometheus(mux *http.ServeMux) {
	os.Setenv("OTEL_SERVICE_NAME", "echosphere.io")

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	otel.SetMeterProvider(metric.NewMeterProvider(metric.WithReader(exporter)))

	mux.Handle("/metrics", promhttp.Handler())
}

func attachHealz(mux *http.ServeMux, i do.Injector) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		checks := i.HealthCheck()

		for dep, err := range checks {
			if err != nil {
				http.Error(w, fmt.Sprintf("dependency: %s:%v", dep, err), http.StatusInternalServerError)
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}

	mux.HandleFunc("/health/live", handler)
	mux.HandleFunc("/health/ready", handler)
}
