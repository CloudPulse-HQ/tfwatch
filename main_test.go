package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
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

func TestParseFlagsFrom(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int // expected exit code; -1 means continue
		checks   func(t *testing.T, cfg Config)
	}{
		{
			name:     "defaults",
			args:     []string{},
			wantExit: -1,
			checks: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.Directory != "." {
					t.Errorf("expected dir '.', got %q", cfg.Directory)
				}
				if cfg.Phase != "plan" {
					t.Errorf("expected phase 'plan', got %q", cfg.Phase)
				}
				if cfg.OTELEndpoint != "localhost:4317" {
					t.Errorf("expected endpoint 'localhost:4317', got %q", cfg.OTELEndpoint)
				}
				if !cfg.OTELInsecure {
					t.Error("expected otel-insecure=true")
				}
				if cfg.ListOnly {
					t.Error("expected list=false")
				}
			},
		},
		{
			name: "all flags",
			args: []string{
				"--dir", "/tmp/infra",
				"--phase", "apply",
				"--otel-endpoint", "otel.example.com:4317",
				"--otel-insecure=false",
				"--list",
			},
			wantExit: -1,
			checks: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.Directory != "/tmp/infra" {
					t.Errorf("expected dir '/tmp/infra', got %q", cfg.Directory)
				}
				if cfg.Phase != "apply" {
					t.Errorf("expected phase 'apply', got %q", cfg.Phase)
				}
				if cfg.OTELEndpoint != "otel.example.com:4317" {
					t.Errorf("expected endpoint 'otel.example.com:4317', got %q", cfg.OTELEndpoint)
				}
				if cfg.OTELInsecure {
					t.Error("expected otel-insecure=false")
				}
				if !cfg.ListOnly {
					t.Error("expected list=true")
				}
			},
		},
		{
			name:     "version flag",
			args:     []string{"--version"},
			wantExit: 0,
		},
		{
			name:     "invalid phase",
			args:     []string{"--phase", "destroy"},
			wantExit: 1,
		},
		{
			name:     "unknown flag",
			args:     []string{"--unknown-flag"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			var exit int

			captureStdout(func() {
				cfg, exit = parseFlagsFrom(tt.args)
			})

			if exit != tt.wantExit {
				t.Fatalf("expected exit %d, got %d", tt.wantExit, exit)
			}
			if tt.checks != nil {
				tt.checks(t, cfg)
			}
		})
	}
}

func TestPrintBanner(t *testing.T) {
	output := captureStdout(func() {
		printBanner()
	})
	expected := fmt.Sprintf("tfwatch %s — Terraform Dependency Tracker\n", version)
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestInitOTEL(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
	}{
		{"insecure", true},
		{"TLS", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				OTELEndpoint: "localhost:4317",
				OTELInsecure: tt.insecure,
			}

			ctx := context.Background()
			shutdown, err := initOTEL(ctx, cfg)
			if err != nil {
				t.Fatalf("initOTEL() error: %v", err)
			}
			if shutdown == nil {
				t.Fatal("expected non-nil shutdown function")
			}
			shutdown(ctx)
		})
	}
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
			cfg := Config{Directory: dir}

			var output string
			var err error
			output = captureStdout(func() {
				err = listDependencies(cfg)
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

func TestCollector_Collect_Workspace(t *testing.T) {
	dir := setupExampleDir(t)

	reader := setupTestMeter(t)
	_ = reader

	cfg := Config{
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
