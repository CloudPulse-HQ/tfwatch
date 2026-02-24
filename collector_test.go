package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestCollector_Collect(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`
terraform {
  backend "s3" {
    bucket = "test-bucket"
    key    = "test/terraform.tfstate"
  }
}
`), 0o644)

	os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(`
provider "registry.terraform.io/hashicorp/aws" {
  version = "5.75.1"
  hashes  = ["h1:abc="]
}
`), 0o644)

	modDir := filepath.Join(dir, ".terraform", "modules")
	os.MkdirAll(modDir, 0o755)
	os.WriteFile(filepath.Join(modDir, "modules.json"), []byte(`{"Modules":[
		{"Key":"","Source":"","Dir":"."},
		{"Key":"vpc","Source":"registry.terraform.io/terraform-aws-modules/vpc/aws","Version":"5.1.2","Dir":".terraform/modules/vpc"}
	]}`), 0o644)

	reader := setupTestMeter(t)

	cfg := Config{
		Directory:    dir,
		Phase:        "plan",
		OTELEndpoint: "localhost:4317",
	}

	collector := NewCollector(cfg)
	ctx := context.Background()
	if err := collector.Collect(ctx); err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	var found int
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "terraform_dependency_version" {
				gauge, ok := m.Data.(metricdata.Gauge[int64])
				if !ok {
					t.Fatal("expected Gauge[int64]")
				}
				found = len(gauge.DataPoints)
			}
		}
	}

	// Expect 2 data points: 1 module (vpc) + 1 provider (aws)
	if found != 2 {
		t.Errorf("expected 2 metric data points, got %d", found)
	}
}

func TestCollector_Collect_NoBackend(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`resource "null_resource" "test" {}`), 0o644)

	setupTestMeter(t)

	cfg := Config{Directory: dir, Phase: "plan"}
	collector := NewCollector(cfg)

	err := collector.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing backend, got nil")
	}
}

func TestGetTerraformVersion(t *testing.T) {
	ver := getTerraformVersion()
	if ver == "" {
		t.Error("expected non-empty version string")
	}
}

func setupTestMeter(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	t.Cleanup(func() { provider.Shutdown(context.Background()) })
	return reader
}

func assertAttrs(t *testing.T, attrs []attribute.KeyValue, expected map[string]string) {
	t.Helper()
	if len(attrs) != len(expected) {
		t.Errorf("expected %d attrs, got %d", len(expected), len(attrs))
		return
	}
	for _, kv := range attrs {
		key := string(kv.Key)
		want, ok := expected[key]
		if !ok {
			t.Errorf("unexpected attr key: %s", key)
			continue
		}
		got := kv.Value.AsString()
		if got != want {
			t.Errorf("attr %s: expected %q, got %q", key, want, got)
		}
	}
}
