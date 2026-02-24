package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseModules(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string // testdata filename, or "" for no setup
		wantCount int
		wantErr   bool
		checks    func(t *testing.T, modules []Module)
	}{
		{
			name:      "two modules",
			fixture:   "testdata/modules.json",
			wantCount: 2,
			checks: func(t *testing.T, modules []Module) {
				t.Helper()
				if modules[0].Name != "s3_bucket" {
					t.Errorf("expected name s3_bucket, got %s", modules[0].Name)
				}
				if modules[0].Source != "registry.terraform.io/terraform-aws-modules/s3-bucket/aws" {
					t.Errorf("unexpected source: %s", modules[0].Source)
				}
				if modules[0].Version != "5.10.0" {
					t.Errorf("expected version 5.10.0, got %s", modules[0].Version)
				}
				if modules[1].Name != "vpc" {
					t.Errorf("expected name vpc, got %s", modules[1].Name)
				}
				if modules[1].Version != "5.1.0" {
					t.Errorf("expected version 5.1.0, got %s", modules[1].Version)
				}
			},
		},
		{
			name:      "empty modules (root only)",
			fixture:   "testdata/modules_empty.json",
			wantCount: 0,
		},
		{
			name:    "invalid JSON",
			fixture: "testdata/modules_invalid.json",
			wantErr: true,
		},
		{
			name:      "missing file",
			fixture:   "", // no file created
			wantCount: 0,
			checks: func(t *testing.T, modules []Module) {
				t.Helper()
				if modules != nil {
					t.Fatalf("expected nil slice, got %v", modules)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.fixture != "" {
				modDir := filepath.Join(dir, ".terraform", "modules")
				if err := os.MkdirAll(modDir, 0o755); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(tt.fixture)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(modDir, "modules.json"), data, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			p := NewParser(dir)
			modules, err := p.ParseModules()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(modules) != tt.wantCount {
				t.Fatalf("expected %d modules, got %d", tt.wantCount, len(modules))
			}
			if tt.checks != nil {
				tt.checks(t, modules)
			}
		})
	}
}

func TestParseProviders(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string // testdata filename for .terraform.lock.hcl, or "" for missing
		wantCount int
		wantErr   bool
		checks    func(t *testing.T, providers []Provider)
	}{
		{
			name:      "two providers",
			fixture:   "testdata/lock_two_providers.hcl",
			wantCount: 2,
			checks: func(t *testing.T, providers []Provider) {
				t.Helper()
				if providers[0].Name != "aws" {
					t.Errorf("expected name aws, got %s", providers[0].Name)
				}
				if providers[0].Source != "registry.terraform.io/hashicorp/aws" {
					t.Errorf("unexpected source: %s", providers[0].Source)
				}
				if providers[0].Version != "6.32.1" {
					t.Errorf("expected version 6.32.1, got %s", providers[0].Version)
				}
				if providers[1].Name != "random" {
					t.Errorf("expected name random, got %s", providers[1].Name)
				}
				if providers[1].Version != "3.6.0" {
					t.Errorf("expected version 3.6.0, got %s", providers[1].Version)
				}
			},
		},
		{
			name:      "no version attribute",
			fixture:   "testdata/lock_no_version.hcl",
			wantCount: 1,
			checks: func(t *testing.T, providers []Provider) {
				t.Helper()
				if providers[0].Version != "" {
					t.Errorf("expected empty version, got %s", providers[0].Version)
				}
			},
		},
		{
			name:    "invalid HCL",
			fixture: "testdata/lock_invalid.hcl",
			wantErr: true,
		},
		{
			name:      "empty lock file",
			fixture:   "", // will create empty file
			wantCount: 0,
		},
		{
			name:      "missing file",
			fixture:   "MISSING", // sentinel: don't create any file
			wantCount: 0,
			checks: func(t *testing.T, providers []Provider) {
				t.Helper()
				if providers != nil {
					t.Fatalf("expected nil slice, got %v", providers)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			switch {
			case tt.fixture == "MISSING":
				// don't create any file
			case tt.fixture == "":
				// empty lock file
				if err := os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			default:
				data, err := os.ReadFile(tt.fixture)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), data, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			p := NewParser(dir)
			providers, err := p.ParseProviders()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(providers) != tt.wantCount {
				t.Fatalf("expected %d providers, got %d", tt.wantCount, len(providers))
			}
			if tt.checks != nil {
				tt.checks(t, providers)
			}
		})
	}
}

func TestParseBackend(t *testing.T) {
	tests := []struct {
		name    string
		fixture string // testdata .tf file
		wantErr bool
		checks  func(t *testing.T, cfg *BackendConfig)
	}{
		{
			name:    "cloud backend",
			fixture: "testdata/backend_cloud.tf",
			checks: func(t *testing.T, cfg *BackendConfig) {
				t.Helper()
				if cfg.Type != "workspace" {
					t.Errorf("expected type workspace, got %s", cfg.Type)
				}
				if cfg.Organization != "myorg" {
					t.Errorf("expected organization myorg, got %s", cfg.Organization)
				}
				if cfg.Workspace != "prod" {
					t.Errorf("expected workspace prod, got %s", cfg.Workspace)
				}
			},
		},
		{
			name:    "s3 backend",
			fixture: "testdata/backend_s3.tf",
			checks: func(t *testing.T, cfg *BackendConfig) {
				t.Helper()
				if cfg.Type != "s3" {
					t.Errorf("expected type s3, got %s", cfg.Type)
				}
				if cfg.Bucket != "mybucket" {
					t.Errorf("expected bucket mybucket, got %s", cfg.Bucket)
				}
				if cfg.Key != "terraform.tfstate" {
					t.Errorf("expected key terraform.tfstate, got %s", cfg.Key)
				}
			},
		},
		{
			name:    "s3 key normalization (slashes to underscores)",
			fixture: "testdata/backend_s3_slashes.tf",
			checks: func(t *testing.T, cfg *BackendConfig) {
				t.Helper()
				if cfg.Key != "path_to_terraform.tfstate" {
					t.Errorf("expected key path_to_terraform.tfstate, got %s", cfg.Key)
				}
			},
		},
		{
			name:    "cloud without workspaces block",
			fixture: "testdata/backend_cloud_no_ws.tf",
			checks: func(t *testing.T, cfg *BackendConfig) {
				t.Helper()
				if cfg.Organization != "solo-org" {
					t.Errorf("expected org 'solo-org', got %q", cfg.Organization)
				}
				if cfg.Workspace != "" {
					t.Errorf("expected empty workspace, got %q", cfg.Workspace)
				}
			},
		},
		{
			name:    "unsupported backend (gcs)",
			fixture: "testdata/backend_gcs.tf",
			wantErr: true,
		},
		{
			name:    "no backend block",
			fixture: "testdata/backend_none.tf",
			wantErr: true,
		},
		{
			name:    "empty directory",
			fixture: "", // no tf file
			wantErr: true,
		},
		{
			name:    "multiple tf files (backend in second)",
			fixture: "MULTI", // sentinel handled below
			checks: func(t *testing.T, cfg *BackendConfig) {
				t.Helper()
				if cfg.Organization != "multi-file-org" {
					t.Errorf("expected org 'multi-file-org', got %q", cfg.Organization)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			switch {
			case tt.fixture == "":
				// empty dir, no files
			case tt.fixture == "MULTI":
				os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`resource "null_resource" "a" {}`), 0o644)
				os.WriteFile(filepath.Join(dir, "versions.tf"), []byte(`
terraform {
  cloud {
    organization = "multi-file-org"
    workspaces { name = "staging" }
  }
}
`), 0o644)
			default:
				data, err := os.ReadFile(tt.fixture)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "versions.tf"), data, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			p := NewParser(dir)
			cfg, err := p.ParseBackend()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checks != nil {
				tt.checks(t, cfg)
			}
		})
	}
}

func TestNeedsInit(t *testing.T) {
	t.Run("lock file exists", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0o644)
		p := NewParser(dir)
		if p.needsInit() {
			t.Error("expected needsInit=false when lock file exists")
		}
	})

	t.Run("no lock file", func(t *testing.T) {
		p := NewParser(t.TempDir())
		if !p.needsInit() {
			t.Error("expected needsInit=true when lock file missing")
		}
	})
}

func TestEnsureInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0o644)
	p := NewParser(dir)
	if err := p.EnsureInit(); err != nil {
		t.Errorf("EnsureInit() should succeed when already initialized, got: %v", err)
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
