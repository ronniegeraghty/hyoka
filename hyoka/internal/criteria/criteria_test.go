package criteria

import (
"io"
"log/slog"
"os"
"path/filepath"
"testing"
)

func TestMain(m *testing.M) {
slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1})))
os.Exit(m.Run())
}

func TestMatchesAll(t *testing.T) {
cs := CriteriaSet{When: map[string]string{"language": "java", "service": "keyvault"}}
props := map[string]string{"language": "java", "service": "keyvault", "plane": "data-plane"}
if !cs.Matches(props) {
t.Error("expected match")
}
}

func TestMatchesCaseInsensitive(t *testing.T) {
cs := CriteriaSet{When: map[string]string{"language": "Java"}}
props := map[string]string{"language": "java"}
if !cs.Matches(props) {
t.Error("expected case-insensitive match")
}
}

func TestMatchesNoMatch(t *testing.T) {
cs := CriteriaSet{When: map[string]string{"language": "python"}}
props := map[string]string{"language": "java"}
if cs.Matches(props) {
t.Error("expected no match")
}
}

func TestMatchesEmptyWhen(t *testing.T) {
cs := CriteriaSet{}
props := map[string]string{"language": "java", "service": "storage"}
if !cs.Matches(props) {
t.Error("empty when should match everything")
}
}

func TestMatchesPartialFields(t *testing.T) {
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
{"custom property match", map[string]string{"framework": "spring"}, map[string]string{"framework": "spring", "language": "java"}, true},
{"missing property no match", map[string]string{"framework": "spring"}, map[string]string{"language": "java"}, false},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
cs := CriteriaSet{When: tt.when}
if got := cs.Matches(tt.props); got != tt.matches {
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
criteria:
  - name: Builder Pattern
    description: SDK clients use builder pattern.
  - name: Try-With-Resources
    description: AutoCloseable clients use try-with-resources.
`), 0644)

kvFile := filepath.Join(dir, "service", "keyvault.yaml")
os.MkdirAll(filepath.Dir(kvFile), 0755)
os.WriteFile(kvFile, []byte(`
when:
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
os.WriteFile(filepath.Join(dir, "empty.yaml"), []byte("when:\n  language: go\ncriteria: []\n"), 0644)
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
When:     map[string]string{"language": "java"},
Criteria: []Criterion{{Name: "Builder Pattern"}, {Name: "Try-With-Resources"}},
},
{
When:     map[string]string{"service": "keyvault"},
Criteria: []Criterion{{Name: "Vault URI"}},
},
{
When:     map[string]string{"language": "python"},
Criteria: []Criterion{{Name: "Async Usage"}},
},
}

got := MatchingCriteria(sets, map[string]string{"language": "java", "service": "keyvault"})
if len(got) != 3 {
t.Fatalf("expected 3 matching criteria, got %d", len(got))
}

got = MatchingCriteria(sets, map[string]string{"language": "python", "service": "storage"})
if len(got) != 1 {
t.Fatalf("expected 1 matching criterion, got %d", len(got))
}

got = MatchingCriteria(sets, map[string]string{"language": "go", "service": "storage"})
if len(got) != 0 {
t.Errorf("expected 0 matching criteria, got %d", len(got))
}
}

func TestMatchingCriteriaEmptyWhen(t *testing.T) {
sets := []CriteriaSet{
{
Criteria: []Criterion{{Name: "Universal Rule"}},
},
{
When:     map[string]string{"language": "python"},
Criteria: []Criterion{{Name: "Python Only"}},
},
}

got := MatchingCriteria(sets, map[string]string{"language": "java"})
if len(got) != 1 || got[0].Name != "Universal Rule" {
t.Errorf("expected only universal rule, got %v", got)
}

got = MatchingCriteria(sets, map[string]string{"language": "python"})
if len(got) != 2 {
t.Errorf("expected 2 criteria (universal + python), got %d", len(got))
}
}
