package criteria

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1})))
	os.Exit(m.Run())
}

func TestMatchesWhenAllFields(t *testing.T) {
	when := map[string]string{"language": "java", "service": "keyvault"}
	props := map[string]string{"language": "java", "service": "keyvault", "plane": "data-plane"}
	if !matchesWhen(when, props) {
		t.Error("expected match")
	}
}

func TestMatchesWhenCaseInsensitive(t *testing.T) {
	when := map[string]string{"language": "Java"}
	props := map[string]string{"language": "java"}
	if !matchesWhen(when, props) {
		t.Error("expected case-insensitive match")
	}
}

func TestMatchesWhenNoMatch(t *testing.T) {
	when := map[string]string{"language": "python"}
	props := map[string]string{"language": "java"}
	if matchesWhen(when, props) {
		t.Error("expected no match")
	}
}

func TestMatchesWhenEmpty(t *testing.T) {
	when := map[string]string{}
	props := map[string]string{"language": "java", "service": "storage"}
	if !matchesWhen(when, props) {
		t.Error("empty when map should match everything")
	}
}

func TestMatchesWhenPartialFields(t *testing.T) {
	tests := []struct {
		name    string
		when    map[string]string
		props   map[string]string
		matches bool
	}{
		{"service only match", map[string]string{"service": "keyvault"}, map[string]string{"service": "keyvault", "language": "go"}, true},
		{"service only no match", map[string]string{"service": "keyvault"}, map[string]string{"service": "storage", "language": "go"}, false},
		{"plane match", map[string]string{"plane": "data-plane"}, map[string]string{"plane": "data-plane"}, true},
		{"category match", map[string]string{"category": "auth"}, map[string]string{"category": "auth"}, true},
		{"sdk match", map[string]string{"sdk": "azure-identity"}, map[string]string{"sdk": "azure-identity"}, true},
		{"multi-field partial fail", map[string]string{"language": "java", "service": "storage"}, map[string]string{"language": "java", "service": "keyvault"}, false},
		{"missing prop key", map[string]string{"language": "go"}, map[string]string{"service": "storage"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesWhen(tt.when, tt.props); got != tt.matches {
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
when:
  language: java
graders:
  - name: Builder Pattern
    weight: 1.0
    prompt: SDK clients use builder pattern.
  - name: Try-With-Resources
    weight: 1.0
    prompt: AutoCloseable clients use try-with-resources.
`), 0644)

	kvFile := filepath.Join(dir, "service", "keyvault.yaml")
	os.MkdirAll(filepath.Dir(kvFile), 0755)
	os.WriteFile(kvFile, []byte(`
when:
  service: keyvault
graders:
  - name: Vault URI Format
    weight: 1.0
    prompt: Uses parameterized vault URI.
`), 0644)

	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 grader configs, got %d", len(configs))
	}
}

func TestLoadDirSkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("not: valid: yaml: ["), 0644)
	os.WriteFile(filepath.Join(dir, "empty.yaml"), []byte("when:\n  language: go\ngraders: []\n"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a yaml file"), 0644)

	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 valid configs, got %d", len(configs))
	}
}

func TestLoadDirNonexistent(t *testing.T) {
	_, err := LoadDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestMatchingGraders(t *testing.T) {
	configs := []GraderConfig{
		{
			When:    map[string]string{"language": "java"},
			Graders: []GraderEntry{{Name: "Builder Pattern", Weight: 1.0}, {Name: "Try-With-Resources", Weight: 1.0}},
		},
		{
			When:    map[string]string{"service": "keyvault"},
			Graders: []GraderEntry{{Name: "Vault URI", Weight: 1.0}},
		},
		{
			When:    map[string]string{"language": "python"},
			Graders: []GraderEntry{{Name: "Async Usage", Weight: 1.0}},
		},
	}

	got := MatchingGraders(configs, map[string]string{"language": "java", "service": "keyvault"})
	if len(got) != 3 {
		t.Fatalf("expected 3 matching graders, got %d", len(got))
	}

	got = MatchingGraders(configs, map[string]string{"language": "python", "service": "storage"})
	if len(got) != 1 {
		t.Fatalf("expected 1 matching grader, got %d", len(got))
	}

	got = MatchingGraders(configs, map[string]string{"language": "go", "service": "storage"})
	if len(got) != 0 {
		t.Errorf("expected 0 matching graders, got %d", len(got))
	}
}

func TestFormatGraders(t *testing.T) {
	graders := []GraderEntry{
		{Name: "Builder Pattern", Weight: 1.0, Prompt: "Use builder pattern for clients."},
		{Name: "Error Handling", Weight: 1.0},
	}
	result := FormatGraders(graders)
	if !strings.Contains(result, "1. **Builder Pattern**") {
		t.Errorf("expected formatted grader name, got %q", result)
	}
	if !strings.Contains(result, "Use builder pattern") {
		t.Errorf("expected prompt text, got %q", result)
	}
	if !strings.Contains(result, "2. **Error Handling**") {
		t.Errorf("expected second grader, got %q", result)
	}
}

func TestFormatGradersEmpty(t *testing.T) {
	if got := FormatGraders(nil); got != "" {
		t.Errorf("expected empty string for nil graders, got %q", got)
	}
}

func TestMergeCriteria(t *testing.T) {
	graders := []GraderEntry{
		{Name: "Builder Pattern", Weight: 1.0, Prompt: "Use builder."},
	}
	promptCriteria := "- Uses correct authentication method"

	result := MergeCriteria(graders, promptCriteria)
	if !strings.Contains(result, "Attribute-Matched") {
		t.Error("expected Attribute-Matched header")
	}
	if !strings.Contains(result, "Prompt-Specific") {
		t.Error("expected Prompt-Specific header")
	}
	if !strings.Contains(result, "Builder Pattern") {
		t.Error("expected grader criterion")
	}
	if !strings.Contains(result, "authentication method") {
		t.Error("expected prompt criteria text")
	}
}

func TestMergeCriteriaGradersOnly(t *testing.T) {
	graders := []GraderEntry{{Name: "Test", Weight: 1.0}}
	result := MergeCriteria(graders, "")
	if !strings.Contains(result, "Attribute-Matched") {
		t.Error("expected grader content")
	}
	if strings.Contains(result, "Prompt-Specific") {
		t.Error("should not contain prompt-specific header when prompt criteria is empty")
	}
}

func TestMergeCriteriaPromptOnly(t *testing.T) {
	result := MergeCriteria(nil, "some criteria")
	if strings.Contains(result, "Attribute-Matched") {
		t.Error("should not contain attribute-matched header when graders is empty")
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

func TestGraderWeightPreserved(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.yaml"), []byte(`
when:
  language: go
graders:
  - name: Critical Check
    weight: 2.0
    prompt: Very important check.
  - name: Minor Check
    weight: 0.5
    prompt: Less important.
`), 0644)

	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Graders[0].Weight != 2.0 {
		t.Errorf("expected weight 2.0, got %f", configs[0].Graders[0].Weight)
	}
	if configs[0].Graders[1].Weight != 0.5 {
		t.Errorf("expected weight 0.5, got %f", configs[0].Graders[1].Weight)
	}
}
