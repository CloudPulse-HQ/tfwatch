package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseModules(t *testing.T) {
	dir := t.TempDir()

	// Create .terraform/modules/ directory
	modDir := filepath.Join(dir, ".terraform", "modules")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `{"Modules":[
		{"Key":"","Source":"","Dir":"."},
		{"Key":"s3_bucket","Source":"registry.terraform.io/terraform-aws-modules/s3-bucket/aws","Version":"5.10.0","Dir":".terraform/modules/s3_bucket"},
		{"Key":"vpc","Source":"registry.terraform.io/terraform-aws-modules/vpc/aws","Version":"5.1.0","Dir":".terraform/modules/vpc"}
	]}`

	if err := os.WriteFile(filepath.Join(modDir, "modules.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(dir)
	modules, err := p.ParseModules()
	if err != nil {
		t.Fatalf("ParseModules() error: %v", err)
	}

	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}

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
}

func TestParseModules_MissingFile(t *testing.T) {
	p := NewParser(t.TempDir())
	modules, err := p.ParseModules()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if modules != nil {
		t.Fatalf("expected nil slice, got %v", modules)
	}
}

func TestParseProviders(t *testing.T) {
	dir := t.TempDir()

	content := `
provider "registry.terraform.io/hashicorp/aws" {
  version     = "6.32.1"
  constraints = ">= 6.28.0"
  hashes = [
    "h1:abc123=",
  ]
}

provider "registry.terraform.io/hashicorp/random" {
  version = "3.6.0"
  hashes = [
    "h1:xyz789=",
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(dir)
	providers, err := p.ParseProviders()
	if err != nil {
		t.Fatalf("ParseProviders() error: %v", err)
	}

	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

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
}

func TestParseProviders_MissingFile(t *testing.T) {
	p := NewParser(t.TempDir())
	providers, err := p.ParseProviders()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if providers != nil {
		t.Fatalf("expected nil slice, got %v", providers)
	}
}

func TestParseBackend_Cloud(t *testing.T) {
	dir := t.TempDir()

	content := `
terraform {
  cloud {
    organization = "myorg"
    workspaces {
      name = "prod"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "versions.tf"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(dir)
	cfg, err := p.ParseBackend()
	if err != nil {
		t.Fatalf("ParseBackend() error: %v", err)
	}

	if cfg.Type != "workspace" {
		t.Errorf("expected type workspace, got %s", cfg.Type)
	}
	if cfg.Organization != "myorg" {
		t.Errorf("expected organization myorg, got %s", cfg.Organization)
	}
	if cfg.Workspace != "prod" {
		t.Errorf("expected workspace prod, got %s", cfg.Workspace)
	}
}

func TestParseBackend_S3(t *testing.T) {
	dir := t.TempDir()

	content := `
terraform {
  backend "s3" {
    bucket = "mybucket"
    key    = "terraform.tfstate"
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "versions.tf"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(dir)
	cfg, err := p.ParseBackend()
	if err != nil {
		t.Fatalf("ParseBackend() error: %v", err)
	}

	if cfg.Type != "s3" {
		t.Errorf("expected type s3, got %s", cfg.Type)
	}
	if cfg.Bucket != "mybucket" {
		t.Errorf("expected bucket mybucket, got %s", cfg.Bucket)
	}
	if cfg.Key != "terraform.tfstate" {
		t.Errorf("expected key terraform.tfstate, got %s", cfg.Key)
	}
}

func TestParseBackend_S3KeyNormalization(t *testing.T) {
	dir := t.TempDir()

	content := `
terraform {
  backend "s3" {
    bucket = "mybucket"
    key    = "path/to/terraform.tfstate"
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "versions.tf"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(dir)
	cfg, err := p.ParseBackend()
	if err != nil {
		t.Fatalf("ParseBackend() error: %v", err)
	}

	if cfg.Key != "path_to_terraform.tfstate" {
		t.Errorf("expected key path_to_terraform.tfstate, got %s", cfg.Key)
	}
}

func TestParseBackend_NoBackend(t *testing.T) {
	dir := t.TempDir()

	content := `
terraform {
  required_version = ">= 1.0"
}
`
	if err := os.WriteFile(filepath.Join(dir, "versions.tf"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(dir)
	_, err := p.ParseBackend()
	if err == nil {
		t.Fatal("expected error for missing backend, got nil")
	}
}
