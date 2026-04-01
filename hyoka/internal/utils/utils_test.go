package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain json", `{"a":1}`, `{"a":1}`},
		{"with text", `Here is the result: {"a":1} done.`, `{"a":1}`},
		{"markdown fenced", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"no json", "hello world", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractJSON(tt.input)
			if got != tt.want {
				t.Errorf("ExtractJSON(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReadDirFilesFiltered_SkipsDirs(t *testing.T) {
	dir := t.TempDir()
	// Create a source file
	os.WriteFile(filepath.Join(dir, "main.py"), []byte("print('hello')"), 0644)
	// Create a dependency directory with files
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg", "index.js"), []byte("module.exports = {}"), 0644)
	// Create another dependency directory
	os.MkdirAll(filepath.Join(dir, "__pycache__"), 0755)
	os.WriteFile(filepath.Join(dir, "__pycache__", "main.cpython-311.pyc"), []byte("bytecode"), 0644)

	skipDirs := map[string]bool{"node_modules": true, "__pycache__": true}
	files, err := ReadDirFilesFiltered(dir, skipDirs)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), keys(files))
	}
	if _, ok := files["main.py"]; !ok {
		t.Error("expected main.py in result")
	}
}

func TestReadDirFilesFiltered_NilSkipDirsReadsAll(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.MkdirAll(filepath.Join(dir, "vendor", "lib"), 0755)
	os.WriteFile(filepath.Join(dir, "vendor", "lib", "dep.go"), []byte("package lib"), 0644)

	files, err := ReadDirFilesFiltered(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), keys(files))
	}
}

func TestReadDirFilesFiltered_NestedIgnoredDirs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('hi')"), 0644)
	// Nested node_modules (e.g., inside a subproject)
	os.MkdirAll(filepath.Join(dir, "sub", "node_modules", "dep"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "node_modules", "dep", "index.js"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "index.js"), []byte("require('./dep')"), 0644)

	skipDirs := map[string]bool{"node_modules": true}
	files, err := ReadDirFilesFiltered(dir, skipDirs)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files (app.js, sub/index.js), got %d: %v", len(files), keys(files))
	}
	if _, ok := files["app.js"]; !ok {
		t.Error("expected app.js in result")
	}
	if _, ok := files[filepath.Join("sub", "index.js")]; !ok {
		t.Error("expected sub/index.js in result")
	}
}

func TestReadDirFiles_BackwardCompatible(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "dep.js"), []byte("dep"), 0644)

	// ReadDirFiles (no filter) should include everything
	files, err := ReadDirFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), keys(files))
	}
}

func keys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
