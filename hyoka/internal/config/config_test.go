package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseValidConfig(t *testing.T) {
	data := []byte(`
configs:
  - name: test-config
    description: "Test configuration"
    generator:
      model: "gpt-4"
  - name: test-config-2
    description: "Second test"
    generator:
      model: "claude-sonnet-4.5"
      mcp_servers:
        azure:
          type: local
          command: npx
          args: ["-y", "@azure/mcp@latest"]
          tools: ["*"]
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
	if cfg.Configs[0].Generator.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %q", cfg.Configs[0].Generator.Model)
	}
	// Check MCP server on second config
	if cfg.Configs[1].Generator.MCPServers == nil {
		t.Fatal("expected MCP servers on second config")
	}
	azure, ok := cfg.Configs[1].Generator.MCPServers["azure"]
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
    generator:
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
    generator:
      model: "claude-opus-4.6"
    reviewer:
      models:
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
    generator:
      model: "claude-sonnet-4.5"
    reviewer:
      models:
        - "gpt-4.1"
        - "gemini-3-pro-preview"
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	models := cfg.Configs[0].Reviewer.Models
	if len(models) != 2 {
		t.Errorf("expected 2 reviewer models, got %d", len(models))
	}
}

func TestValidateNoReviewerModelAccepted(t *testing.T) {
	data := []byte(`
configs:
  - name: no-reviewer
    description: "No reviewer model specified"
    generator:
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
    generator:
      model: "claude-sonnet-4.5"
    reviewer:
      models:
        - "gpt-4.1"
        - "gpt-4.1"
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for duplicate reviewer models")
	}
}

func TestReviewerSingleModel(t *testing.T) {
	data := []byte(`
configs:
  - name: single-reviewer
    description: "Single reviewer model"
    generator:
      model: "claude-sonnet-4.5"
    reviewer:
      models:
        - "gpt-4.1"
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	models := cfg.Configs[0].Reviewer.Models
	if len(models) != 1 || models[0] != "gpt-4.1" {
		t.Errorf("expected [gpt-4.1], got %v", models)
	}
}

func TestGetConfig(t *testing.T) {
	data := []byte(`
configs:
  - name: alpha
    description: "Alpha"
    generator:
      model: "gpt-4"
  - name: beta
    description: "Beta"
    generator:
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
    generator:
      model: "gpt-4"
  - name: beta
    description: "Beta"
    generator:
      model: "claude-sonnet-4.5"
  - name: gamma
    description: "Gamma"
    generator:
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
    generator:
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

func TestParseGeneratorSkillsAndPlugins(t *testing.T) {
	data := []byte(`
configs:
  - name: with-skills
    description: "Config with skills and plugins"
    generator:
      model: "gpt-4"
      skills:
        - type: local
          path: "./skills/tool-use"
        - type: remote
          name: org-skill
          repo: "github:org/repo"
    plugins:
      - "@azure/functions"
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]
	if len(c.Generator.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(c.Generator.Skills))
	}
	if c.Generator.Skills[0].Type != "local" || c.Generator.Skills[0].Path != "./skills/tool-use" {
		t.Errorf("expected local skill './skills/tool-use', got %+v", c.Generator.Skills[0])
	}
	if c.Generator.Skills[1].Type != "remote" || c.Generator.Skills[1].Repo != "github:org/repo" {
		t.Errorf("expected remote skill 'github:org/repo', got %+v", c.Generator.Skills[1])
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
    generator:
      model: "gpt-4"
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]
	if len(c.Generator.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(c.Generator.Skills))
	}
	if len(c.Plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(c.Plugins))
	}
}

func TestInstallSkillsAndPluginsEmpty(t *testing.T) {
	configs := []ToolConfig{
		{Name: "empty", Description: "No skills", Generator: &GeneratorConfig{Model: "gpt-4"}},
	}
	if err := InstallSkillsAndPlugins(configs); err != nil {
		t.Fatalf("expected no error for empty skills/plugins, got: %v", err)
	}
}

func TestParseNewFormatGeneratorReviewer(t *testing.T) {
data := []byte(`
configs:
  - name: new-format
    description: "New format with generator/reviewer"
    generator:
      model: "claude-sonnet-4.5"
      skills:
        - type: local
          path: "./skills/generator"
      mcp_servers:
        azure:
          type: local
          command: npx
          args: ["-y", "@azure/mcp@latest"]
          tools: ["*"]
      available_tools: ["create", "edit"]
      excluded_tools: ["web_fetch"]
    reviewer:
      models:
        - "claude-opus-4.6"
        - "gemini-3-pro-preview"
      skills:
        - type: local
          path: "./skills/reviewer"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
c := cfg.Configs[0]

if c.Generator.Model != "claude-sonnet-4.5" {
t.Errorf("expected model 'claude-sonnet-4.5', got %q", c.Generator.Model)
}
models := c.Reviewer.Models
if len(models) != 2 || models[0] != "claude-opus-4.6" {
t.Errorf("expected reviewer models [claude-opus-4.6 gemini-3-pro-preview], got %v", models)
}
genSkills := c.Generator.Skills
if len(genSkills) != 1 || genSkills[0].Type != "local" {
t.Errorf("expected 1 generator skill (local), got %v", genSkills)
}
revSkills := c.Reviewer.Skills
if len(revSkills) != 1 || revSkills[0].Path != "./skills/reviewer" {
t.Errorf("expected 1 reviewer skill, got %v", revSkills)
}
mcpServers := c.Generator.MCPServers
if len(mcpServers) != 1 {
t.Errorf("expected 1 MCP server, got %d", len(mcpServers))
}
if len(c.Generator.AvailableTools) != 2 {
t.Errorf("expected 2 available tools, got %d", len(c.Generator.AvailableTools))
}
if len(c.Generator.ExcludedTools) != 1 {
t.Errorf("expected 1 excluded tool, got %d", len(c.Generator.ExcludedTools))
}
}

func TestGeneratorReviewerFieldsPopulated(t *testing.T) {
	data := []byte(`
configs:
  - name: full-config
    description: "Full config with generator and reviewer"
    generator:
      model: "claude-opus-4.6"
      mcp_servers:
        azure:
          type: local
          command: npx
          args: ["-y", "@azure/mcp@latest"]
      skills:
        - type: local
          path: "./skills/generator"
      available_tools: ["create"]
      excluded_tools: ["bash"]
    reviewer:
      models:
        - "gpt-4.1"
      skills:
        - type: local
          path: "./skills/reviewer"
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	if c.Generator == nil {
		t.Fatal("Generator should not be nil")
	}
	if c.Reviewer == nil {
		t.Fatal("Reviewer should not be nil")
	}
	if c.Generator.Model != "claude-opus-4.6" {
		t.Errorf("expected Generator.Model 'claude-opus-4.6', got %q", c.Generator.Model)
	}
	if len(c.Reviewer.Models) != 1 || c.Reviewer.Models[0] != "gpt-4.1" {
		t.Errorf("expected Reviewer.Models [gpt-4.1], got %v", c.Reviewer.Models)
	}
	if len(c.Generator.Skills) != 1 || c.Generator.Skills[0].Path != "./skills/generator" {
		t.Errorf("expected 1 generator skill, got %v", c.Generator.Skills)
	}
	if len(c.Reviewer.Skills) != 1 || c.Reviewer.Skills[0].Path != "./skills/reviewer" {
		t.Errorf("expected 1 reviewer skill, got %v", c.Reviewer.Skills)
	}
	if len(c.Generator.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(c.Generator.MCPServers))
	}
	if len(c.Generator.AvailableTools) != 1 {
		t.Errorf("expected 1 available tool, got %d", len(c.Generator.AvailableTools))
	}
}

func TestParseRemoteSkill(t *testing.T) {
data := []byte(`
configs:
  - name: with-remote
    description: "Config with remote skill"
    generator:
      model: "gpt-4"
      skills:
        - type: remote
          name: azure-keyvault-py
          repo: microsoft/skills
        - type: local
          path: "./skills/local"
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
c := cfg.Configs[0]
skills := c.Generator.Skills
if len(skills) != 2 {
t.Fatalf("expected 2 skills, got %d", len(skills))
}
if skills[0].Type != "remote" || skills[0].Name != "azure-keyvault-py" || skills[0].Repo != "microsoft/skills" {
t.Errorf("unexpected remote skill: %+v", skills[0])
}
if skills[1].Type != "local" || skills[1].Path != "./skills/local" {
t.Errorf("unexpected local skill: %+v", skills[1])
}
}

func TestValidateRejectsInvalidSkillType(t *testing.T) {
data := []byte(`
configs:
  - name: bad-skill
    description: "Bad skill type"
    generator:
      model: "gpt-4"
      skills:
        - type: invalid
          path: "./foo"
`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for invalid skill type")
}
}

func TestValidateRejectsLocalSkillMissingPath(t *testing.T) {
data := []byte(`
configs:
  - name: no-path
    description: "Local skill missing path"
    generator:
      model: "gpt-4"
      skills:
        - type: local
`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for local skill without path")
}
}

func TestValidateRejectsRemoteSkillMissingRepo(t *testing.T) {
data := []byte(`
configs:
  - name: no-repo
    description: "Remote skill missing repo"
    generator:
      model: "gpt-4"
      skills:
        - type: remote
          name: some-skill
`)
_, err := Parse(data)
if err == nil {
t.Fatal("expected error for remote skill without repo")
}
}

func TestParseDuplicateConfigNamesRejected(t *testing.T) {
	data := []byte(`
configs:
  - name: same-name
    description: "First"
    generator:
      model: "gpt-4"
  - name: same-name
    description: "Second"
    generator:
      model: "claude-sonnet-4.5"
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for duplicate config names within a file")
	}
	if got := err.Error(); !strings.Contains(got, "duplicate config name") {
		t.Errorf("expected duplicate config name error, got: %v", err)
	}
}

func TestLoadDirDuplicateConfigNamesAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	file1 := []byte(`
configs:
  - name: shared-name
    description: "In file1"
    generator:
      model: "gpt-4"
`)
	file2 := []byte(`
configs:
  - name: shared-name
    description: "In file2"
    generator:
      model: "claude-sonnet-4.5"
`)
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), file1, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yaml"), file2, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for duplicate config names across files")
	}
	got := err.Error()
	if !strings.Contains(got, "duplicate config name") || !strings.Contains(got, "a.yaml") || !strings.Contains(got, "b.yaml") {
		t.Errorf("expected error mentioning duplicate name and both files, got: %v", err)
	}
}

func TestGeneratorModelDirectAccess(t *testing.T) {
	c := ToolConfig{
		Name: "test",
		Generator: &GeneratorConfig{
			Model: "new-model",
		},
	}
	if c.Generator.Model != "new-model" {
		t.Errorf("expected 'new-model', got %q", c.Generator.Model)
	}
}

func TestValidateRejectsNilGenerator(t *testing.T) {
	cf := &ConfigFile{
		Configs: []ToolConfig{
			{Name: "no-gen"},
		},
	}
	err := cf.Validate()
	if err == nil {
		t.Fatal("expected error for nil generator")
	}
	want := `config "no-gen": generator.model is required`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestParseSystemPrompt(t *testing.T) {
	data := []byte(`
configs:
  - name: with-system-prompt
    description: "Config with system prompts"
    generator:
      model: "claude-opus-4.6"
      system_prompt: "You are an Azure SDK expert."
    reviewer:
      models:
        - "gpt-4.1"
      system_prompt: "Review code for Azure best practices."
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]
	if c.Generator.SystemPrompt != "You are an Azure SDK expert." {
		t.Errorf("expected generator system_prompt 'You are an Azure SDK expert.', got %q", c.Generator.SystemPrompt)
	}
	if c.Reviewer.SystemPrompt != "Review code for Azure best practices." {
		t.Errorf("expected reviewer system_prompt 'Review code for Azure best practices.', got %q", c.Reviewer.SystemPrompt)
	}
}

func TestParseSystemPromptOmitted(t *testing.T) {
	data := []byte(`
configs:
  - name: no-system-prompt
    description: "Config without system prompts"
    generator:
      model: "gpt-4"
    reviewer:
      models:
        - "gpt-4.1"
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]
	if c.Generator.SystemPrompt != "" {
		t.Errorf("expected empty generator system_prompt, got %q", c.Generator.SystemPrompt)
	}
	if c.Reviewer.SystemPrompt != "" {
		t.Errorf("expected empty reviewer system_prompt, got %q", c.Reviewer.SystemPrompt)
	}
}

func TestValidateRejectsEmptyGeneratorModel(t *testing.T) {
	cf := &ConfigFile{
		Configs: []ToolConfig{
			{Name: "empty-model", Generator: &GeneratorConfig{Model: ""}},
		},
	}
	err := cf.Validate()
	if err == nil {
		t.Fatal("expected error for empty generator model")
	}
	want := `config "empty-model": generator.model is required`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}
