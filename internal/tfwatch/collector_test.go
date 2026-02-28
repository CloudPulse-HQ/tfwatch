package tfwatch

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// setupExampleDir creates a minimal Terraform directory for testing.
func setupExampleDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`
terraform {
  cloud {
    organization = "test-org"
    workspaces { name = "test-ws" }
  }
}
`), 0o644)

	os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(`
provider "registry.terraform.io/hashicorp/aws" {
  version = "5.75.1"
  hashes  = ["h1:abc="]
}
provider "registry.terraform.io/hashicorp/null" {
  version = "3.2.3"
  hashes  = ["h1:xyz="]
}
`), 0o644)

	modDir := filepath.Join(dir, ".terraform", "modules")
	os.MkdirAll(modDir, 0o755)
	os.WriteFile(filepath.Join(modDir, "modules.json"), []byte(`{"Modules":[
		{"Key":"","Source":"","Dir":"."},
		{"Key":"vpc","Source":"registry.terraform.io/terraform-aws-modules/vpc/aws","Version":"5.1.2","Dir":".terraform/modules/vpc"},
		{"Key":"eks","Source":"registry.terraform.io/terraform-aws-modules/eks/aws","Version":"20.5.0","Dir":".terraform/modules/eks"}
	]}`), 0o644)

	return dir
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

	cfg := CollectorConfig{
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

	cfg := CollectorConfig{Directory: dir, Phase: "plan"}
	collector := NewCollector(cfg)

	err := collector.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing backend, got nil")
	}
}

func TestCollector_Collect_Workspace(t *testing.T) {
	dir := setupExampleDir(t)

	reader := setupTestMeter(t)
	_ = reader

	cfg := CollectorConfig{
		Directory:    dir,
		Phase:        "apply",
		OTELEndpoint: "localhost:4317",
	}

	collector := NewCollector(cfg)
	output := captureStdout(func() {
		if err := collector.Collect(context.Background()); err != nil {
			t.Fatalf("Collect() error: %v", err)
		}
	})

	checks := []string{
		"Backend Type:      workspace",
		"Organization:      test-org",
		"Workspace:         test-ws",
		"module: vpc v5.1.2",
		"module: eks v20.5.0",
		"provider: aws v5.75.1",
		"provider: null v3.2.3",
		"Found 2 module(s)",
		"Found 2 provider(s)",
	}
	for _, check := range checks {
		if !bytes.Contains([]byte(output), []byte(check)) {
			t.Errorf("output missing %q", check)
		}
	}
}

func TestGetTerraformVersion(t *testing.T) {
	ver := getTerraformVersion()
	if ver == "" {
		t.Error("expected non-empty version string")
	}
}

func TestBackendAttrs(t *testing.T) {
	tests := []struct {
		name     string
		backend  *BackendConfig
		expected map[string]string
	}{
		{
			name: "workspace",
			backend: &BackendConfig{
				Type:         "workspace",
				Organization: "acme-corp",
				Workspace:    "production",
			},
			expected: map[string]string{
				"backend_type":      "workspace",
				"backend_org":       "acme-corp",
				"backend_workspace": "production",
			},
		},
		{
			name: "s3",
			backend: &BackendConfig{
				Type:   "s3",
				Bucket: "my-state-bucket",
				Key:    "prod_vpc_terraform.tfstate",
			},
			expected: map[string]string{
				"backend_type":      "s3",
				"backend_org":       "my-state-bucket",
				"backend_workspace": "prod_vpc_terraform.tfstate",
			},
		},
		{
			name: "unknown type",
			backend: &BackendConfig{
				Type: "local",
			},
			expected: map[string]string{
				"backend_type":      "local",
				"backend_org":       "",
				"backend_workspace": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := backendAttrs(tt.backend)
			assertAttrs(t, attrs, tt.expected)
		})
	}
}

func TestListDependencies(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns dir
		wantErr bool
		checks  []string // substrings expected in stdout
	}{
		{
			name: "no backend",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`resource "null" "a" {}`), 0o644)
				return dir
			},
			wantErr: true,
		},
		{
			name: "cloud backend",
			setup: func(t *testing.T) string {
				t.Helper()
				return setupExampleDir(t)
			},
			checks: []string{
				"Backend Type:      workspace",
				"Organization:      test-org",
				"Workspace:         test-ws",
				"Modules:",
				"vpc",
				"eks",
				"Providers:",
				"aws",
				"null",
			},
		},
		{
			name: "s3 backend",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`
terraform {
  backend "s3" {
    bucket = "my-bucket"
    key    = "env/prod/terraform.tfstate"
  }
}
`), 0o644)
				os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(`
provider "registry.terraform.io/hashicorp/aws" {
  version = "5.50.0"
  hashes  = ["h1:abc="]
}
`), 0o644)
				modDir := filepath.Join(dir, ".terraform", "modules")
				os.MkdirAll(modDir, 0o755)
				os.WriteFile(filepath.Join(modDir, "modules.json"), []byte(`{"Modules":[{"Key":"","Source":"","Dir":"."}]}`), 0o644)
				return dir
			},
			checks: []string{
				"Backend Type:      s3",
				"S3 Bucket:         my-bucket",
				"S3 Key:            env_prod_terraform.tfstate",
				"Providers:",
				"aws",
			},
		},
		{
			name: "no dependencies",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`
terraform {
  backend "s3" {
    bucket = "empty-bucket"
    key    = "empty.tfstate"
  }
}
`), 0o644)
				os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0o644)
				modDir := filepath.Join(dir, ".terraform", "modules")
				os.MkdirAll(modDir, 0o755)
				os.WriteFile(filepath.Join(modDir, "modules.json"), []byte(`{"Modules":[{"Key":"","Source":"","Dir":"."}]}`), 0o644)
				return dir
			},
			checks: []string{"No modules or providers found."},
		},
		{
			name: "invalid modules.json",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`
terraform {
  backend "s3" { bucket = "b"; key = "k" }
}
`), 0o644)
				os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0o644)
				modDir := filepath.Join(dir, ".terraform", "modules")
				os.MkdirAll(modDir, 0o755)
				os.WriteFile(filepath.Join(modDir, "modules.json"), []byte("{bad json}"), 0o644)
				return dir
			},
			wantErr: true,
		},
		{
			name: "invalid providers lock",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`
terraform {
  backend "s3" { bucket = "b"; key = "k" }
}
`), 0o644)
				os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte("{{{ invalid"), 0o644)
				modDir := filepath.Join(dir, ".terraform", "modules")
				os.MkdirAll(modDir, 0o755)
				os.WriteFile(filepath.Join(modDir, "modules.json"), []byte(`{"Modules":[{"Key":"","Source":"","Dir":"."}]}`), 0o644)
				return dir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)

			var output string
			var err error
			output = captureStdout(func() {
				err = ListDependencies(dir)
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, check := range tt.checks {
				if !bytes.Contains([]byte(output), []byte(check)) {
					t.Errorf("output missing %q", check)
				}
			}
		})
	}
}
