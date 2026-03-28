package criteria

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchRule_Matches(t *testing.T) {
	tests := []struct {
		name  string
		rule  MatchRule
		attrs PromptAttributes
		want  bool
	}{
		{
			name:  "empty rule matches everything",
			rule:  MatchRule{},
			attrs: PromptAttributes{Language: "java", Service: "keyvault"},
			want:  true,
		},
		{
			name:  "language match",
			rule:  MatchRule{Language: "java"},
			attrs: PromptAttributes{Language: "java", Service: "keyvault"},
			want:  true,
		},
		{
			name:  "language mismatch",
			rule:  MatchRule{Language: "python"},
			attrs: PromptAttributes{Language: "java"},
			want:  false,
		},
		{
			name:  "service match",
			rule:  MatchRule{Service: "keyvault"},
			attrs: PromptAttributes{Language: "java", Service: "keyvault"},
			want:  true,
		},
		{
			name:  "multi-field match",
			rule:  MatchRule{Language: "java", Service: "keyvault"},
			attrs: PromptAttributes{Language: "java", Service: "keyvault", Plane: "data-plane"},
			want:  true,
		},
		{
			name:  "multi-field partial mismatch",
			rule:  MatchRule{Language: "java", Service: "storage"},
			attrs: PromptAttributes{Language: "java", Service: "keyvault"},
			want:  false,
		},
		{
			name:  "plane match",
			rule:  MatchRule{Plane: "management-plane"},
			attrs: PromptAttributes{Language: "python", Plane: "management-plane"},
			want:  true,
		},
		{
			name:  "category match",
			rule:  MatchRule{Category: "provisioning"},
			attrs: PromptAttributes{Category: "provisioning"},
			want:  true,
		},
		{
			name:  "empty attrs against empty rule",
			rule:  MatchRule{},
			attrs: PromptAttributes{},
			want:  true,
		},
		{
			name:  "empty attrs against non-empty rule",
			rule:  MatchRule{Language: "java"},
			attrs: PromptAttributes{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.Matches(tt.attrs)
			if got != tt.want {
				t.Errorf("MatchRule.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadCriteriaSets(t *testing.T) {
	// Create temp dir with criteria files
	dir := t.TempDir()

	langDir := filepath.Join(dir, "language")
	if err := os.MkdirAll(langDir, 0o755); err != nil {
		t.Fatal(err)
	}

	javaYAML := `match:
  language: java
criteria:
  - name: Maven Dependency Declaration
    description: Generated code includes correct Maven dependencies.
  - name: Try-With-Resources
    description: AutoCloseable clients use try-with-resources.
`
	if err := os.WriteFile(filepath.Join(langDir, "java.yaml"), []byte(javaYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	pythonYAML := `match:
  language: python
criteria:
  - name: Context Manager Usage
    description: SDK clients use context managers (with statements).
`
	if err := os.WriteFile(filepath.Join(langDir, "python.yml"), []byte(pythonYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	svcDir := filepath.Join(dir, "service")
	if err := os.MkdirAll(svcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	kvYAML := `match:
  service: keyvault
criteria:
  - name: Key Vault URI Format
    description: Vault URI follows https://{vault-name}.vault.azure.net pattern.
`
	if err := os.WriteFile(filepath.Join(svcDir, "keyvault.yaml"), []byte(kvYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	sets, err := LoadCriteriaSets(dir)
	if err != nil {
		t.Fatalf("LoadCriteriaSets() error = %v", err)
	}

	if len(sets) != 3 {
		t.Fatalf("expected 3 criteria sets, got %d", len(sets))
	}

	// Count total criteria
	total := 0
	for _, s := range sets {
		total += len(s.Criteria)
	}
	if total != 4 {
		t.Errorf("expected 4 total criteria, got %d", total)
	}
}

func TestLoadCriteriaSets_NonexistentDir(t *testing.T) {
	sets, err := LoadCriteriaSets("/nonexistent/path/criteria")
	if err != nil {
		t.Fatalf("expected nil error for nonexistent dir, got %v", err)
	}
	if len(sets) != 0 {
		t.Errorf("expected 0 sets for nonexistent dir, got %d", len(sets))
	}
}

func TestLoadCriteriaSets_EmptyCriteria(t *testing.T) {
	dir := t.TempDir()
	emptyYAML := `match:
  language: rust
criteria: []
`
	if err := os.WriteFile(filepath.Join(dir, "empty.yaml"), []byte(emptyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	sets, err := LoadCriteriaSets(dir)
	if err != nil {
		t.Fatalf("LoadCriteriaSets() error = %v", err)
	}
	if len(sets) != 0 {
		t.Errorf("expected 0 sets (empty criteria skipped), got %d", len(sets))
	}
}

func TestLoadCriteriaSets_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Also add a valid file to ensure it's still loaded
	validYAML := `match:
  language: go
criteria:
  - name: Error Wrapping
    description: Errors are wrapped with fmt.Errorf.
`
	if err := os.WriteFile(filepath.Join(dir, "good.yaml"), []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	sets, err := LoadCriteriaSets(dir)
	if err != nil {
		t.Fatalf("LoadCriteriaSets() error = %v", err)
	}
	if len(sets) != 1 {
		t.Errorf("expected 1 valid set (bad file skipped), got %d", len(sets))
	}
}

func TestLoadCriteriaSets_IgnoresNonYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	sets, err := LoadCriteriaSets(dir)
	if err != nil {
		t.Fatalf("LoadCriteriaSets() error = %v", err)
	}
	if len(sets) != 0 {
		t.Errorf("expected 0 sets (non-YAML ignored), got %d", len(sets))
	}
}

func TestMatchCriteria(t *testing.T) {
	sets := []CriteriaSet{
		{
			Match:    MatchRule{Language: "java"},
			Criteria: []Criterion{{Name: "Java Check 1"}, {Name: "Java Check 2"}},
			Source:   "language/java.yaml",
		},
		{
			Match:    MatchRule{Language: "python"},
			Criteria: []Criterion{{Name: "Python Check"}},
			Source:   "language/python.yaml",
		},
		{
			Match:    MatchRule{Service: "keyvault"},
			Criteria: []Criterion{{Name: "KV Check"}},
			Source:   "service/keyvault.yaml",
		},
		{
			Match:    MatchRule{Language: "java", Service: "keyvault"},
			Criteria: []Criterion{{Name: "Java+KV Check"}},
			Source:   "combined/java-keyvault.yaml",
		},
	}

	tests := []struct {
		name      string
		attrs     PromptAttributes
		wantCount int
		wantNames []string
	}{
		{
			name:      "java keyvault gets 4 criteria",
			attrs:     PromptAttributes{Language: "java", Service: "keyvault"},
			wantCount: 4,
			wantNames: []string{"Java Check 1", "Java Check 2", "KV Check", "Java+KV Check"},
		},
		{
			name:      "python keyvault gets 2 criteria",
			attrs:     PromptAttributes{Language: "python", Service: "keyvault"},
			wantCount: 2,
			wantNames: []string{"Python Check", "KV Check"},
		},
		{
			name:      "java storage gets 2 criteria",
			attrs:     PromptAttributes{Language: "java", Service: "storage"},
			wantCount: 2,
			wantNames: []string{"Java Check 1", "Java Check 2"},
		},
		{
			name:      "go storage gets 0 criteria",
			attrs:     PromptAttributes{Language: "go", Service: "storage"},
			wantCount: 0,
		},
		{
			name:      "empty attrs gets 0 criteria",
			attrs:     PromptAttributes{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := MatchCriteria(sets, tt.attrs)
			if len(matched) != tt.wantCount {
				t.Errorf("MatchCriteria() returned %d criteria, want %d", len(matched), tt.wantCount)
			}
			for _, wantName := range tt.wantNames {
				found := false
				for _, c := range matched {
					if c.Name == wantName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("MatchCriteria() missing expected criterion %q", wantName)
				}
			}
		})
	}
}

func TestMatchCriteria_NilSets(t *testing.T) {
	matched := MatchCriteria(nil, PromptAttributes{Language: "java"})
	if len(matched) != 0 {
		t.Errorf("expected 0 criteria from nil sets, got %d", len(matched))
	}
}

func TestFormatCriteria(t *testing.T) {
	criteria := []Criterion{
		{Name: "Maven Deps", Description: "Correct Maven dependencies."},
		{Name: "Try-With-Resources", Description: "AutoCloseable clients use try-with-resources."},
	}

	result := FormatCriteria(criteria)

	if result == "" {
		t.Fatal("expected non-empty formatted criteria")
	}

	// Check numbered list format
	if !containsStr(result, "1. **Maven Deps**") {
		t.Error("expected numbered criterion 1")
	}
	if !containsStr(result, "2. **Try-With-Resources**") {
		t.Error("expected numbered criterion 2")
	}
	if !containsStr(result, "Correct Maven dependencies") {
		t.Error("expected description in output")
	}
}

func TestFormatCriteria_Empty(t *testing.T) {
	result := FormatCriteria(nil)
	if result != "" {
		t.Errorf("expected empty string for nil criteria, got %q", result)
	}
}

func TestFormatCriteria_NoDescription(t *testing.T) {
	criteria := []Criterion{{Name: "Simple Check"}}
	result := FormatCriteria(criteria)
	if !containsStr(result, "1. **Simple Check**") {
		t.Errorf("expected criterion name in output, got %q", result)
	}
	if containsStr(result, "—") {
		t.Error("should not contain dash separator when description is empty")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && len(s) > 0 && findSubstring(s, sub)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
