// Package config provides configuration loading and parsing for the evaluation tool.
package config

import (
"fmt"
"os"

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
type ToolConfig struct {
Name             string                `yaml:"name" json:"name"`
Description      string                `yaml:"description" json:"description"`
Model            string                `yaml:"model" json:"model"`
MCPServers       map[string]*MCPServer `yaml:"mcp_servers" json:"mcp_servers"`
SkillDirectories []string              `yaml:"skill_directories" json:"skill_directories"`
AvailableTools   []string              `yaml:"available_tools" json:"available_tools"`
ExcludedTools    []string              `yaml:"excluded_tools" json:"excluded_tools"`
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

// Parse parses configuration from YAML bytes.
func Parse(data []byte) (*ConfigFile, error) {
var cfg ConfigFile
if err := yaml.Unmarshal(data, &cfg); err != nil {
return nil, fmt.Errorf("parsing config YAML: %w", err)
}
if len(cfg.Configs) == 0 {
return nil, fmt.Errorf("no configs defined")
}
for i, c := range cfg.Configs {
if c.Name == "" {
return nil, fmt.Errorf("config at index %d has no name", i)
}
}
return &cfg, nil
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
