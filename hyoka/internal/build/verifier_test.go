package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
tests := []struct {
input    string
expected string
isNil    bool
}{
{"dotnet", "dotnet", false},
{"go", "go", false},
{"python", "python", false},
{"typescript", "typescript", false},
{"javascript", "javascript", false},
{"rust", "rust", false},
{"java-maven", "java-maven", false},
{"cpp", "cpp", false},
// Aliases
{"csharp", "dotnet", false},
{"c#", "dotnet", false},
{"py", "python", false},
{"golang", "go", false},
{"ts", "typescript", false},
{"js", "javascript", false},
{"c++", "cpp", false},
// Unknown
{"unknown-lang", "", true},
}

for _, tt := range tests {
t.Run(tt.input, func(t *testing.T) {
lc := DetectLanguage(tt.input)
if tt.isNil {
if lc != nil {
t.Errorf("expected nil for %q, got %+v", tt.input, lc)
}
return
}
if lc == nil {
t.Fatalf("expected non-nil for %q", tt.input)
}
if lc.Name != tt.expected {
t.Errorf("expected name %q, got %q", tt.expected, lc.Name)
}
})
}
}

func TestBuildCommands_PythonSkipsDependencyDirs(t *testing.T) {
	dir := t.TempDir()
	// Source file
	os.WriteFile(filepath.Join(dir, "main.py"), []byte("print('hello')"), 0644)
	// Dependency dirs with .py files that should be skipped
	os.MkdirAll(filepath.Join(dir, "venv", "lib"), 0755)
	os.WriteFile(filepath.Join(dir, "venv", "lib", "site.py"), []byte("import os"), 0644)
	os.MkdirAll(filepath.Join(dir, "__pycache__"), 0755)
	os.WriteFile(filepath.Join(dir, "__pycache__", "main.cpython-311.pyc"), []byte("bytecode"), 0644)

	lc := DetectLanguage("python")
	steps := buildCommands(lc, dir)
	if len(steps) != 1 {
		t.Fatalf("expected 1 build step, got %d", len(steps))
	}
	// Should only have main.py, not venv or __pycache__ files
	args := steps[0].Args
	pyFileCount := 0
	for _, a := range args {
		if strings.HasSuffix(a, ".py") || strings.HasSuffix(a, ".pyc") {
			pyFileCount++
			if strings.Contains(a, "venv") || strings.Contains(a, "__pycache__") {
				t.Errorf("dependency file should be excluded: %s", a)
			}
		}
	}
	if pyFileCount != 1 {
		t.Errorf("expected 1 python file, got %d in args: %v", pyFileCount, args)
	}
}

func TestBuildCommands_JavaScriptSkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	// Source file
	os.WriteFile(filepath.Join(dir, "index.js"), []byte("console.log('hi')"), 0644)
	// node_modules
	os.MkdirAll(filepath.Join(dir, "node_modules", "express"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "express", "index.js"), []byte("module.exports = {}"), 0644)

	lc := DetectLanguage("javascript")
	steps := buildCommands(lc, dir)
	if len(steps) != 1 {
		t.Fatalf("expected 1 build step, got %d", len(steps))
	}
	args := steps[0].Args
	jsFileCount := 0
	for _, a := range args {
		if strings.HasSuffix(a, ".js") {
			jsFileCount++
			if strings.Contains(a, "node_modules") {
				t.Errorf("node_modules file should be excluded: %s", a)
			}
		}
	}
	if jsFileCount != 1 {
		t.Errorf("expected 1 js file, got %d in args: %v", jsFileCount, args)
	}
}

func TestSupportedLanguages(t *testing.T) {
langs := SupportedLanguages()
if len(langs) == 0 {
t.Fatal("expected at least one supported language")
}
names := make(map[string]bool)
for _, l := range langs {
if l.Name == "" {
t.Error("language config has empty name")
}
if l.BuildCmd == "" {
t.Errorf("language %q has empty build command", l.Name)
}
if len(l.Extensions) == 0 {
t.Errorf("language %q has no extensions", l.Name)
}
names[l.Name] = true
}
// Check some expected languages exist
for _, expected := range []string{"dotnet", "go", "python", "typescript", "rust"} {
if !names[expected] {
t.Errorf("expected language %q not found", expected)
}
}
}
