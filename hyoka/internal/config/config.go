// Package config provides configuration loading and parsing for the evaluation tool.
package config

import (
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

// HooksConfig holds declarative lifecycle hook rules that map to the SDK's
// OnPreToolUse and OnPostToolUse callbacks. Hook names are strings that
// the eval engine resolves to concrete handler functions at runtime.
//
// Built-in hooks:
//   - validate_workspace_paths  — deny file writes outside workspace
//   - block_destructive_commands — deny destructive shell commands (rm -rf, az delete, etc.)
//   - validate_file_sizes       — warn on files exceeding size threshold
type HooksConfig struct {
	PreToolUse  []string `yaml:"pre_tool_use,omitempty" json:"pre_tool_use,omitempty"`
	PostToolUse []string `yaml:"post_tool_use,omitempty" json:"post_tool_use,omitempty"`
}

// Plugin represents a composable bundle of skills, MCP servers, and hooks.
// Plugins let users define a cohesive set of capabilities once and reference
// them by name across multiple configs.
//
// Example YAML:
//
//	plugins:
//	  - name: azure-sdk-tools
//	    skills:
//	      - type: local
//	        path: ../skills/generator
//	    mcp_servers:
//	      azure:
//	        type: sse
//	        command: npx
//	        args: ["-y", "@azure/mcp@latest"]
//	    hooks:
//	      pre_tool_use:
//	        - validate_workspace_paths
type Plugin struct {
	Name       string                `yaml:"name" json:"name"`
	Skills     []Skill               `yaml:"skills,omitempty" json:"skills,omitempty"`
	MCPServers map[string]*MCPServer `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	Hooks      *HooksConfig          `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// GeneratorConfig holds all configuration for the code generation agent.
type GeneratorConfig struct {
	Model          string                `yaml:"model" json:"model"`
	Skills         []Skill               `yaml:"skills,omitempty" json:"skills,omitempty"`
	MCPServers     map[string]*MCPServer `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	AvailableTools []string              `yaml:"available_tools,omitempty" json:"available_tools,omitempty"`
	ExcludedTools  []string              `yaml:"excluded_tools,omitempty" json:"excluded_tools,omitempty"`
	Plugins        []Plugin              `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	Hooks          *HooksConfig          `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	pluginsMerged bool // internal flag to prevent re-merging on repeated Normalize
}

// ReviewerConfig holds all configuration for the review/grading plane.
type ReviewerConfig struct {
	Model   string       `yaml:"model,omitempty" json:"model,omitempty"`
	Models  []string     `yaml:"models,omitempty" json:"models,omitempty"`
	Skills  []Skill      `yaml:"skills,omitempty" json:"skills,omitempty"`
	Plugins []Plugin     `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	Hooks   *HooksConfig `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	pluginsMerged bool // internal flag to prevent re-merging on repeated Normalize
}

// ToolConfig represents a single evaluation configuration.
// New-format configs use Generator and Reviewer sub-structs.
// Legacy top-level fields are supported for backward compatibility
// and are migrated to the sub-structs during Normalize().
type ToolConfig struct {
	Name        string           `yaml:"name" json:"name"`
	Description string           `yaml:"description" json:"description"`
	Generator   *GeneratorConfig `yaml:"generator,omitempty" json:"generator,omitempty"`
	Reviewer    *ReviewerConfig  `yaml:"reviewer,omitempty" json:"reviewer,omitempty"`

	// Legacy fields — kept for backward compatibility.
	// Normalize() maps these into Generator/Reviewer sub-structs.
	Model                      string                `yaml:"model,omitempty" json:"model,omitempty"`
	ReviewerModel              string                `yaml:"reviewer_model,omitempty" json:"reviewer_model,omitempty"`
	ReviewerModels             []string              `yaml:"reviewer_models,omitempty" json:"reviewer_models,omitempty"`
	MCPServers                 map[string]*MCPServer `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	SkillDirectories           []string              `yaml:"skill_directories,omitempty" json:"skill_directories,omitempty"`
	GeneratorSkillDirectories  []string              `yaml:"generator_skill_directories,omitempty" json:"generator_skill_directories,omitempty"`
	ReviewerSkillDirectories   []string              `yaml:"reviewer_skill_directories,omitempty" json:"reviewer_skill_directories,omitempty"`
	AvailableTools             []string              `yaml:"available_tools,omitempty" json:"available_tools,omitempty"`
	ExcludedTools              []string              `yaml:"excluded_tools,omitempty" json:"excluded_tools,omitempty"`
	Skills                     []string              `yaml:"skills,omitempty" json:"skills,omitempty"`
	Plugins                    []string              `yaml:"plugins,omitempty" json:"plugins,omitempty"`
}

// Normalize migrates legacy top-level fields into the Generator/Reviewer
// sub-structs and merges plugin contributions. It is idempotent — safe to
// call multiple times.
func (tc *ToolConfig) Normalize() {
	if tc.Generator == nil {
		tc.Generator = &GeneratorConfig{}
	}
	if tc.Reviewer == nil {
		tc.Reviewer = &ReviewerConfig{}
	}

	// Model
	if tc.Generator.Model == "" && tc.Model != "" {
		tc.Generator.Model = tc.Model
	}

	// Reviewer models
	if len(tc.Reviewer.Models) == 0 {
		if len(tc.ReviewerModels) > 0 {
			tc.Reviewer.Models = tc.ReviewerModels
		} else if tc.ReviewerModel != "" {
			tc.Reviewer.Models = []string{tc.ReviewerModel}
		}
	}

	// MCP servers
	if tc.Generator.MCPServers == nil && tc.MCPServers != nil {
		tc.Generator.MCPServers = tc.MCPServers
	}

	// Available/excluded tools
	if len(tc.Generator.AvailableTools) == 0 && len(tc.AvailableTools) > 0 {
		tc.Generator.AvailableTools = tc.AvailableTools
	}
	if len(tc.Generator.ExcludedTools) == 0 && len(tc.ExcludedTools) > 0 {
		tc.Generator.ExcludedTools = tc.ExcludedTools
	}

	// Skill directories → Generator.Skills (type: local)
	if len(tc.Generator.Skills) == 0 {
		dirs := tc.GeneratorSkillDirectories
		if len(dirs) == 0 {
			dirs = tc.SkillDirectories
		}
		for _, d := range dirs {
			tc.Generator.Skills = append(tc.Generator.Skills, Skill{Type: "local", Path: d})
		}
	}

	// Reviewer skill directories → Reviewer.Skills (type: local)
	if len(tc.Reviewer.Skills) == 0 {
		for _, d := range tc.ReviewerSkillDirectories {
			tc.Reviewer.Skills = append(tc.Reviewer.Skills, Skill{Type: "local", Path: d})
		}
	}

	// Merge generator plugins into generator config (only once)
	if !tc.Generator.pluginsMerged {
		mergePlugins(tc.Generator.Plugins, &tc.Generator.Skills, &tc.Generator.MCPServers, &tc.Generator.Hooks)
		tc.Generator.pluginsMerged = true
	}

	// Merge reviewer plugins into reviewer config (only once)
	if !tc.Reviewer.pluginsMerged {
		mergePlugins(tc.Reviewer.Plugins, &tc.Reviewer.Skills, nil, &tc.Reviewer.Hooks)
		tc.Reviewer.pluginsMerged = true
	}
}

// mergePlugins folds plugin contributions (skills, MCP servers, hooks) into
// the parent config's fields. Plugin-contributed items are appended after
// directly-configured items, preserving plugin declaration order.
func mergePlugins(plugins []Plugin, skills *[]Skill, mcpServers *map[string]*MCPServer, hooks **HooksConfig) {
	for _, p := range plugins {
		// Merge skills — plugin skills appended after direct skills
		*skills = append(*skills, p.Skills...)

		// Merge MCP servers — plugin servers added if not already present
		if mcpServers != nil && len(p.MCPServers) > 0 {
			if *mcpServers == nil {
				*mcpServers = make(map[string]*MCPServer)
			}
			for name, srv := range p.MCPServers {
				if _, exists := (*mcpServers)[name]; !exists {
					(*mcpServers)[name] = srv
				} else {
					slog.Debug("Plugin MCP server skipped (already defined at config level)",
						"plugin", p.Name, "server", name)
				}
			}
		}

		// Merge hooks — plugin hooks appended after direct hooks
		if p.Hooks != nil {
			if *hooks == nil {
				*hooks = &HooksConfig{}
			}
			(*hooks).PreToolUse = appendUnique((*hooks).PreToolUse, p.Hooks.PreToolUse)
			(*hooks).PostToolUse = appendUnique((*hooks).PostToolUse, p.Hooks.PostToolUse)
		}
	}
}

// appendUnique appends items from src to dst, skipping duplicates.
func appendUnique(dst, src []string) []string {
	seen := make(map[string]bool, len(dst))
	for _, s := range dst {
		seen[s] = true
	}
	for _, s := range src {
		if !seen[s] {
			dst = append(dst, s)
			seen[s] = true
		}
	}
	return dst
}

// EffectiveModel returns the generator model, preferring Generator.Model.
func (tc *ToolConfig) EffectiveModel() string {
	if tc.Generator != nil && tc.Generator.Model != "" {
		return tc.Generator.Model
	}
	return tc.Model
}

// EffectiveReviewerModels returns the list of reviewer models to use.
func (tc *ToolConfig) EffectiveReviewerModels() []string {
	if tc.Reviewer != nil && len(tc.Reviewer.Models) > 0 {
		return tc.Reviewer.Models
	}
	if len(tc.ReviewerModels) > 0 {
		return tc.ReviewerModels
	}
	if tc.ReviewerModel != "" {
		return []string{tc.ReviewerModel}
	}
	return nil
}

// EffectiveMCPServers returns the MCP servers config, preferring Generator.MCPServers.
func (tc *ToolConfig) EffectiveMCPServers() map[string]*MCPServer {
	if tc.Generator != nil && len(tc.Generator.MCPServers) > 0 {
		return tc.Generator.MCPServers
	}
	return tc.MCPServers
}

// EffectiveAvailableTools returns available tools, preferring Generator.AvailableTools.
func (tc *ToolConfig) EffectiveAvailableTools() []string {
	if tc.Generator != nil && len(tc.Generator.AvailableTools) > 0 {
		return tc.Generator.AvailableTools
	}
	return tc.AvailableTools
}

// EffectiveExcludedTools returns excluded tools, preferring Generator.ExcludedTools.
func (tc *ToolConfig) EffectiveExcludedTools() []string {
	if tc.Generator != nil && len(tc.Generator.ExcludedTools) > 0 {
		return tc.Generator.ExcludedTools
	}
	return tc.ExcludedTools
}

// EffectiveGeneratorSkills returns the generator's skill list from the normalized config.
func (tc *ToolConfig) EffectiveGeneratorSkills() []Skill {
	if tc.Generator != nil {
		return tc.Generator.Skills
	}
	return nil
}

// EffectiveReviewerSkills returns the reviewer's skill list from the normalized config.
func (tc *ToolConfig) EffectiveReviewerSkills() []Skill {
	if tc.Reviewer != nil {
		return tc.Reviewer.Skills
	}
	return nil
}

// EffectiveGeneratorHooks returns the merged hook config for the generator,
// combining direct hooks with plugin-contributed hooks (after Normalize).
func (tc *ToolConfig) EffectiveGeneratorHooks() *HooksConfig {
	if tc.Generator != nil {
		return tc.Generator.Hooks
	}
	return nil
}

// EffectiveReviewerHooks returns the merged hook config for the reviewer,
// combining direct hooks with plugin-contributed hooks (after Normalize).
func (tc *ToolConfig) EffectiveReviewerHooks() *HooksConfig {
	if tc.Reviewer != nil {
		return tc.Reviewer.Hooks
	}
	return nil
}

// EffectiveGeneratorPlugins returns the generator's plugin list.
func (tc *ToolConfig) EffectiveGeneratorPlugins() []Plugin {
	if tc.Generator != nil {
		return tc.Generator.Plugins
	}
	return nil
}

// EffectiveReviewerPlugins returns the reviewer's plugin list.
func (tc *ToolConfig) EffectiveReviewerPlugins() []Plugin {
	if tc.Reviewer != nil {
		return tc.Reviewer.Plugins
	}
	return nil
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
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}
	// Normalize all configs (migrate legacy fields → sub-structs)
	for i := range cfg.Configs {
		cfg.Configs[i].Normalize()
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
		// Validate skills have correct type
		for _, s := range c.EffectiveGeneratorSkills() {
			if err := validateSkill(s); err != nil {
				return fmt.Errorf("config %q generator skill: %w", c.Name, err)
			}
		}
		for _, s := range c.EffectiveReviewerSkills() {
			if err := validateSkill(s); err != nil {
				return fmt.Errorf("config %q reviewer skill: %w", c.Name, err)
			}
		}
		// Validate generator plugins
		if err := validatePlugins(c.EffectiveGeneratorPlugins(), c.Name, "generator"); err != nil {
			return err
		}
		// Validate reviewer plugins
		if err := validatePlugins(c.EffectiveReviewerPlugins(), c.Name, "reviewer"); err != nil {
			return err
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
	}
	return nil
}

// validatePlugins checks that plugins have valid names, no duplicate names,
// no conflicting MCP server names, and valid skill entries.
func validatePlugins(plugins []Plugin, configName, section string) error {
	seenNames := make(map[string]bool, len(plugins))
	seenMCP := make(map[string]string) // mcp server name → plugin name

	for _, p := range plugins {
		if p.Name == "" {
			return fmt.Errorf("config %q %s: plugin missing name", configName, section)
		}
		if seenNames[p.Name] {
			return fmt.Errorf("config %q %s: duplicate plugin name %q", configName, section, p.Name)
		}
		seenNames[p.Name] = true

		// Validate plugin skills
		for _, s := range p.Skills {
			if err := validateSkill(s); err != nil {
				return fmt.Errorf("config %q %s plugin %q skill: %w", configName, section, p.Name, err)
			}
		}

		// Check for MCP server name conflicts across plugins
		for name := range p.MCPServers {
			if prevPlugin, exists := seenMCP[name]; exists {
				return fmt.Errorf("config %q %s: MCP server %q defined in both plugin %q and %q",
					configName, section, name, prevPlugin, p.Name)
			}
			seenMCP[name] = p.Name
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
// skill and plugin across the given configs. It deduplicates entries so each
// package is only installed once.
func InstallSkillsAndPlugins(configs []ToolConfig) error {
seen := make(map[string]bool)
type entry struct {
kind  string
value string
}
var entries []entry

for _, c := range configs {
for _, s := range c.Skills {
if !seen["skill:"+s] {
seen["skill:"+s] = true
entries = append(entries, entry{"skill", s})
}
}
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
