package config

import (
"testing"
)

func TestResolveTools_EmptyEntries(t *testing.T) {
result := ResolveTools(nil, map[string]string{"language": "python"})
if result != nil {
t.Errorf("expected nil for empty entries, got %v", result)
}
}

func TestResolveTools_UnconditionalTools(t *testing.T) {
entries := []ToolEntry{
{Name: "create"},
{Name: "edit"},
{Name: "bash"},
}
result := ResolveTools(entries, map[string]string{"language": "python"})
if len(result) != 3 {
t.Fatalf("expected 3 tools, got %d", len(result))
}
want := []string{"create", "edit", "bash"}
for i, name := range want {
if result[i] != name {
t.Errorf("result[%d] = %q, want %q", i, result[i], name)
}
}
}

func TestResolveTools_ConditionalMatch(t *testing.T) {
entries := []ToolEntry{
{Name: "create"},
{Name: "azure_mcp", When: map[string]string{"language": "python"}},
}
props := map[string]string{"language": "python", "service": "identity"}
result := ResolveTools(entries, props)
if len(result) != 2 {
t.Fatalf("expected 2 tools, got %d: %v", len(result), result)
}
if result[0] != "create" || result[1] != "azure_mcp" {
t.Errorf("unexpected result: %v", result)
}
}

func TestResolveTools_ConditionalNoMatch(t *testing.T) {
entries := []ToolEntry{
{Name: "create"},
{Name: "azure_mcp", When: map[string]string{"language": "python"}},
}
props := map[string]string{"language": "dotnet"}
result := ResolveTools(entries, props)
if len(result) != 1 {
t.Fatalf("expected 1 tool, got %d: %v", len(result), result)
}
if result[0] != "create" {
t.Errorf("expected 'create', got %q", result[0])
}
}

func TestResolveTools_MultipleConditions(t *testing.T) {
entries := []ToolEntry{
{Name: "bash"},
{Name: "azure_mcp", When: map[string]string{
"language": "python",
"service":  "key-vault",
}},
}

// Both match
result := ResolveTools(entries, map[string]string{
"language": "python", "service": "key-vault",
})
if len(result) != 2 {
t.Errorf("expected 2 tools when both conditions match, got %d: %v", len(result), result)
}

// Only one condition matches
result = ResolveTools(entries, map[string]string{
"language": "python", "service": "identity",
})
if len(result) != 1 || result[0] != "bash" {
t.Errorf("expected [bash] when only one condition matches, got %v", result)
}
}

func TestResolveTools_EmptyProperties(t *testing.T) {
entries := []ToolEntry{
{Name: "create"},
{Name: "azure_mcp", When: map[string]string{"language": "python"}},
}
result := ResolveTools(entries, nil)
if len(result) != 1 || result[0] != "create" {
t.Errorf("expected [create] with nil properties, got %v", result)
}

result = ResolveTools(entries, map[string]string{})
if len(result) != 1 || result[0] != "create" {
t.Errorf("expected [create] with empty properties, got %v", result)
}
}

func TestResolveTools_AllConditionalNoneMatch(t *testing.T) {
entries := []ToolEntry{
{Name: "tool-a", When: map[string]string{"language": "python"}},
{Name: "tool-b", When: map[string]string{"language": "dotnet"}},
}
result := ResolveTools(entries, map[string]string{"language": "java"})
if len(result) != 0 {
t.Errorf("expected 0 tools, got %d: %v", len(result), result)
}
}

func TestMatchesWhen_EmptyWhen(t *testing.T) {
if !matchesWhen(nil, map[string]string{"a": "b"}) {
t.Error("nil when should always match")
}
if !matchesWhen(map[string]string{}, map[string]string{"a": "b"}) {
t.Error("empty when should always match")
}
}

func TestMatchesWhen_MissingProperty(t *testing.T) {
when := map[string]string{"language": "python"}
if matchesWhen(when, map[string]string{"service": "identity"}) {
t.Error("should not match when required property is missing")
}
}

func TestValidateToolEntry_Valid(t *testing.T) {
cases := []ToolEntry{
{Name: "bash"},
{Name: "mcp", When: map[string]string{"language": "python"}},
}
for i, tc := range cases {
if err := validateToolEntry(tc, "test", i); err != nil {
t.Errorf("case %d: unexpected error: %v", i, err)
}
}
}

func TestValidateToolEntry_MissingName(t *testing.T) {
err := validateToolEntry(ToolEntry{}, "test", 0)
if err == nil {
t.Fatal("expected error for missing name")
}
}

func TestParseConfigWithTools(t *testing.T) {
data := []byte(`
configs:
  - name: with-tools
    description: "Config with conditional tools"
    generator:
      model: "claude-opus-4.6"
      tools:
        - name: "bash"
        - name: "azure-mcp"
          when:
            language: python
      available_tools: ["view"]
      excluded_tools: ["dangerous"]
`)
cfg, err := Parse(data)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
c := cfg.Configs[0]
if len(c.Generator.Tools) != 2 {
t.Fatalf("expected 2 tool entries, got %d", len(c.Generator.Tools))
}
if c.Generator.Tools[0].Name != "bash" {
t.Errorf("expected first tool 'bash', got %q", c.Generator.Tools[0].Name)
}
if c.Generator.Tools[1].When["language"] != "python" {
t.Errorf("expected when language=python, got %v", c.Generator.Tools[1].When)
}
}

func TestResolveTools_Deduplication(t *testing.T) {
entries := []ToolEntry{
{Name: "create"},
{Name: "create", When: map[string]string{"language": "python"}},
{Name: "edit"},
}
result := ResolveTools(entries, map[string]string{"language": "python"})
if len(result) != 2 {
t.Fatalf("expected 2 tools after dedup, got %d: %v", len(result), result)
}
if result[0] != "create" || result[1] != "edit" {
t.Errorf("expected [create edit], got %v", result)
}
}

func TestResolveTools_DeduplicationPreservesOrder(t *testing.T) {
entries := []ToolEntry{
{Name: "bash"},
{Name: "create", When: map[string]string{"language": "python"}},
{Name: "bash", When: map[string]string{"language": "python"}},
{Name: "edit"},
}
result := ResolveTools(entries, map[string]string{"language": "python"})
want := []string{"bash", "create", "edit"}
if len(result) != len(want) {
t.Fatalf("expected %d tools, got %d: %v", len(want), len(result), result)
}
for i, name := range want {
if result[i] != name {
t.Errorf("result[%d] = %q, want %q", i, result[i], name)
}
}
}
