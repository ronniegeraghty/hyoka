// Package plugin implements a composable plugin system for hyoka configs.
//
// A plugin bundles skills, MCP servers, and hooks into a reusable unit
// that can be referenced by name from evaluation configurations.
package plugin

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Plugin represents a reusable, composable unit of configuration that bundles
// skills, MCP servers, and hooks.
type Plugin struct {
	Name        string                `yaml:"name" json:"name"`
	Description string                `yaml:"description,omitempty" json:"description,omitempty"`
	Skills      []PluginSkill         `yaml:"skills,omitempty" json:"skills,omitempty"`
	MCPServers  map[string]*MCPServer `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	Hooks       *HookConfig           `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Source      string                `yaml:"-" json:"source,omitempty"` // file path
}

// PluginSkill is the same as config.Skill but defined here to avoid import cycles.
type PluginSkill struct {
	Type string `yaml:"type" json:"type"`
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Repo string `yaml:"repo,omitempty" json:"repo,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// MCPServer represents an MCP server configuration within a plugin.
type MCPServer struct {
	Type    string   `yaml:"type" json:"type"`
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args" json:"args"`
	Tools   []string `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// HookConfig defines declarative pre/post tool-use hooks.
type HookConfig struct {
	PreToolUse  []string `yaml:"pre_tool_use,omitempty" json:"pre_tool_use,omitempty"`
	PostToolUse []string `yaml:"post_tool_use,omitempty" json:"post_tool_use,omitempty"`
}

// Registry holds loaded plugins indexed by name.
type Registry struct {
	plugins map[string]*Plugin
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]*Plugin)}
}

// LoadDir loads all plugin YAML files from a directory.
func (r *Registry) LoadDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Debug("Plugin directory does not exist, skipping", "dir", dir)
		return nil
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		p, err := loadPlugin(path)
		if err != nil {
			slog.Warn("Skipping invalid plugin file", "path", path, "error", err)
			return nil
		}
		if _, exists := r.plugins[p.Name]; exists {
			return fmt.Errorf("duplicate plugin name %q (first: %s, second: %s)",
				p.Name, r.plugins[p.Name].Source, path)
		}
		r.plugins[p.Name] = p
		slog.Debug("Loaded plugin", "name", p.Name, "path", path,
			"skills", len(p.Skills), "mcp_servers", len(p.MCPServers))
		return nil
	})
}

func loadPlugin(path string) (*Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Plugin
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if p.Name == "" {
		return nil, fmt.Errorf("%s: plugin missing name", path)
	}
	p.Source = path
	return &p, nil
}

// Get returns a plugin by name, or an error if not found.
func (r *Registry) Get(name string) (*Plugin, error) {
	p, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found", name)
	}
	return p, nil
}

// List returns all loaded plugin names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// All returns all loaded plugins.
func (r *Registry) All() []*Plugin {
	plugins := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// Count returns the number of registered plugins.
func (r *Registry) Count() int {
	return len(r.plugins)
}

// ApplyToGenerator resolves plugin names and merges their skills and MCP
// servers into the generator's existing configuration. Duplicate skills and
// MCP servers are skipped.
func (r *Registry) ApplyToGenerator(pluginNames []string, skills *[]PluginSkill, mcpServers *map[string]*MCPServer) error {
	for _, name := range pluginNames {
		p, err := r.Get(name)
		if err != nil {
			return err
		}
		for _, s := range p.Skills {
			if !containsSkill(*skills, s) {
				*skills = append(*skills, s)
			}
		}
		if len(p.MCPServers) > 0 {
			if *mcpServers == nil {
				*mcpServers = make(map[string]*MCPServer)
			}
			for k, v := range p.MCPServers {
				if _, exists := (*mcpServers)[k]; !exists {
					(*mcpServers)[k] = v
				}
			}
		}
	}
	return nil
}

func containsSkill(skills []PluginSkill, s PluginSkill) bool {
	for _, existing := range skills {
		if existing.Type == s.Type && existing.Name == s.Name &&
			existing.Repo == s.Repo && existing.Path == s.Path {
			return true
		}
	}
	return false
}
