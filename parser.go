package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type BackendConfig struct {
	Type         string // "workspace" or "s3"
	Organization string // cloud backend: tf_org
	Workspace    string // cloud backend: workspace name
	Bucket       string // s3 backend: bucket name
	Key          string // s3 backend: normalized key (slashes → underscores)
}

type Parser struct {
	directory string
}

func NewParser(directory string) *Parser {
	return &Parser{directory: directory}
}

// modulesJSON matches the structure of .terraform/modules/modules.json
type modulesJSON struct {
	Modules []moduleEntry `json:"Modules"`
}

type moduleEntry struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version"`
	Dir     string `json:"Dir"`
}

// needsInit checks whether terraform init has been run by looking for generated files.
func (p *Parser) needsInit() bool {
	lockFile := filepath.Join(p.directory, ".terraform.lock.hcl")
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		return true
	}
	return false
}

// runInit runs terraform init in the configured directory.
func (p *Parser) runInit() error {
	fmt.Printf("Running terraform init in %s...\n", p.directory)
	cmd := exec.Command("terraform", "init")
	cmd.Dir = p.directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnsureInit runs terraform init if generated files are missing.
func (p *Parser) EnsureInit() error {
	if !p.needsInit() {
		return nil
	}
	return p.runInit()
}

// ParseModules reads .terraform/modules/modules.json (created by terraform init)
// and returns all non-root module entries. Returns empty slice if the file doesn't exist.
func (p *Parser) ParseModules() ([]Module, error) {
	path := filepath.Join(p.directory, ".terraform", "modules", "modules.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Warning: %s not found", path)
			return nil, nil
		}
		return nil, err
	}

	var mj modulesJSON
	if err := json.Unmarshal(data, &mj); err != nil {
		return nil, err
	}

	var modules []Module
	for _, entry := range mj.Modules {
		// Skip root entry
		if entry.Key == "" {
			continue
		}
		modules = append(modules, Module{
			Name:    entry.Key,
			Source:  entry.Source,
			Version: entry.Version,
		})
	}

	return modules, nil
}

// ParseProviders reads .terraform.lock.hcl (created by terraform init)
// and extracts provider source + version using the HCL parser.
// Returns empty slice if the file doesn't exist.
func (p *Parser) ParseProviders() ([]Provider, error) {
	path := filepath.Join(p.directory, ".terraform.lock.hcl")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Warning: %s not found", path)
			return nil, nil
		}
		return nil, err
	}

	parser := hclparse.NewParser()
	f, diag := parser.ParseHCL(data, path)
	if diag.HasErrors() {
		return nil, diag
	}

	content, _, diag := f.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "provider", LabelNames: []string{"source"}},
		},
	})
	if diag.HasErrors() {
		return nil, diag
	}

	var providers []Provider
	for _, block := range content.Blocks {
		if len(block.Labels) == 0 {
			continue
		}

		source := block.Labels[0]

		attrs, _ := block.Body.JustAttributes()
		var version string
		if v, ok := attrs["version"]; ok {
			val, _ := v.Expr.Value(nil)
			version = val.AsString()
		}

		// Derive short name from source path
		parts := filepath.Base(source)

		providers = append(providers, Provider{
			Name:    parts,
			Source:  source,
			Version: version,
		})
	}

	return providers, nil
}

// ParseBackend scans *.tf files in the directory for terraform {} blocks
// and auto-detects the backend type (cloud or s3).
func (p *Parser) ParseBackend() (*BackendConfig, error) {
	files, err := filepath.Glob(filepath.Join(p.directory, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob tf files: %w", err)
	}

	parser := hclparse.NewParser()

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		f, diag := parser.ParseHCL(data, file)
		if diag.HasErrors() {
			continue
		}

		content, _, diag := f.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "terraform"},
			},
		})
		if diag.HasErrors() || len(content.Blocks) == 0 {
			continue
		}

		for _, tfBlock := range content.Blocks {
			if cfg := p.parseCloudBlock(tfBlock); cfg != nil {
				return cfg, nil
			}
			if cfg := p.parseS3Backend(tfBlock); cfg != nil {
				return cfg, nil
			}
		}
	}

	return nil, fmt.Errorf("no backend configuration found in %s", p.directory)
}

func (p *Parser) parseCloudBlock(tfBlock *hcl.Block) *BackendConfig {
	content, _, diag := tfBlock.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "cloud"},
		},
	})
	if diag.HasErrors() || len(content.Blocks) == 0 {
		return nil
	}

	cloudBlock := content.Blocks[0]
	attrs, _ := cloudBlock.Body.JustAttributes()

	var org string
	if v, ok := attrs["organization"]; ok {
		val, _ := v.Expr.Value(nil)
		org = val.AsString()
	}

	// Parse workspaces {} nested block
	var workspace string
	wsContent, _, _ := cloudBlock.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "workspaces"},
		},
	})
	if wsContent != nil && len(wsContent.Blocks) > 0 {
		wsAttrs, _ := wsContent.Blocks[0].Body.JustAttributes()
		if v, ok := wsAttrs["name"]; ok {
			val, _ := v.Expr.Value(nil)
			workspace = val.AsString()
		}
	}

	return &BackendConfig{
		Type:         "workspace",
		Organization: org,
		Workspace:    workspace,
	}
}

func (p *Parser) parseS3Backend(tfBlock *hcl.Block) *BackendConfig {
	content, _, diag := tfBlock.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "backend", LabelNames: []string{"type"}},
		},
	})
	if diag.HasErrors() || len(content.Blocks) == 0 {
		return nil
	}

	for _, block := range content.Blocks {
		if len(block.Labels) == 0 || block.Labels[0] != "s3" {
			continue
		}

		attrs, _ := block.Body.JustAttributes()

		var bucket, key string
		if v, ok := attrs["bucket"]; ok {
			val, _ := v.Expr.Value(nil)
			bucket = val.AsString()
		}
		if v, ok := attrs["key"]; ok {
			val, _ := v.Expr.Value(nil)
			key = strings.ReplaceAll(val.AsString(), "/", "_")
		}

		return &BackendConfig{
			Type:   "s3",
			Bucket: bucket,
			Key:    key,
		}
	}

	return nil
}
