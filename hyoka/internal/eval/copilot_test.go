package eval

import (
	"testing"

	"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/config"
)

func TestBuildSessionConfig_EmptyAvailableToolsIsNil(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	// When config has an empty available_tools slice (parsed from YAML "available_tools: []"),
	// the SDK must receive nil — not an empty slice — so the CLI exposes all default tools.
	cfg := &config.ToolConfig{
		Name:           "test",
		Model:          "gpt-4",
		AvailableTools: []string{},
		ExcludedTools:  []string{},
	}
	sc := e.buildSessionConfig(cfg, "/tmp/test")

	if sc.AvailableTools != nil {
		t.Errorf("expected AvailableTools nil (all tools), got %v", sc.AvailableTools)
	}
	if sc.ExcludedTools != nil {
		t.Errorf("expected ExcludedTools nil (no exclusions), got %v", sc.ExcludedTools)
	}
}

func TestBuildSessionConfig_NilAvailableToolsIsNil(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{
		Name:  "test",
		Model: "gpt-4",
	}
	sc := e.buildSessionConfig(cfg, "/tmp/test")

	if sc.AvailableTools != nil {
		t.Errorf("expected AvailableTools nil, got %v", sc.AvailableTools)
	}
	if sc.ExcludedTools != nil {
		t.Errorf("expected ExcludedTools nil, got %v", sc.ExcludedTools)
	}
}

func TestBuildSessionConfig_PopulatedAvailableToolsPreserved(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{
		Name:           "test",
		Model:          "gpt-4",
		AvailableTools: []string{"create", "edit", "bash"},
		ExcludedTools:  []string{"web_fetch"},
	}
	sc := e.buildSessionConfig(cfg, "/tmp/test")

	if len(sc.AvailableTools) != 3 {
		t.Errorf("expected 3 AvailableTools, got %d", len(sc.AvailableTools))
	}
	if len(sc.ExcludedTools) != 1 {
		t.Errorf("expected 1 ExcludedTools, got %d", len(sc.ExcludedTools))
	}
}

func TestBuildSessionConfig_WorkingDirectory(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{Name: "test", Model: "gpt-4"}
	sc := e.buildSessionConfig(cfg, "/workspace/eval-123")

	if sc.WorkingDirectory != "/workspace/eval-123" {
		t.Errorf("expected WorkingDirectory '/workspace/eval-123', got %q", sc.WorkingDirectory)
	}
}

func TestBuildSessionConfig_PermissionHandler(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{Name: "test", Model: "gpt-4"}
	sc := e.buildSessionConfig(cfg, "/tmp/test")

	if sc.OnPermissionRequest == nil {
		t.Error("expected OnPermissionRequest to be set (ApproveAll)")
	}
}

func TestBuildSessionConfig_MCPServers(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{
		Name:  "test",
		Model: "gpt-4",
		MCPServers: map[string]*config.MCPServer{
			"azure": {
				Type:    "local",
				Command: "npx",
				Args:    []string{"-y", "@azure/mcp@latest"},
			},
		},
	}
	sc := e.buildSessionConfig(cfg, "/tmp/test")

	if len(sc.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(sc.MCPServers))
	}
	azure, ok := sc.MCPServers["azure"]
	if !ok {
		t.Fatal("expected 'azure' MCP server")
	}
	if azure["command"] != "npx" {
		t.Errorf("expected MCP command 'npx', got %v", azure["command"])
	}
}
