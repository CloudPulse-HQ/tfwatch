package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
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
