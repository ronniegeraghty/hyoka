package config

import (
	"testing"
)

func TestResolveTools_NilGenerator(t *testing.T) {
	avail, excl := ResolveTools(nil, PromptProperties{})
	if avail != nil || excl != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", avail, excl)
	}
}

func TestResolveTools_LegacyOnly(t *testing.T) {
	gen := &GeneratorConfig{
		Model:          "gpt-4",
		AvailableTools: []string{"bash", "edit"},
		ExcludedTools:  []string{"web_fetch"},
	}
	avail, excl := ResolveTools(gen, PromptProperties{Language: "python"})
	assertSliceEqual(t, avail, []string{"bash", "edit"})
	assertSliceEqual(t, excl, []string{"web_fetch"})
}

func TestResolveTools_LegacyNilPreserved(t *testing.T) {
	gen := &GeneratorConfig{Model: "gpt-4"}
	avail, excl := ResolveTools(gen, PromptProperties{})
	if avail != nil {
		t.Errorf("expected nil available (all defaults), got %v", avail)
	}
	if excl != nil {
		t.Errorf("expected nil excluded, got %v", excl)
	}
}

func TestResolveTools_UnconditionalEntry(t *testing.T) {
	gen := &GeneratorConfig{
		Model: "gpt-4",
		Tools: []ToolEntry{
			{Name: "bash"},
			{Name: "edit"},
		},
	}
	avail, excl := ResolveTools(gen, PromptProperties{})
	assertSliceEqual(t, avail, []string{"bash", "edit"})
	if len(excl) != 0 {
		t.Errorf("expected no excluded tools, got %v", excl)
	}
}

func TestResolveTools_WhenMatches(t *testing.T) {
	gen := &GeneratorConfig{
		Model: "gpt-4",
		Tools: []ToolEntry{
			{Name: "bash"},
			{Name: "azure-mcp", When: map[string]string{"language": "python"}},
		},
	}

	// Python prompt → azure-mcp included.
	avail, _ := ResolveTools(gen, PromptProperties{Language: "python"})
	assertSliceEqual(t, avail, []string{"bash", "azure-mcp"})

	// Go prompt → azure-mcp excluded.
	avail, _ = ResolveTools(gen, PromptProperties{Language: "go"})
	assertSliceEqual(t, avail, []string{"bash"})
}

func TestResolveTools_WhenMultipleConditionsANDed(t *testing.T) {
	gen := &GeneratorConfig{
		Model: "gpt-4",
		Tools: []ToolEntry{
			{Name: "cosmosdb-mcp", When: map[string]string{
				"language": "python",
				"service":  "cosmos-db",
			}},
		},
	}

	// Both match → included.
	avail, _ := ResolveTools(gen, PromptProperties{Language: "python", Service: "cosmos-db"})
	assertSliceEqual(t, avail, []string{"cosmosdb-mcp"})

	// Only one matches → not included.
	avail, _ = ResolveTools(gen, PromptProperties{Language: "python", Service: "key-vault"})
	if len(avail) != 0 {
		t.Errorf("expected empty available, got %v", avail)
	}
}

func TestResolveTools_ExcludeWhenMatches(t *testing.T) {
	gen := &GeneratorConfig{
		Model: "gpt-4",
		Tools: []ToolEntry{
			{Name: "pip-tools", ExcludeWhen: map[string]string{"language": "go"}},
		},
	}

	// Go prompt → pip-tools excluded.
	_, excl := ResolveTools(gen, PromptProperties{Language: "go"})
	assertSliceEqual(t, excl, []string{"pip-tools"})

	// Python prompt → not excluded.
	_, excl = ResolveTools(gen, PromptProperties{Language: "python"})
	if len(excl) != 0 {
		t.Errorf("expected no excluded tools, got %v", excl)
	}
}

func TestResolveTools_MergeLegacyAndConditional(t *testing.T) {
	gen := &GeneratorConfig{
		Model:          "gpt-4",
		AvailableTools: []string{"bash"},
		ExcludedTools:  []string{"dangerous"},
		Tools: []ToolEntry{
			{Name: "edit"},
			{Name: "azure-mcp", When: map[string]string{"language": "python"}},
		},
	}
	avail, excl := ResolveTools(gen, PromptProperties{Language: "python"})
	assertSliceEqual(t, avail, []string{"bash", "edit", "azure-mcp"})
	assertSliceEqual(t, excl, []string{"dangerous"})
}

func TestResolveTools_ExcludedWinsOverAvailable(t *testing.T) {
	gen := &GeneratorConfig{
		Model: "gpt-4",
		Tools: []ToolEntry{
			{Name: "bash"},
			{Name: "bash", ExcludeWhen: map[string]string{"language": "go"}},
		},
	}
	avail, excl := ResolveTools(gen, PromptProperties{Language: "go"})

	// bash is in both → excluded wins, removed from available.
	assertSliceEqual(t, excl, []string{"bash"})
	if len(avail) != 0 {
		t.Errorf("expected bash removed from available, got %v", avail)
	}
}

func TestResolveTools_Deduplication(t *testing.T) {
	gen := &GeneratorConfig{
		Model:          "gpt-4",
		AvailableTools: []string{"bash", "edit"},
		Tools: []ToolEntry{
			{Name: "bash"}, // duplicate of legacy
			{Name: "edit"}, // duplicate of legacy
			{Name: "view"},
		},
	}
	avail, _ := ResolveTools(gen, PromptProperties{})
	assertSliceEqual(t, avail, []string{"bash", "edit", "view"})
}

func TestResolveTools_AllPropertyKeys(t *testing.T) {
	gen := &GeneratorConfig{
		Model: "gpt-4",
		Tools: []ToolEntry{
			{Name: "t1", When: map[string]string{"language": "python"}},
			{Name: "t2", When: map[string]string{"service": "identity"}},
			{Name: "t3", When: map[string]string{"plane": "data-plane"}},
			{Name: "t4", When: map[string]string{"category": "auth"}},
			{Name: "t5", When: map[string]string{"difficulty": "hard"}},
		},
	}

	props := PromptProperties{
		Language:   "python",
		Service:    "identity",
		Plane:      "data-plane",
		Category:   "auth",
		Difficulty: "hard",
	}
	avail, _ := ResolveTools(gen, props)
	assertSliceEqual(t, avail, []string{"t1", "t2", "t3", "t4", "t5"})
}

func TestValidateToolEntry_Valid(t *testing.T) {
	cases := []ToolEntry{
		{Name: "bash"},
		{Name: "mcp", When: map[string]string{"language": "python"}},
		{Name: "pip", ExcludeWhen: map[string]string{"language": "go"}},
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

func TestValidateToolEntry_BothWhenAndExcludeWhen(t *testing.T) {
	entry := ToolEntry{
		Name:        "bad",
		When:        map[string]string{"language": "python"},
		ExcludeWhen: map[string]string{"service": "identity"},
	}
	err := validateToolEntry(entry, "test", 0)
	if err == nil {
		t.Fatal("expected error for both when and exclude_when")
	}
}

func TestValidateToolEntry_UnrecognizedPropertyKey(t *testing.T) {
	entry := ToolEntry{
		Name: "bad",
		When: map[string]string{"unknown_field": "value"},
	}
	err := validateToolEntry(entry, "test", 0)
	if err == nil {
		t.Fatal("expected error for unrecognized property key")
	}
}

func TestValidateToolEntry_EmptyPropertyValue(t *testing.T) {
	entry := ToolEntry{
		Name: "bad",
		When: map[string]string{"language": ""},
	}
	err := validateToolEntry(entry, "test", 0)
	if err == nil {
		t.Fatal("expected error for empty property value")
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
        - name: "pip-tools"
          exclude_when:
            language: go
      available_tools: ["view"]
      excluded_tools: ["dangerous"]
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Configs[0]
	if len(c.Generator.Tools) != 3 {
		t.Fatalf("expected 3 tool entries, got %d", len(c.Generator.Tools))
	}
	if c.Generator.Tools[0].Name != "bash" {
		t.Errorf("expected first tool 'bash', got %q", c.Generator.Tools[0].Name)
	}
	if c.Generator.Tools[1].When["language"] != "python" {
		t.Errorf("expected when language=python, got %v", c.Generator.Tools[1].When)
	}
	if c.Generator.Tools[2].ExcludeWhen["language"] != "go" {
		t.Errorf("expected exclude_when language=go, got %v", c.Generator.Tools[2].ExcludeWhen)
	}
}

func TestParseConfigRejectsInvalidToolEntry(t *testing.T) {
	data := []byte(`
configs:
  - name: bad-tool
    description: "Invalid tool entry"
    generator:
      model: "gpt-4"
      tools:
        - name: "conflicting"
          when:
            language: python
          exclude_when:
            service: identity
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for tool with both when and exclude_when")
	}
}

func TestParseConfigRejectsUnknownToolProperty(t *testing.T) {
	data := []byte(`
configs:
  - name: bad-key
    description: "Unknown key in when"
    generator:
      model: "gpt-4"
      tools:
        - name: "bad"
          when:
            bogus_field: value
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for unrecognized property key")
	}
}

// assertSliceEqual compares two string slices for equality.
func assertSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("length mismatch: got %v (len %d), want %v (len %d)", got, len(got), want, len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
