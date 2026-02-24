package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Collector publishes Terraform dependency metrics via OpenTelemetry.
type Collector struct {
	config    Config
	gauge     metric.Int64Gauge
	tfVersion string
}

// Module represents a Terraform module dependency.
type Module struct {
	Name    string
	Source  string
	Version string
}

// Provider represents a Terraform provider dependency.
type Provider struct {
	Name    string
	Source  string
	Version string
}

// NewCollector creates a Collector with an OTEL gauge metric.
func NewCollector(cfg Config) *Collector {
	meter := otel.Meter("tfwatch")

	gauge, err := meter.Int64Gauge(
		"terraform_dependency_version",
		metric.WithDescription("Terraform module and provider dependencies (version in labels)"),
	)
	if err != nil {
		log.Fatalf("Failed to create gauge: %v", err)
	}

	tfVer := getTerraformVersion()

	return &Collector{
		config:    cfg,
		gauge:     gauge,
		tfVersion: tfVer,
	}
}

// Collect parses dependencies and publishes them as OTEL metrics.
func (c *Collector) Collect(ctx context.Context) error {
	parser := NewParser(c.config.Directory)

	backend, err := parser.ParseBackend()
	if err != nil {
		return fmt.Errorf("failed to detect backend: %w", err)
	}

	fmt.Printf("\nDirectory:         %s\n", c.config.Directory)
	fmt.Printf("Phase:             %s\n", c.config.Phase)
	fmt.Printf("Backend Type:      %s\n", backend.Type)
	switch backend.Type {
	case "workspace":
		fmt.Printf("Organization:      %s\n", backend.Organization)
		fmt.Printf("Workspace:         %s\n", backend.Workspace)
	case "s3":
		fmt.Printf("S3 Bucket:         %s\n", backend.Bucket)
		fmt.Printf("S3 Key:            %s\n", backend.Key)
	}
	fmt.Printf("Terraform Version: %s\n", c.tfVersion)
	fmt.Printf("OTEL Endpoint:     %s\n", c.config.OTELEndpoint)
	fmt.Println()

	if err := parser.EnsureInit(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	modules, err := parser.ParseModules()
	if err != nil {
		return fmt.Errorf("failed to parse modules: %w", err)
	}
	fmt.Printf("Found %d module(s)\n", len(modules))

	providers, err := parser.ParseProviders()
	if err != nil {
		return fmt.Errorf("failed to parse providers: %w", err)
	}
	fmt.Printf("Found %d provider(s)\n\n", len(providers))

	for _, mod := range modules {
		c.publishDependencyMetric(ctx, "module", mod.Name, mod.Source, mod.Version, backend)
	}

	for _, prov := range providers {
		c.publishDependencyMetric(ctx, "provider", prov.Name, prov.Source, prov.Version, backend)
	}

	return nil
}

func backendAttrs(backend *BackendConfig) []attribute.KeyValue {
	var org, ws string
	switch backend.Type {
	case "workspace":
		org = backend.Organization
		ws = backend.Workspace
	case "s3":
		org = backend.Bucket
		ws = backend.Key
	}
	return []attribute.KeyValue{
		attribute.String("backend_type", backend.Type),
		attribute.String("backend_org", org),
		attribute.String("backend_workspace", ws),
	}
}

func (c *Collector) publishDependencyMetric(ctx context.Context, depType, name, source, version string, backend *BackendConfig) {
	attrs := backendAttrs(backend)
	attrs = append(attrs,
		attribute.String("phase", c.config.Phase),
		attribute.String("type", depType),
		attribute.String("dependency_name", name),
		attribute.String("dependency_source", source),
		attribute.String("dependency_version", version),
		attribute.String("terraform_version", c.tfVersion),
	)

	c.gauge.Record(ctx, 1, metric.WithAttributes(attrs...))
	fmt.Printf("  %s: %s v%s\n", depType, name, version)
}

func getTerraformVersion() string {
	cmd := exec.Command("terraform", "version", "-json")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	var result struct {
		TerraformVersion string `json:"terraform_version"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return "unknown"
	}

	return result.TerraformVersion
}
