package criteria

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchConditionMatchesAll(t *testing.T) {
	m := MatchCondition{Language: "java", Service: "keyvault"}
	attrs := PromptAttrs{Language: "java", Service: "keyvault", Plane: "data-plane"}
	if !m.Matches(attrs) {
		t.Error("expected match")
	}
}

func TestMatchConditionCaseInsensitive(t *testing.T) {
	m := MatchCondition{Language: "Java"}
	attrs := PromptAttrs{Language: "java"}
	if !m.Matches(attrs) {
		t.Error("expected case-insensitive match")
	}
}

func TestMatchConditionNoMatch(t *testing.T) {
	m := MatchCondition{Language: "python"}
	attrs := PromptAttrs{Language: "java"}
	if m.Matches(attrs) {
		t.Error("expected no match")
	}
}

func TestMatchConditionEmpty(t *testing.T) {
	m := MatchCondition{}
	attrs := PromptAttrs{Language: "java", Service: "storage"}
	if !m.Matches(attrs) {
		t.Error("empty match condition should match everything")
	}
}

func TestMatchConditionPartialFields(t *testing.T) {
	tests := []struct {
		name    string
		cond    MatchCondition
		attrs   PromptAttrs
		matches bool
	}{
		{"service only match", MatchCondition{Service: "keyvault"}, PromptAttrs{Service: "keyvault", Language: "go"}, true},
		{"service only no match", MatchCondition{Service: "keyvault"}, PromptAttrs{Service: "storage", Language: "go"}, false},
		{"plane match", MatchCondition{Plane: "data-plane"}, PromptAttrs{Plane: "data-plane"}, true},
		{"category match", MatchCondition{Category: "auth"}, PromptAttrs{Category: "auth"}, true},
		{"sdk match", MatchCondition{SDK: "azure-identity"}, PromptAttrs{SDK: "azure-identity"}, true},
		{"multi-field partial fail", MatchCondition{Language: "java", Service: "storage"}, PromptAttrs{Language: "java", Service: "keyvault"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cond.Matches(tt.attrs); got != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, got)
			}
		})
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()

	javaFile := filepath.Join(dir, "language", "java.yaml")
	os.MkdirAll(filepath.Dir(javaFile), 0755)
	os.WriteFile(javaFile, []byte(`
match:
  language: java
criteria:
  - name: Builder Pattern
    description: SDK clients use builder pattern.
  - name: Try-With-Resources
    description: AutoCloseable clients use try-with-resources.
`), 0644)

	kvFile := filepath.Join(dir, "service", "keyvault.yaml")
	os.MkdirAll(filepath.Dir(kvFile), 0755)
	os.WriteFile(kvFile, []byte(`
match:
  service: keyvault
criteria:
  - name: Vault URI Format
    description: Uses parameterized vault URI.
`), 0644)

	sets, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets) != 2 {
		t.Fatalf("expected 2 criteria sets, got %d", len(sets))
	}
}

func TestLoadDirSkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("not: valid: yaml: ["), 0644)
	os.WriteFile(filepath.Join(dir, "empty.yaml"), []byte("match:\n  language: go\ncriteria: []\n"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a yaml file"), 0644)

	sets, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets) != 0 {
		t.Errorf("expected 0 valid sets, got %d", len(sets))
	}
}

func TestLoadDirNonexistent(t *testing.T) {
	_, err := LoadDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestMatchingCriteria(t *testing.T) {
	sets := []CriteriaSet{
		{
			Match:    MatchCondition{Language: "java"},
			Criteria: []Criterion{{Name: "Builder Pattern"}, {Name: "Try-With-Resources"}},
		},
		{
			Match:    MatchCondition{Service: "keyvault"},
			Criteria: []Criterion{{Name: "Vault URI"}},
		},
		{
			Match:    MatchCondition{Language: "python"},
			Criteria: []Criterion{{Name: "Async Usage"}},
		},
	}

	got := MatchingCriteria(sets, PromptAttrs{Language: "java", Service: "keyvault"})
	if len(got) != 3 {
		t.Fatalf("expected 3 matching criteria, got %d", len(got))
	}

	got = MatchingCriteria(sets, PromptAttrs{Language: "python", Service: "storage"})
	if len(got) != 1 {
		t.Fatalf("expected 1 matching criterion, got %d", len(got))
	}

	got = MatchingCriteria(sets, PromptAttrs{Language: "go", Service: "storage"})
	if len(got) != 0 {
		t.Errorf("expected 0 matching criteria, got %d", len(got))
	}
}

func TestFormatCriteria(t *testing.T) {
	criteria := []Criterion{
		{Name: "Builder Pattern", Description: "Use builder pattern for clients."},
		{Name: "Error Handling"},
	}
	result := FormatCriteria(criteria)
	if !strings.Contains(result, "1. **Builder Pattern**") {
		t.Errorf("expected formatted criterion name, got %q", result)
	}
	if !strings.Contains(result, "Use builder pattern") {
		t.Errorf("expected description, got %q", result)
	}
	if !strings.Contains(result, "2. **Error Handling**") {
		t.Errorf("expected second criterion, got %q", result)
	}
}

func TestFormatCriteriaEmpty(t *testing.T) {
	if got := FormatCriteria(nil); got != "" {
		t.Errorf("expected empty string for nil criteria, got %q", got)
	}
}

func TestMergeCriteria(t *testing.T) {
	tier2 := []Criterion{
		{Name: "Builder Pattern", Description: "Use builder."},
	}
	tier3 := "- Uses correct authentication method"

	result := MergeCriteria(tier2, tier3)
	if !strings.Contains(result, "Attribute-Matched") {
		t.Error("expected Attribute-Matched header")
	}
	if !strings.Contains(result, "Prompt-Specific") {
		t.Error("expected Prompt-Specific header")
	}
	if !strings.Contains(result, "Builder Pattern") {
		t.Error("expected tier 2 criterion")
	}
	if !strings.Contains(result, "authentication method") {
		t.Error("expected tier 3 text")
	}
}

func TestMergeCriteriaTier2Only(t *testing.T) {
	tier2 := []Criterion{{Name: "Test"}}
	result := MergeCriteria(tier2, "")
	if !strings.Contains(result, "Attribute-Matched") {
		t.Error("expected tier 2 content")
	}
	if strings.Contains(result, "Prompt-Specific") {
		t.Error("should not contain prompt-specific header when tier3 is empty")
	}
}

func TestMergeCriteriaTier3Only(t *testing.T) {
	result := MergeCriteria(nil, "some criteria")
	if strings.Contains(result, "Attribute-Matched") {
		t.Error("should not contain attribute-matched header when tier2 is empty")
	}
	if !strings.Contains(result, "Prompt-Specific") {
		t.Error("expected prompt-specific header")
	}
}

func TestMergeCriteriaBothEmpty(t *testing.T) {
	result := MergeCriteria(nil, "")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}
