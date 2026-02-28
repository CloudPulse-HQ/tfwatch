// Package main implements the tfwatch CLI, a tool that extracts Terraform
// dependency metadata (modules, providers, and backend configuration) and
// publishes it as OpenTelemetry metrics.
//
// Usage:
//
//	# List detected dependencies
//	tfwatch --list --dir ./infra
//
//	# Publish metrics to an OTEL collector
//	tfwatch --dir ./infra --otel-endpoint otel.example.com:4317
//
//	# Show version
//	tfwatch --version
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/CloudPulse-HQ/tfwatch/internal/tfwatch"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc/credentials"
)

var version = "dev"

// Config holds CLI flag values for a tfwatch run.
type Config struct {
	Directory    string
	Phase        string // "plan" or "apply"
	OTELEndpoint string
	OTELInsecure bool
	ListOnly     bool
}

func main() {
	cfg := parseFlags()
	printBanner()

	if cfg.ListOnly {
		if err := listDependencies(cfg); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx := context.Background()
	shutdown, err := initOTEL(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize OTEL: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	collector := tfwatch.NewCollector(tfwatch.CollectorConfig{
		Directory:    cfg.Directory,
		Phase:        cfg.Phase,
		OTELEndpoint: cfg.OTELEndpoint,
	})
	if err := collector.Collect(ctx); err != nil {
		log.Fatalf("Failed to collect dependencies: %v", err)
	}

	fmt.Println("\nDone. Metrics published to", cfg.OTELEndpoint)
}

func listDependencies(cfg Config) error {
	return tfwatch.ListDependencies(cfg.Directory)
}

func parseFlags() Config {
	cfg, exit := parseFlagsFrom(os.Args[1:])
	if exit >= 0 {
		os.Exit(exit)
	}
	return cfg
}

// parseFlagsFrom parses flags from the given args. Returns (config, exitCode).
// exitCode < 0 means continue; >= 0 means the caller should exit with that code.
func parseFlagsFrom(args []string) (Config, int) {
	var cfg Config
	fs := flag.NewFlagSet("tfwatch", flag.ContinueOnError)

	fs.StringVar(&cfg.Directory, "dir", ".", "Terraform configuration directory (default: current directory)")
	fs.StringVar(&cfg.Phase, "phase", "plan", "Terraform phase: plan or apply")
	fs.StringVar(&cfg.OTELEndpoint, "otel-endpoint", "localhost:4317", "OTEL collector endpoint")
	fs.BoolVar(&cfg.OTELInsecure, "otel-insecure", true, "Use insecure gRPC connection")
	fs.BoolVar(&cfg.ListOnly, "list", false, "List modules and providers without publishing metrics")
	showVersion := fs.Bool("version", false, "Show version")

	if err := fs.Parse(args); err != nil {
		return cfg, 1
	}

	if *showVersion {
		fmt.Printf("tfwatch %s\n", version)
		return cfg, 0
	}

	if cfg.Phase != "plan" && cfg.Phase != "apply" {
		fmt.Fprintln(os.Stderr, "Error: --phase must be 'plan' or 'apply'")
		fs.Usage()
		return cfg, 1
	}

	return cfg, -1
}

func initOTEL(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("tfwatch"),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.OTELEndpoint),
	}

	if cfg.OTELInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	} else {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}

func printBanner() {
	fmt.Printf("tfwatch %s — Terraform Dependency Tracker\n", version)
}
