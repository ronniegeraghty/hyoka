package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestLoadPlugin(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "azure-sdk-tools.yaml"), []byte(`
name: azure-sdk-tools
description: Azure SDK development tools
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
  post_tool_use:
    - validate_file_sizes
`), 0644)

	reg := NewRegistry()
	if err := reg.LoadDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Count() != 1 {
		t.Fatalf("expected 1 plugin, got %d", reg.Count())
	}

	p, err := reg.Get("azure-sdk-tools")
	if err != nil {
		t.Fatalf("plugin not found: %v", err)
	}
	if p.Description != "Azure SDK development tools" {
		t.Errorf("wrong description: %q", p.Description)
	}
	if len(p.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(p.Skills))
	}
	if len(p.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(p.MCPServers))
	}
	if p.Hooks == nil {
		t.Fatal("expected hooks")
	}
	if len(p.Hooks.PreToolUse) != 1 {
		t.Errorf("expected 1 pre_tool_use hook, got %d", len(p.Hooks.PreToolUse))
	}
	if len(p.Hooks.PostToolUse) != 1 {
		t.Errorf("expected 1 post_tool_use hook, got %d", len(p.Hooks.PostToolUse))
	}
}

func TestLoadDirMultiplePlugins(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin-a.yaml"), []byte(`
name: plugin-a
skills:
  - type: local
    path: ./a
`), 0644)
	os.WriteFile(filepath.Join(dir, "plugin-b.yaml"), []byte(`
name: plugin-b
skills:
  - type: local
    path: ./b
`), 0644)

	reg := NewRegistry()
	if err := reg.LoadDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Count() != 2 {
		t.Errorf("expected 2 plugins, got %d", reg.Count())
	}

	names := reg.List()
	sort.Strings(names)
	if names[0] != "plugin-a" || names[1] != "plugin-b" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestLoadDirDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("name: same-name\nskills:\n  - type: local\n    path: ./a\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.yaml"), []byte("name: same-name\nskills:\n  - type: local\n    path: ./b\n"), 0644)

	reg := NewRegistry()
	err := reg.LoadDir(dir)
	if err == nil {
		t.Error("expected error for duplicate plugin names")
	}
}

func TestLoadDirSkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("not: valid: yaml: ["), 0644)
	os.WriteFile(filepath.Join(dir, "noname.yaml"), []byte("skills:\n  - type: local\n    path: ./x\n"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not yaml"), 0644)

	reg := NewRegistry()
	if err := reg.LoadDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Count() != 0 {
		t.Errorf("expected 0 valid plugins, got %d", reg.Count())
	}
}

func TestLoadDirNonexistent(t *testing.T) {
	reg := NewRegistry()
	// Non-existent directory should not error (just skip)
	if err := reg.LoadDir("/nonexistent/path"); err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for missing plugin")
	}
}

func TestPluginJSON(t *testing.T) {
	p := Plugin{
		Name:        "test",
		Description: "Test plugin",
		Skills: []PluginSkill{
			{Type: "local", Path: "./skills"},
		},
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Plugin
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "test" {
		t.Errorf("expected name 'test', got %q", decoded.Name)
	}
}

func TestAll(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "p.yaml"), []byte("name: p\nskills:\n  - type: local\n    path: ./x\n"), 0644)

	reg := NewRegistry()
	reg.LoadDir(dir)

	all := reg.All()
	if len(all) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(all))
	}
}

func TestApplyToGenerator(t *testing.T) {
	reg := NewRegistry()
	reg.plugins["sdk-tools"] = &Plugin{
		Name: "sdk-tools",
		Skills: []PluginSkill{
			{Type: "local", Path: "./skills/sdk"},
			{Type: "remote", Repo: "github.com/example/repo"},
		},
		MCPServers: map[string]*MCPServer{
			"azure-cli": {Type: "stdio", Command: "az", Args: []string{"mcp"}},
		},
	}
	reg.plugins["review-helpers"] = &Plugin{
		Name: "review-helpers",
		Skills: []PluginSkill{
			{Type: "local", Path: "./skills/review"},
		},
	}

	var skills []PluginSkill
	mcpServers := map[string]*MCPServer{}

	err := reg.ApplyToGenerator([]string{"sdk-tools", "review-helpers"}, &skills, &mcpServers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(skills))
	}
	if len(mcpServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(mcpServers))
	}
	if _, ok := mcpServers["azure-cli"]; !ok {
		t.Error("expected azure-cli MCP server")
	}
}

func TestApplyToGeneratorDedup(t *testing.T) {
	reg := NewRegistry()
	reg.plugins["a"] = &Plugin{
		Name:   "a",
		Skills: []PluginSkill{{Type: "local", Path: "./same"}},
		MCPServers: map[string]*MCPServer{
			"srv": {Type: "stdio", Command: "cmd"},
		},
	}
	reg.plugins["b"] = &Plugin{
		Name:   "b",
		Skills: []PluginSkill{{Type: "local", Path: "./same"}}, // duplicate
		MCPServers: map[string]*MCPServer{
			"srv": {Type: "stdio", Command: "other"}, // duplicate key
		},
	}

	var skills []PluginSkill
	mcpServers := map[string]*MCPServer{}

	err := reg.ApplyToGenerator([]string{"a", "b"}, &skills, &mcpServers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("expected 1 deduplicated skill, got %d", len(skills))
	}
	if len(mcpServers) != 1 {
		t.Errorf("expected 1 deduplicated MCP server, got %d", len(mcpServers))
	}
}

func TestApplyToGeneratorNotFound(t *testing.T) {
	reg := NewRegistry()
	var skills []PluginSkill
	mcpServers := map[string]*MCPServer{}

	err := reg.ApplyToGenerator([]string{"nonexistent"}, &skills, &mcpServers)
	if err == nil {
		t.Fatal("expected error for missing plugin")
	}
}
