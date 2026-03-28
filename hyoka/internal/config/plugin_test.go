package config

import (
	"strings"
	"testing"
)

func TestParseGeneratorPlugins(t *testing.T) {
	data := []byte(`
configs:
  - name: with-plugins
    description: "Config with generator plugins"
    generator:
      model: "claude-opus-4.6"
      skills:
        - type: local
          path: "./skills/direct"
      plugins:
        - name: azure-sdk-tools
          skills:
            - type: local
              path: ../skills/generator
            - type: remote
              repo: github.com/Azure/ai-hub-sdk
              name: azure-sdk-tools
          mcp_servers:
            azure:
              type: sse
              command: npx
              args: ["-y", "@azure/mcp@latest"]
          hooks:
            pre_tool_use:
              - validate_workspace_paths
              - block_destructive_commands
            post_tool_use:
              - validate_file_sizes
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	// Plugin should be present
	plugins := c.EffectiveGeneratorPlugins()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "azure-sdk-tools" {
		t.Errorf("expected plugin name 'azure-sdk-tools', got %q", plugins[0].Name)
	}

	// Skills should be merged: 1 direct + 2 from plugin = 3
	skills := c.EffectiveGeneratorSkills()
	if len(skills) != 3 {
		t.Fatalf("expected 3 merged skills, got %d", len(skills))
	}
	if skills[0].Path != "./skills/direct" {
		t.Errorf("first skill should be direct, got %+v", skills[0])
	}
	if skills[1].Path != "../skills/generator" {
		t.Errorf("second skill should be from plugin, got %+v", skills[1])
	}
	if skills[2].Type != "remote" || skills[2].Name != "azure-sdk-tools" {
		t.Errorf("third skill should be remote from plugin, got %+v", skills[2])
	}

	// MCP servers from plugin should be merged
	mcpServers := c.EffectiveMCPServers()
	if len(mcpServers) != 1 {
		t.Fatalf("expected 1 MCP server from plugin, got %d", len(mcpServers))
	}
	azure, ok := mcpServers["azure"]
	if !ok {
		t.Fatal("expected 'azure' MCP server from plugin")
	}
	if azure.Command != "npx" {
		t.Errorf("expected command 'npx', got %q", azure.Command)
	}

	// Hooks from plugin should be merged
	hooks := c.EffectiveGeneratorHooks()
	if hooks == nil {
		t.Fatal("expected hooks from plugin")
	}
	if len(hooks.PreToolUse) != 2 {
		t.Errorf("expected 2 pre_tool_use hooks, got %d", len(hooks.PreToolUse))
	}
	if hooks.PreToolUse[0] != "validate_workspace_paths" {
		t.Errorf("expected first hook 'validate_workspace_paths', got %q", hooks.PreToolUse[0])
	}
	if len(hooks.PostToolUse) != 1 || hooks.PostToolUse[0] != "validate_file_sizes" {
		t.Errorf("expected 1 post_tool_use hook 'validate_file_sizes', got %v", hooks.PostToolUse)
	}
}

func TestParseReviewerPlugins(t *testing.T) {
	data := []byte(`
configs:
  - name: reviewer-plugins
    description: "Config with reviewer plugins"
    generator:
      model: "gpt-4"
    reviewer:
      models:
        - "claude-opus-4.6"
      plugins:
        - name: sdk-review-standards
          skills:
            - type: local
              path: ../skills/reviewer
          hooks:
            pre_tool_use:
              - validate_workspace_paths
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	plugins := c.EffectiveReviewerPlugins()
	if len(plugins) != 1 || plugins[0].Name != "sdk-review-standards" {
		t.Fatalf("expected 1 reviewer plugin 'sdk-review-standards', got %v", plugins)
	}

	skills := c.EffectiveReviewerSkills()
	if len(skills) != 1 || skills[0].Path != "../skills/reviewer" {
		t.Errorf("expected reviewer skills from plugin, got %v", skills)
	}

	hooks := c.EffectiveReviewerHooks()
	if hooks == nil || len(hooks.PreToolUse) != 1 {
		t.Fatalf("expected 1 reviewer pre_tool_use hook, got %v", hooks)
	}
}

func TestMultiplePluginsMerge(t *testing.T) {
	data := []byte(`
configs:
  - name: multi-plugin
    description: "Config with multiple plugins"
    generator:
      model: "claude-opus-4.6"
      plugins:
        - name: plugin-a
          skills:
            - type: local
              path: ./skills/a
          mcp_servers:
            server-a:
              type: sse
              command: cmd-a
          hooks:
            pre_tool_use:
              - validate_workspace_paths
        - name: plugin-b
          skills:
            - type: local
              path: ./skills/b
          mcp_servers:
            server-b:
              type: sse
              command: cmd-b
          hooks:
            pre_tool_use:
              - block_destructive_commands
            post_tool_use:
              - validate_file_sizes
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	// Skills merged from both plugins
	skills := c.EffectiveGeneratorSkills()
	if len(skills) != 2 {
		t.Errorf("expected 2 skills from 2 plugins, got %d", len(skills))
	}

	// MCP servers merged from both plugins
	mcpServers := c.EffectiveMCPServers()
	if len(mcpServers) != 2 {
		t.Errorf("expected 2 MCP servers from 2 plugins, got %d", len(mcpServers))
	}

	// Hooks merged from both plugins
	hooks := c.EffectiveGeneratorHooks()
	if hooks == nil {
		t.Fatal("expected hooks")
	}
	if len(hooks.PreToolUse) != 2 {
		t.Errorf("expected 2 pre_tool_use hooks, got %d: %v", len(hooks.PreToolUse), hooks.PreToolUse)
	}
	if len(hooks.PostToolUse) != 1 {
		t.Errorf("expected 1 post_tool_use hook, got %d", len(hooks.PostToolUse))
	}
}

func TestPluginMCPServerConfigLevelPrecedence(t *testing.T) {
	data := []byte(`
configs:
  - name: precedence
    description: "Config-level MCP servers take precedence over plugin"
    generator:
      model: "gpt-4"
      mcp_servers:
        azure:
          type: sse
          command: config-cmd
      plugins:
        - name: tools
          mcp_servers:
            azure:
              type: sse
              command: plugin-cmd
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	// Config-level 'azure' should win over plugin's 'azure'
	mcpServers := c.EffectiveMCPServers()
	if len(mcpServers) != 1 {
		t.Fatalf("expected 1 MCP server, got %d", len(mcpServers))
	}
	if mcpServers["azure"].Command != "config-cmd" {
		t.Errorf("expected config-level command 'config-cmd', got %q", mcpServers["azure"].Command)
	}
}

func TestPluginHookDeduplication(t *testing.T) {
	data := []byte(`
configs:
  - name: dedup-hooks
    description: "Duplicate hook names across plugins are deduplicated"
    generator:
      model: "gpt-4"
      hooks:
        pre_tool_use:
          - validate_workspace_paths
      plugins:
        - name: plugin-a
          hooks:
            pre_tool_use:
              - validate_workspace_paths
              - block_destructive_commands
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	hooks := c.EffectiveGeneratorHooks()
	if hooks == nil {
		t.Fatal("expected hooks")
	}
	// validate_workspace_paths should appear only once
	if len(hooks.PreToolUse) != 2 {
		t.Errorf("expected 2 unique pre_tool_use hooks, got %d: %v", len(hooks.PreToolUse), hooks.PreToolUse)
	}
	if hooks.PreToolUse[0] != "validate_workspace_paths" || hooks.PreToolUse[1] != "block_destructive_commands" {
		t.Errorf("unexpected hooks: %v", hooks.PreToolUse)
	}
}

func TestValidateRejectsPluginMissingName(t *testing.T) {
	data := []byte(`
configs:
  - name: bad-plugin
    description: "Plugin with no name"
    generator:
      model: "gpt-4"
      plugins:
        - skills:
            - type: local
              path: ./skills
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for plugin missing name")
	}
	if !strings.Contains(err.Error(), "plugin missing name") {
		t.Errorf("expected 'plugin missing name' error, got: %v", err)
	}
}

func TestValidateRejectsDuplicatePluginNames(t *testing.T) {
	data := []byte(`
configs:
  - name: dupe-plugins
    description: "Two plugins with same name"
    generator:
      model: "gpt-4"
      plugins:
        - name: tools
          skills:
            - type: local
              path: ./skills/a
        - name: tools
          skills:
            - type: local
              path: ./skills/b
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for duplicate plugin names")
	}
	if !strings.Contains(err.Error(), "duplicate plugin name") {
		t.Errorf("expected 'duplicate plugin name' error, got: %v", err)
	}
}

func TestValidateRejectsConflictingMCPServerAcrossPlugins(t *testing.T) {
	data := []byte(`
configs:
  - name: mcp-conflict
    description: "Two plugins define same MCP server"
    generator:
      model: "gpt-4"
      plugins:
        - name: plugin-a
          mcp_servers:
            azure:
              type: sse
              command: cmd-a
        - name: plugin-b
          mcp_servers:
            azure:
              type: sse
              command: cmd-b
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for conflicting MCP server names across plugins")
	}
	if !strings.Contains(err.Error(), "MCP server") && !strings.Contains(err.Error(), "azure") {
		t.Errorf("expected MCP server conflict error, got: %v", err)
	}
}

func TestValidateRejectsInvalidSkillInPlugin(t *testing.T) {
	data := []byte(`
configs:
  - name: bad-plugin-skill
    description: "Plugin with invalid skill type"
    generator:
      model: "gpt-4"
      plugins:
        - name: bad-skills
          skills:
            - type: invalid
              path: ./foo
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for invalid skill type in plugin")
	}
	if !strings.Contains(err.Error(), "plugin") && !strings.Contains(err.Error(), "skill") {
		t.Errorf("expected plugin skill validation error, got: %v", err)
	}
}

func TestPluginsWithDirectHooksMerge(t *testing.T) {
	data := []byte(`
configs:
  - name: direct-and-plugin-hooks
    description: "Direct hooks and plugin hooks merge"
    generator:
      model: "gpt-4"
      hooks:
        pre_tool_use:
          - validate_workspace_paths
        post_tool_use:
          - validate_file_sizes
      plugins:
        - name: safety
          hooks:
            pre_tool_use:
              - block_destructive_commands
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	hooks := c.EffectiveGeneratorHooks()
	if hooks == nil {
		t.Fatal("expected hooks")
	}
	if len(hooks.PreToolUse) != 2 {
		t.Errorf("expected 2 pre_tool_use hooks (1 direct + 1 plugin), got %d: %v",
			len(hooks.PreToolUse), hooks.PreToolUse)
	}
	if len(hooks.PostToolUse) != 1 {
		t.Errorf("expected 1 post_tool_use hook (direct only), got %d", len(hooks.PostToolUse))
	}
}

func TestNoPluginsConfigUnchanged(t *testing.T) {
	data := []byte(`
configs:
  - name: no-plugins
    description: "Config without plugins"
    generator:
      model: "gpt-4"
      skills:
        - type: local
          path: ./skills
    reviewer:
      models:
        - claude-opus-4.6
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	if len(c.EffectiveGeneratorPlugins()) != 0 {
		t.Error("expected 0 generator plugins")
	}
	if len(c.EffectiveReviewerPlugins()) != 0 {
		t.Error("expected 0 reviewer plugins")
	}
	if c.EffectiveGeneratorHooks() != nil {
		t.Error("expected nil generator hooks")
	}
	if c.EffectiveReviewerHooks() != nil {
		t.Error("expected nil reviewer hooks")
	}
}

func TestPluginNormalizeIdempotent(t *testing.T) {
	c := ToolConfig{
		Name: "test",
		Generator: &GeneratorConfig{
			Model: "gpt-4",
			Plugins: []Plugin{
				{
					Name:   "tools",
					Skills: []Skill{{Type: "local", Path: "./skills/plugin"}},
					Hooks:  &HooksConfig{PreToolUse: []string{"validate_workspace_paths"}},
				},
			},
		},
	}
	c.Normalize()
	skillCount1 := len(c.Generator.Skills)
	hookCount1 := len(c.Generator.Hooks.PreToolUse)

	c.Normalize()
	skillCount2 := len(c.Generator.Skills)
	hookCount2 := len(c.Generator.Hooks.PreToolUse)

	if skillCount1 != skillCount2 {
		t.Errorf("Normalize not idempotent for skills: %d vs %d", skillCount1, skillCount2)
	}
	if hookCount1 != hookCount2 {
		t.Errorf("Normalize not idempotent for hooks: %d vs %d", hookCount1, hookCount2)
	}
}

func TestFullPluginConfigFromIssue(t *testing.T) {
	// This test uses the exact config format from issue #50
	data := []byte(`
configs:
  - name: azure-mcp/claude-opus-4.6
    generator:
      model: claude-opus-4.6
      plugins:
        - name: azure-sdk-tools
          skills:
            - type: local
              path: ../skills/generator
            - type: remote
              repo: github.com/Azure/ai-hub-sdk
              name: azure-sdk-tools
          mcp_servers:
            azure:
              type: sse
              command: npx
              args: ["-y", "@azure/mcp@latest"]
          hooks:
            pre_tool_use:
              - validate_workspace_paths
              - block_destructive_commands
            post_tool_use:
              - validate_file_sizes
    reviewer:
      models:
        - claude-opus-4.6
        - gpt-5.3-codex
      plugins:
        - name: sdk-review-standards
          skills:
            - type: local
              path: ../skills/reviewer
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error parsing issue #50 config: %v", err)
	}
	c := cfg.Configs[0]

	// Generator: 1 plugin with 2 skills, 1 MCP server, 3 hooks
	if len(c.EffectiveGeneratorPlugins()) != 1 {
		t.Errorf("expected 1 generator plugin")
	}
	if len(c.EffectiveGeneratorSkills()) != 2 {
		t.Errorf("expected 2 generator skills (from plugin), got %d", len(c.EffectiveGeneratorSkills()))
	}
	if len(c.EffectiveMCPServers()) != 1 {
		t.Errorf("expected 1 MCP server (from plugin), got %d", len(c.EffectiveMCPServers()))
	}
	hooks := c.EffectiveGeneratorHooks()
	if hooks == nil || len(hooks.PreToolUse) != 2 || len(hooks.PostToolUse) != 1 {
		t.Errorf("expected 2 pre + 1 post hooks, got %v", hooks)
	}

	// Reviewer: 1 plugin with 1 skill
	if len(c.EffectiveReviewerPlugins()) != 1 {
		t.Errorf("expected 1 reviewer plugin")
	}
	if len(c.EffectiveReviewerSkills()) != 1 {
		t.Errorf("expected 1 reviewer skill (from plugin), got %d", len(c.EffectiveReviewerSkills()))
	}
}

func TestDirectHooksWithoutPlugins(t *testing.T) {
	data := []byte(`
configs:
  - name: direct-hooks
    description: "Hooks without any plugins"
    generator:
      model: "gpt-4"
      hooks:
        pre_tool_use:
          - validate_workspace_paths
        post_tool_use:
          - validate_file_sizes
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	hooks := c.EffectiveGeneratorHooks()
	if hooks == nil {
		t.Fatal("expected hooks without plugins")
	}
	if len(hooks.PreToolUse) != 1 || hooks.PreToolUse[0] != "validate_workspace_paths" {
		t.Errorf("unexpected pre_tool_use: %v", hooks.PreToolUse)
	}
	if len(hooks.PostToolUse) != 1 || hooks.PostToolUse[0] != "validate_file_sizes" {
		t.Errorf("unexpected post_tool_use: %v", hooks.PostToolUse)
	}
}

func TestPluginWithOnlyMCPServers(t *testing.T) {
	data := []byte(`
configs:
  - name: mcp-only-plugin
    description: "Plugin that only contributes MCP servers"
    generator:
      model: "gpt-4"
      plugins:
        - name: azure-servers
          mcp_servers:
            azure:
              type: sse
              command: npx
              args: ["-y", "@azure/mcp@latest"]
            github:
              type: sse
              command: npx
              args: ["-y", "@github/mcp@latest"]
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]

	mcpServers := c.EffectiveMCPServers()
	if len(mcpServers) != 2 {
		t.Errorf("expected 2 MCP servers from plugin, got %d", len(mcpServers))
	}
	if len(c.EffectiveGeneratorSkills()) != 0 {
		t.Error("expected 0 skills when plugin only has MCP servers")
	}
}
