// Package config provides configuration loading and parsing for the evaluation tool.
package config

import (
"bytes"
"fmt"
"log/slog"
"os"
"os/exec"
"path/filepath"

"gopkg.in/yaml.v3"
)

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	Type    string   `yaml:"type" json:"type"`
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args" json:"args"`
	Tools   []string `yaml:"tools" json:"tools"`
}

// Skill represents a unified skill entry supporting both local and remote sources.
//   - type: local  → Path contains a local directory path (supports globs)
//   - type: remote → Name + Repo identify a skill to fetch from a GitHub repo
type Skill struct {
	Type string `yaml:"type" json:"type"`
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Repo string `yaml:"repo,omitempty" json:"repo,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// GeneratorConfig holds all configuration for the code generation agent.
type GeneratorConfig struct {
	Model          string                `yaml:"model" json:"model"`
	Skills         []Skill               `yaml:"skills,omitempty" json:"skills,omitempty"`
	MCPServers     map[string]*MCPServer `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	AvailableTools []string              `yaml:"available_tools,omitempty" json:"available_tools,omitempty"`
	ExcludedTools  []string              `yaml:"excluded_tools,omitempty" json:"excluded_tools,omitempty"`
}

// ReviewerConfig holds all configuration for the review/grading plane.
type ReviewerConfig struct {
	Model  string  `yaml:"model,omitempty" json:"model,omitempty"`
	Models []string `yaml:"models,omitempty" json:"models,omitempty"`
	Skills []Skill  `yaml:"skills,omitempty" json:"skills,omitempty"`
}

// ToolConfig represents a single evaluation configuration.
type ToolConfig struct {
	Name        string           `yaml:"name" json:"name"`
	Description string           `yaml:"description" json:"description"`
	Generator   *GeneratorConfig `yaml:"generator,omitempty" json:"generator,omitempty"`
	Reviewer    *ReviewerConfig  `yaml:"reviewer,omitempty" json:"reviewer,omitempty"`
	Plugins     []string         `yaml:"plugins,omitempty" json:"plugins,omitempty"`
}

// ConfigFile represents the top-level config file structure.
type ConfigFile struct {
Configs []ToolConfig `yaml:"configs"`
}

// Load reads and parses a configuration file from the given path.
func Load(path string) (*ConfigFile, error) {
slog.Debug("Loading config file", "path", path)
data, err := os.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("reading config file: %w", err)
}
return Parse(data)
}

// LoadDir reads all .yaml files in a directory and merges their configs.
// This allows splitting configs across multiple files (e.g., baseline.yaml, azure-mcp.yaml).
func LoadDir(dir string) (*ConfigFile, error) {
slog.Debug("Loading config directory", "dir", dir)
entries, err := os.ReadDir(dir)
if err != nil {
return nil, fmt.Errorf("reading config directory %s: %w", dir, err)
}

merged := &ConfigFile{}
for _, e := range entries {
if e.IsDir() || (filepath.Ext(e.Name()) != ".yaml" && filepath.Ext(e.Name()) != ".yml") {
continue
}
cf, err := Load(filepath.Join(dir, e.Name()))
if err != nil {
return nil, fmt.Errorf("loading %s: %w", e.Name(), err)
}
merged.Configs = append(merged.Configs, cf.Configs...)
}

if len(merged.Configs) == 0 {
return nil, fmt.Errorf("no configs found in %s", dir)
}
return merged, nil
}

// Parse parses configuration from YAML bytes.
func Parse(data []byte) (*ConfigFile, error) {
	var cfg ConfigFile
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}
	for _, c := range cfg.Configs {
		slog.Info("Config loaded", "name", c.Name, "model", c.Generator.Model)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks all configs for required fields and constraint violations.
func (cf *ConfigFile) Validate() error {
	if len(cf.Configs) == 0 {
		return fmt.Errorf("no configs defined")
	}
	for i, c := range cf.Configs {
		if c.Name == "" {
			return fmt.Errorf("config at index %d has no name", i)
		}
		// Generator with a model is required for every config.
		if c.Generator == nil || c.Generator.Model == "" {
			return fmt.Errorf("config %q: generator.model is required", c.Name)
		}
		// Validate generator/reviewer skills have correct type
		if c.Generator != nil {
			for _, s := range c.Generator.Skills {
				if err := validateSkill(s); err != nil {
					return fmt.Errorf("config %q generator skill: %w", c.Name, err)
				}
			}
		}
		if c.Reviewer != nil {
			for _, s := range c.Reviewer.Skills {
				if err := validateSkill(s); err != nil {
					return fmt.Errorf("config %q reviewer skill: %w", c.Name, err)
				}
			}
			// Check for duplicate reviewer models
			seen := make(map[string]bool, len(c.Reviewer.Models))
			for _, rm := range c.Reviewer.Models {
				if seen[rm] {
					return fmt.Errorf("config %q: duplicate reviewer model %q", c.Name, rm)
				}
				seen[rm] = true
			}
		}
	}
	return nil
}

// validateSkill checks that a Skill has valid type and required fields.
func validateSkill(s Skill) error {
	switch s.Type {
	case "local":
		if s.Path == "" {
			return fmt.Errorf("local skill missing path")
		}
	case "remote":
		if s.Repo == "" {
			return fmt.Errorf("remote skill missing repo")
		}
	default:
		return fmt.Errorf("unknown skill type %q (expected \"local\" or \"remote\")", s.Type)
	}
	return nil
}

// GetConfig returns a config by name, or an error if not found.
func (cf *ConfigFile) GetConfig(name string) (*ToolConfig, error) {
for i := range cf.Configs {
if cf.Configs[i].Name == name {
return &cf.Configs[i], nil
}
}
return nil, fmt.Errorf("config %q not found", name)
}

// GetConfigs returns configs matching the given names. If names is empty, returns all.
func (cf *ConfigFile) GetConfigs(names []string) ([]ToolConfig, error) {
if len(names) == 0 {
return cf.Configs, nil
}
nameSet := make(map[string]bool, len(names))
for _, n := range names {
nameSet[n] = true
}
var result []ToolConfig
for _, c := range cf.Configs {
if nameSet[c.Name] {
result = append(result, c)
delete(nameSet, c.Name)
}
}
if len(nameSet) > 0 {
var missing []string
for n := range nameSet {
missing = append(missing, n)
}
return nil, fmt.Errorf("configs not found: %v", missing)
}
return result, nil
}

// InstallSkillsAndPlugins runs "npx skills add <entry>" for each declared
// plugin across the given configs. It deduplicates entries so each
// package is only installed once.
func InstallSkillsAndPlugins(configs []ToolConfig) error {
seen := make(map[string]bool)
type entry struct {
kind  string
value string
}
var entries []entry

for _, c := range configs {
for _, p := range c.Plugins {
if !seen["plugin:"+p] {
seen["plugin:"+p] = true
entries = append(entries, entry{"plugin", p})
}
}
}

if len(entries) == 0 {
return nil
}

for _, e := range entries {
fmt.Printf("Installing %s: %s\n", e.kind, e.value)
cmd := exec.Command("npx", "skills", "add", e.value)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
return fmt.Errorf("installing %s %q: %w", e.kind, e.value, err)
}
}

return nil
}
