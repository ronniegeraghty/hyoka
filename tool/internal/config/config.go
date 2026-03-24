// Package config provides configuration loading and parsing for the evaluation tool.
package config

import (
"fmt"
"os"
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

// ToolConfig represents a single evaluation configuration.
// A config defines the tooling environment (MCP servers, skills) and can list
// multiple generator models to test. The engine expands each model into a
// separate eval run: prompts × configs × models.
type ToolConfig struct {
Name                       string                `yaml:"name" json:"name"`
Description                string                `yaml:"description" json:"description"`
Model                      string                `yaml:"model" json:"model"`
Models                     []string              `yaml:"models" json:"models"`
ReviewerModel              string                `yaml:"reviewer_model" json:"reviewer_model"`
ReviewerModels             []string              `yaml:"reviewer_models" json:"reviewer_models"`
MCPServers                 map[string]*MCPServer `yaml:"mcp_servers" json:"mcp_servers"`
SkillDirectories           []string              `yaml:"skill_directories" json:"skill_directories"`
GeneratorSkillDirectories  []string              `yaml:"generator_skill_directories" json:"generator_skill_directories"`
ReviewerSkillDirectories   []string              `yaml:"reviewer_skill_directories" json:"reviewer_skill_directories"`
AvailableTools             []string              `yaml:"available_tools" json:"available_tools"`
ExcludedTools              []string              `yaml:"excluded_tools" json:"excluded_tools"`
}

// EffectiveReviewerModels returns the list of reviewer models to use.
func (tc *ToolConfig) EffectiveReviewerModels() []string {
if len(tc.ReviewerModels) > 0 {
return tc.ReviewerModels
}
if tc.ReviewerModel != "" {
return []string{tc.ReviewerModel}
}
return nil
}

// Expand returns one ToolConfig per generator model. If models (plural) is set,
// each model gets its own config with name "{config}-{model}". If only model
// (singular) is set, returns the config as-is. This enables
// prompts × configs × models evaluation matrix.
func (tc *ToolConfig) Expand() []ToolConfig {
models := tc.Models
if len(models) == 0 {
if tc.Model != "" {
return []ToolConfig{*tc}
}
// No model at all — return as-is, engine will use default
return []ToolConfig{*tc}
}

expanded := make([]ToolConfig, 0, len(models))
for _, m := range models {
c := *tc
c.Model = m
c.Models = nil // clear to avoid re-expansion
c.Name = tc.Name + "/" + m
expanded = append(expanded, c)
}
return expanded
}

// ConfigFile represents the top-level config file structure.
type ConfigFile struct {
Configs []ToolConfig `yaml:"configs"`
}

// Load reads and parses a configuration file from the given path.
func Load(path string) (*ConfigFile, error) {
data, err := os.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("reading config file: %w", err)
}
return Parse(data)
}

// LoadDir reads all .yaml files in a directory and merges their configs.
// This allows splitting configs across multiple files (e.g., baseline.yaml, azure-mcp.yaml).
func LoadDir(dir string) (*ConfigFile, error) {
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
if err := yaml.Unmarshal(data, &cfg); err != nil {
return nil, fmt.Errorf("parsing config YAML: %w", err)
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
// Check for duplicate reviewer models
reviewerModels := c.EffectiveReviewerModels()
seen := make(map[string]bool, len(reviewerModels))
for _, rm := range reviewerModels {
if seen[rm] {
return fmt.Errorf("config %q: duplicate reviewer model %q", c.Name, rm)
}
seen[rm] = true
}
// For each expanded config (per generator model), check that
// at least one reviewer model differs from the generator model.
// A reviewer model matching the generator is allowed — it will be
// auto-skipped at runtime — but ALL reviewers matching is an error.
for _, expanded := range c.Expand() {
if expanded.Model == "" {
continue
}
allMatch := true
for _, rm := range reviewerModels {
if rm != expanded.Model {
allMatch = false
break
}
}
if allMatch && len(reviewerModels) > 0 {
return fmt.Errorf("config %q: all reviewer models match generator model %q — at least one must differ", c.Name, expanded.Model)
}
}
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
