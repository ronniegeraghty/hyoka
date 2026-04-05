package eval

import (
"testing"

"github.com/ronniegeraghty/hyoka/internal/config"
"github.com/ronniegeraghty/hyoka/internal/prompt"
)

func TestBuildSessionConfig_EmptyAvailableToolsIsNil(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model:          "gpt-4",
AvailableTools: []string{},
ExcludedTools:  []string{},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
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
Name:      "test",
Generator: &config.GeneratorConfig{Model: "gpt-4"},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
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
Name: "test",
Generator: &config.GeneratorConfig{
Model:          "gpt-4",
AvailableTools: []string{"create", "edit", "bash"},
ExcludedTools:  []string{"web_fetch"},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
if len(sc.AvailableTools) != 3 {
t.Errorf("expected 3 AvailableTools, got %d", len(sc.AvailableTools))
}
if len(sc.ExcludedTools) != 1 {
t.Errorf("expected 1 ExcludedTools, got %d", len(sc.ExcludedTools))
}
}

func TestBuildSessionConfig_WorkingDirectory(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{Name: "test", Generator: &config.GeneratorConfig{Model: "gpt-4"}}
sc := e.buildSessionConfig(cfg, "/workspace/eval-123", "", nil)
if sc.WorkingDirectory != "/workspace/eval-123" {
t.Errorf("expected WorkingDirectory '/workspace/eval-123', got %q", sc.WorkingDirectory)
}
}

func TestBuildSessionConfig_ConfigDir(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{Name: "test", Generator: &config.GeneratorConfig{Model: "gpt-4"}}
sc := e.buildSessionConfig(cfg, "/workspace/eval-123", "/isolated/config", nil)
if sc.ConfigDir != "/isolated/config" {
t.Errorf("expected ConfigDir '/isolated/config', got %q", sc.ConfigDir)
}
}

func TestBuildSessionConfig_PermissionHandler(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{Name: "test", Generator: &config.GeneratorConfig{Model: "gpt-4"}}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
if sc.OnPermissionRequest == nil {
t.Error("expected OnPermissionRequest to be set (ApproveAll)")
}
}

func TestBuildSessionConfig_MCPServers(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model: "gpt-4",
MCPServers: map[string]*config.MCPServer{
"azure": {Type: "local", Command: "npx", Args: []string{"-y", "@azure/mcp@latest"}},
},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
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

// --- Tool filter resolution tests ---

func TestBuildSessionConfig_ToolEntryResolution(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model: "gpt-4",
Tools: []config.ToolEntry{
{Name: "create"},
{Name: "edit"},
{Name: "azure_mcp", When: map[string]string{"language": "python"}},
},
},
}

// Python prompt should get all 3 tools
sc := e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{"language": "python", "service": "identity"})
if len(sc.AvailableTools) != 3 {
t.Fatalf("expected 3 AvailableTools for python, got %d: %v", len(sc.AvailableTools), sc.AvailableTools)
}

// Dotnet prompt should get only 2 tools (azure_mcp excluded)
sc = e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{"language": "dotnet"})
if len(sc.AvailableTools) != 2 {
t.Fatalf("expected 2 AvailableTools for dotnet, got %d: %v", len(sc.AvailableTools), sc.AvailableTools)
}
for _, tool := range sc.AvailableTools {
if tool == "azure_mcp" {
t.Error("azure_mcp should not be included for dotnet")
}
}
}

func TestBuildSessionConfig_ToolEntryOverridesAvailableTools(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model:          "gpt-4",
Tools:          []config.ToolEntry{{Name: "create"}},
AvailableTools: []string{"create", "edit", "bash"},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
if len(sc.AvailableTools) != 1 || sc.AvailableTools[0] != "create" {
t.Errorf("expected Tools to override AvailableTools, got %v", sc.AvailableTools)
}
}

func TestBuildSessionConfig_LegacyAvailableToolsFallback(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model:          "gpt-4",
AvailableTools: []string{"create", "edit"},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{"language": "python"})
if len(sc.AvailableTools) != 2 {
t.Errorf("expected legacy AvailableTools [create edit], got %v", sc.AvailableTools)
}
}

func TestBuildSessionConfig_ToolEntryAllConditionalNoneMatch(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model: "gpt-4",
Tools: []config.ToolEntry{
{Name: "azure_mcp", When: map[string]string{"language": "python"}},
},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{"language": "dotnet"})
if sc.AvailableTools != nil {
t.Errorf("expected nil AvailableTools when no tools match, got %v", sc.AvailableTools)
}
}

func TestBuildSessionConfig_ExcludedToolsWithToolEntries(t *testing.T) {
e := &CopilotSDKEvaluator{}
cfg := &config.ToolConfig{
Name: "test",
Generator: &config.GeneratorConfig{
Model:         "gpt-4",
Tools:         []config.ToolEntry{{Name: "create"}, {Name: "edit"}},
ExcludedTools: []string{"web_fetch"},
},
}
sc := e.buildSessionConfig(cfg, "/workspace/test", "", nil)
if len(sc.AvailableTools) != 2 {
t.Errorf("expected 2 AvailableTools, got %d", len(sc.AvailableTools))
}
if len(sc.ExcludedTools) != 1 || sc.ExcludedTools[0] != "web_fetch" {
t.Errorf("expected ExcludedTools [web_fetch], got %v", sc.ExcludedTools)
}
}

func TestMergePromptProperties(t *testing.T) {
tests := []struct {
name string
p    *prompt.Prompt
want map[string]string
}{
{
name: "all fields populated",
p: &prompt.Prompt{
Properties: map[string]string{
"service": "identity", "language": "python", "plane": "data-plane",
"category": "auth", "difficulty": "medium",
},
},
want: map[string]string{
"service": "identity", "language": "python", "plane": "data-plane",
"category": "auth", "difficulty": "medium",
},
},
{
name: "partial fields",
p: &prompt.Prompt{Properties: map[string]string{"service": "storage", "language": "dotnet"}},
want: map[string]string{"service": "storage", "language": "dotnet"},
},
{
name: "empty prompt",
p:    &prompt.Prompt{},
want: map[string]string{},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := mergePromptProperties(tt.p)
for k, v := range tt.want {
if got[k] != v {
t.Errorf("key %q: got %q, want %q", k, got[k], v)
}
}
for k := range got {
if _, ok := tt.want[k]; !ok {
t.Errorf("unexpected key %q=%q in merged properties", k, got[k])
}
}
})
}
}

func TestBuildSessionConfig_CustomSystemPrompt(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{
		Name: "test",
		Generator: &config.GeneratorConfig{
			Model:        "gpt-4",
			SystemPrompt: "You are a helpful code generator.",
		},
	}
	sc := e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{"language": "python"})

	if sc.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %q", sc.Model)
	}
}

func TestBuildSessionConfig_EmptySystemPrompt(t *testing.T) {
	e := &CopilotSDKEvaluator{}

	cfg := &config.ToolConfig{
		Name: "test",
		Generator: &config.GeneratorConfig{
			Model: "gpt-4",
		},
	}
	sc := e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{})

	if sc.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %q", sc.Model)
	}
}

// --- Integration tests: YAML config → prompt properties → session config tools ---

func TestIntegration_YAMLConfigToSessionTools(t *testing.T) {
	yamlData := `
configs:
  - name: integration-test
    description: "Integration test config"
    generator:
      model: gpt-4
      tools:
        - name: create
        - name: edit
        - name: azure_mcp
          when:
            language: python
        - name: dotnet_tools
          when:
            language: dotnet
            service: identity
      excluded_tools:
        - web_fetch
`
	cf, err := config.Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(cf.Configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(cf.Configs))
	}
	cfg := &cf.Configs[0]

	tests := []struct {
		name          string
		prompt        *prompt.Prompt
		wantAvail     []string
		wantExcluded  []string
	}{
		{
			name:          "python prompt gets azure_mcp",
			prompt:        &prompt.Prompt{Properties: map[string]string{"language": "python", "service": "identity"}},
			wantAvail:     []string{"create", "edit", "azure_mcp"},
			wantExcluded:  []string{"web_fetch"},
		},
		{
			name:          "dotnet+identity prompt gets dotnet_tools",
			prompt:        &prompt.Prompt{Properties: map[string]string{"language": "dotnet", "service": "identity"}},
			wantAvail:     []string{"create", "edit", "dotnet_tools"},
			wantExcluded:  []string{"web_fetch"},
		},
		{
			name:          "dotnet+storage prompt gets only unconditional",
			prompt:        &prompt.Prompt{Properties: map[string]string{"language": "dotnet", "service": "storage"}},
			wantAvail:     []string{"create", "edit"},
			wantExcluded:  []string{"web_fetch"},
		},
		{
			name:          "go prompt gets only unconditional",
			prompt:        &prompt.Prompt{Properties: map[string]string{"language": "go", "service": "key-vault"}},
			wantAvail:     []string{"create", "edit"},
			wantExcluded:  []string{"web_fetch"},
		},
	}

	e := &CopilotSDKEvaluator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := mergePromptProperties(tt.prompt)
			sc := e.buildSessionConfig(cfg, "/workspace/test", "", props)

			if len(sc.AvailableTools) != len(tt.wantAvail) {
				t.Fatalf("AvailableTools: got %v, want %v", sc.AvailableTools, tt.wantAvail)
			}
			for i, tool := range tt.wantAvail {
				if sc.AvailableTools[i] != tool {
					t.Errorf("AvailableTools[%d] = %q, want %q", i, sc.AvailableTools[i], tool)
				}
			}

			if len(sc.ExcludedTools) != len(tt.wantExcluded) {
				t.Fatalf("ExcludedTools: got %v, want %v", sc.ExcludedTools, tt.wantExcluded)
			}
			for i, tool := range tt.wantExcluded {
				if sc.ExcludedTools[i] != tool {
					t.Errorf("ExcludedTools[%d] = %q, want %q", i, sc.ExcludedTools[i], tool)
				}
			}
		})
	}
}

func TestIntegration_LegacyYAMLFallback(t *testing.T) {
	yamlData := `
configs:
  - name: legacy-test
    description: "Legacy format test"
    generator:
      model: gpt-4
      available_tools:
        - create
        - edit
        - bash
      excluded_tools:
        - web_fetch
`
	cf, err := config.Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	cfg := &cf.Configs[0]

	e := &CopilotSDKEvaluator{}
	props := mergePromptProperties(&prompt.Prompt{Properties: map[string]string{"language": "python", "service": "identity"}})
	sc := e.buildSessionConfig(cfg, "/workspace/test", "", props)

	if len(sc.AvailableTools) != 3 {
		t.Fatalf("expected 3 AvailableTools from legacy format, got %d: %v", len(sc.AvailableTools), sc.AvailableTools)
	}
	if len(sc.ExcludedTools) != 1 || sc.ExcludedTools[0] != "web_fetch" {
		t.Errorf("expected ExcludedTools [web_fetch], got %v", sc.ExcludedTools)
	}
}

func TestIntegration_DuplicateToolEntries(t *testing.T) {
	e := &CopilotSDKEvaluator{}
	cfg := &config.ToolConfig{
		Name: "test",
		Generator: &config.GeneratorConfig{
			Model: "gpt-4",
			Tools: []config.ToolEntry{
				{Name: "create"},
				{Name: "create", When: map[string]string{"language": "python"}},
				{Name: "edit"},
			},
		},
	}
	sc := e.buildSessionConfig(cfg, "/workspace/test", "", map[string]string{"language": "python"})
	if len(sc.AvailableTools) != 2 {
		t.Fatalf("expected 2 tools after dedup, got %d: %v", len(sc.AvailableTools), sc.AvailableTools)
	}
	if sc.AvailableTools[0] != "create" || sc.AvailableTools[1] != "edit" {
		t.Errorf("expected [create edit], got %v", sc.AvailableTools)
	}
}
