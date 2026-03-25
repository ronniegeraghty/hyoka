package config

import (
"os"
"path/filepath"
"testing"
)

func TestParseValidConfig(t *testing.T) {
data := []byte(`
configs:
  - name: test-config
    description: "Test configuration"
    model: "gpt-4"
    mcp_servers: {}
    skill_directories: []
    available_tools: []
    excluded_tools: []
  - name: test-config-2
    description: "Second test"
    model: "claude-sonnet-4.5"
    mcp_servers:
      azure:
        type: local
        command: npx
        args: ["-y", "@azure/mcp@latest"]
        tools: ["*"]
    skill_directories: []
    available_tools: []
    excluded_tools: []
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(cfg.Configs) != 2 {
t.Fatalf("expected 2 configs, got %d", len(cfg.Configs))
}
if cfg.Configs[0].Name != "test-config" {
t.Errorf("expected name 'test-config', got %q", cfg.Configs[0].Name)
}
if cfg.Configs[0].Model != "gpt-4" {
t.Errorf("expected model 'gpt-4', got %q", cfg.Configs[0].Model)
}
// Check MCP server on second config
if cfg.Configs[1].MCPServers == nil {
t.Fatal("expected MCP servers on second config")
}
azure, ok := cfg.Configs[1].MCPServers["azure"]
if !ok {
t.Fatal("expected 'azure' MCP server")
}
if azure.Command != "npx" {
t.Errorf("expected command 'npx', got %q", azure.Command)
}
}

func TestParseEmptyConfig(t *testing.T) {
data := []byte(`configs: []`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for empty configs")
}
}

func TestParseConfigMissingName(t *testing.T) {
data := []byte(`
configs:
  - description: "No name"
    model: "gpt-4"
`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for config missing name")
}
}

func TestParseInvalidYAML(t *testing.T) {
data := []byte(`not: valid: yaml: [`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for invalid YAML")
}
}

func TestValidateSameModelAccepted(t *testing.T) {
data := []byte(`
configs:
  - name: same-model-ok
    description: "Same model for generator and reviewer is allowed"
    model: "claude-opus-4.6"
    reviewer_models:
      - "claude-opus-4.6"
      - "gpt-4.1"
`)
_, err := Parse(data)
if err != nil {
t.Fatalf("expected no error when reviewer model matches generator, got: %v", err)
}
}

func TestValidateDifferentModelsAccepted(t *testing.T) {
data := []byte(`
configs:
  - name: good-config
    description: "Different models"
    model: "claude-sonnet-4.5"
    reviewer_models:
      - "gpt-4.1"
      - "gemini-3-pro-preview"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
models := cfg.Configs[0].EffectiveReviewerModels()
if len(models) != 2 {
t.Errorf("expected 2 reviewer models, got %d", len(models))
}
}

func TestValidateNoReviewerModelAccepted(t *testing.T) {
data := []byte(`
configs:
  - name: no-reviewer
    description: "No reviewer model specified"
    model: "claude-sonnet-4.5"
`)
_, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestValidateDuplicateReviewerModelsRejected(t *testing.T) {
data := []byte(`
configs:
  - name: dupes
    description: "Duplicate reviewer models"
    model: "claude-sonnet-4.5"
    reviewer_models:
      - "gpt-4.1"
      - "gpt-4.1"
`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for duplicate reviewer models")
}
}

func TestBackwardCompatSingularReviewerModel(t *testing.T) {
data := []byte(`
configs:
  - name: compat
    description: "Old-style singular reviewer_model"
    model: "claude-sonnet-4.5"
    reviewer_model: "gpt-4.1"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
models := cfg.Configs[0].EffectiveReviewerModels()
if len(models) != 1 || models[0] != "gpt-4.1" {
t.Errorf("expected [gpt-4.1] from backward compat, got %v", models)
}
}

func TestGetConfig(t *testing.T) {
data := []byte(`
configs:
  - name: alpha
    description: "Alpha"
    model: "gpt-4"
  - name: beta
    description: "Beta"
    model: "claude-sonnet-4.5"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

tc, err := cfg.GetConfig("beta")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if tc.Name != "beta" {
t.Errorf("expected 'beta', got %q", tc.Name)
}

_, err = cfg.GetConfig("nonexistent")
if err == nil {
t.Fatal("expected error for nonexistent config")
}
}

func TestGetConfigs(t *testing.T) {
data := []byte(`
configs:
  - name: alpha
    description: "Alpha"
    model: "gpt-4"
  - name: beta
    description: "Beta"
    model: "claude-sonnet-4.5"
  - name: gamma
    description: "Gamma"
    model: "gpt-4"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

// Empty names returns all
all, err := cfg.GetConfigs(nil)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(all) != 3 {
t.Errorf("expected 3 configs, got %d", len(all))
}

// Specific names
subset, err := cfg.GetConfigs([]string{"alpha", "gamma"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(subset) != 2 {
t.Errorf("expected 2 configs, got %d", len(subset))
}

// Missing name
_, err = cfg.GetConfigs([]string{"alpha", "missing"})
if err == nil {
t.Fatal("expected error for missing config name")
}
}

func TestLoadFromFile(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "config.yaml")
content := []byte(`
configs:
  - name: file-test
    description: "From file"
    model: "gpt-4"
`)
if err := os.WriteFile(path, content, 0644); err != nil {
t.Fatalf("failed to write test file: %v", err)
}

cfg, err := Load(path)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.Configs[0].Name != "file-test" {
t.Errorf("expected 'file-test', got %q", cfg.Configs[0].Name)
}

// Non-existent file
_, err = Load(filepath.Join(dir, "nonexistent.yaml"))
if err == nil {
t.Fatal("expected error for nonexistent file")
}
}

func TestParseSkillsAndPlugins(t *testing.T) {
data := []byte(`
configs:
  - name: with-skills
    description: "Config with skills and plugins"
    model: "gpt-4"
    skills:
      - "@anthropic/tool-use"
      - "github:org/repo"
    plugins:
      - "@azure/functions"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
c := cfg.Configs[0]
if len(c.Skills) != 2 {
t.Errorf("expected 2 skills, got %d", len(c.Skills))
}
if c.Skills[0] != "@anthropic/tool-use" {
t.Errorf("expected skill '@anthropic/tool-use', got %q", c.Skills[0])
}
if c.Skills[1] != "github:org/repo" {
t.Errorf("expected skill 'github:org/repo', got %q", c.Skills[1])
}
if len(c.Plugins) != 1 {
t.Errorf("expected 1 plugin, got %d", len(c.Plugins))
}
if c.Plugins[0] != "@azure/functions" {
t.Errorf("expected plugin '@azure/functions', got %q", c.Plugins[0])
}
}

func TestParseNoSkillsOrPlugins(t *testing.T) {
data := []byte(`
configs:
  - name: no-extras
    description: "Config without skills or plugins"
    model: "gpt-4"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
c := cfg.Configs[0]
if len(c.Skills) != 0 {
t.Errorf("expected 0 skills, got %d", len(c.Skills))
}
if len(c.Plugins) != 0 {
t.Errorf("expected 0 plugins, got %d", len(c.Plugins))
}
}

func TestInstallSkillsAndPluginsEmpty(t *testing.T) {
configs := []ToolConfig{
{Name: "empty", Description: "No skills", Model: "gpt-4"},
}
if err := InstallSkillsAndPlugins(configs); err != nil {
t.Fatalf("expected no error for empty skills/plugins, got: %v", err)
}
}

func TestParseRemoteSkills(t *testing.T) {
data := []byte(`
configs:
  - name: remote-test
    model: "gpt-4"
    remote_skills:
      - repo: microsoft/skills
        skills:
          - keyvault-secrets-java
          - storage-blob-python
      - repo: myorg/custom-skills
        skills:
          - my-skill
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
c := cfg.Configs[0]
if len(c.RemoteSkills) != 2 {
t.Fatalf("expected 2 remote_skills entries, got %d", len(c.RemoteSkills))
}
rs := c.RemoteSkills[0]
if rs.Repo != "microsoft/skills" {
t.Errorf("expected repo %q, got %q", "microsoft/skills", rs.Repo)
}
if len(rs.Skills) != 2 {
t.Errorf("expected 2 skills, got %d", len(rs.Skills))
}
if rs.Skills[0] != "keyvault-secrets-java" {
t.Errorf("expected first skill %q, got %q", "keyvault-secrets-java", rs.Skills[0])
}
rs2 := c.RemoteSkills[1]
if rs2.Repo != "myorg/custom-skills" {
t.Errorf("expected repo %q, got %q", "myorg/custom-skills", rs2.Repo)
}
}
