package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc/credentials"
)

var version = "dev"

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
		listDependencies(cfg)
		return
	}

	ctx := context.Background()
	shutdown, err := initOTEL(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize OTEL: %v", err)
	}
	defer shutdown(ctx)

	collector := NewCollector(cfg)
	if err := collector.Collect(ctx); err != nil {
		log.Fatalf("Failed to collect dependencies: %v", err)
	}

	fmt.Println("\nDone. Metrics published to", cfg.OTELEndpoint)
}

func listDependencies(cfg Config) {
	parser := NewParser(cfg.Directory)

	backend, err := parser.ParseBackend()
	if err != nil {
		log.Fatalf("Failed to detect backend: %v", err)
	}

	fmt.Printf("\nBackend Type:      %s\n", backend.Type)
	switch backend.Type {
	case "workspace":
		fmt.Printf("Organization:      %s\n", backend.Organization)
		fmt.Printf("Workspace:         %s\n", backend.Workspace)
	case "s3":
		fmt.Printf("S3 Bucket:         %s\n", backend.Bucket)
		fmt.Printf("S3 Key:            %s\n", backend.Key)
	}

	if err := parser.EnsureInit(); err != nil {
		log.Fatalf("terraform init failed: %v", err)
	}

	modules, err := parser.ParseModules()
	if err != nil {
		log.Fatalf("Failed to parse modules: %v", err)
	}

	providers, err := parser.ParseProviders()
	if err != nil {
		log.Fatalf("Failed to parse providers: %v", err)
	}

	if len(modules) > 0 {
		fmt.Println("\nModules:")
		for _, m := range modules {
			fmt.Printf("  %-30s %s @ %s\n", m.Name, m.Source, m.Version)
		}
	}

	if len(providers) > 0 {
		fmt.Println("\nProviders:")
		for _, p := range providers {
			fmt.Printf("  %-30s %s @ %s\n", p.Name, p.Source, p.Version)
		}
	}

	if len(modules) == 0 && len(providers) == 0 {
		fmt.Println("\nNo modules or providers found.")
	}
}

func parseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.Directory, "dir", ".", "Terraform configuration directory (default: current directory)")
	flag.StringVar(&cfg.Phase, "phase", "plan", "Terraform phase: plan or apply")
	flag.StringVar(&cfg.OTELEndpoint, "otel-endpoint", "localhost:4317", "OTEL collector endpoint")
	flag.BoolVar(&cfg.OTELInsecure, "otel-insecure", true, "Use insecure gRPC connection")
	flag.BoolVar(&cfg.ListOnly, "list", false, "List modules and providers without publishing metrics")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("tfwatch %s\n", version)
		os.Exit(0)
	}

	if cfg.Phase != "plan" && cfg.Phase != "apply" {
		fmt.Fprintln(os.Stderr, "Error: --phase must be 'plan' or 'apply'")
		flag.Usage()
		os.Exit(1)
	}

	return cfg
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
