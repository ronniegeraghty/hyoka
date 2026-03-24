package eval

import (
"os"
"path/filepath"
"testing"
)

func TestSnapshotDir_CapturesFilesAndDirs(t *testing.T) {
dir := t.TempDir()

// Create a file and a subdirectory
os.WriteFile(filepath.Join(dir, "hello.py"), []byte("print('hi')"), 0644)
os.Mkdir(filepath.Join(dir, "subdir"), 0755)
os.Mkdir(filepath.Join(dir, ".hidden"), 0755) // should be skipped

snap := snapshotDir(dir)
if snap == nil {
t.Fatal("snapshotDir returned nil")
}
if !snap["hello.py"] {
t.Error("snapshotDir should capture files")
}
if !snap["subdir"] {
t.Error("snapshotDir should capture directories")
}
if snap[".hidden"] {
t.Error("snapshotDir should skip hidden entries")
}
}

func TestRecoverMisplacedFiles_RecoversNewFiles(t *testing.T) {
home := t.TempDir()
workspace := t.TempDir()

// Pre-existing file
os.WriteFile(filepath.Join(home, "existing.txt"), []byte("old"), 0644)
snap := snapshotDir(home)

// Simulate agent creating a new file
os.WriteFile(filepath.Join(home, "main.py"), []byte("print('hello')"), 0644)

recovered := recoverMisplacedFiles(home, snap, workspace, "test", false)
if recovered != 1 {
t.Fatalf("expected 1 recovered, got %d", recovered)
}
// File should exist in workspace
if _, err := os.Stat(filepath.Join(workspace, "main.py")); err != nil {
t.Error("main.py should be in workspace")
}
// File should be removed from home
if _, err := os.Stat(filepath.Join(home, "main.py")); err == nil {
t.Error("main.py should be removed from home")
}
}

func TestRecoverMisplacedFiles_DeletesJunkDirs(t *testing.T) {
home := t.TempDir()
workspace := t.TempDir()

snap := snapshotDir(home)

// Simulate __pycache__ appearing
pycache := filepath.Join(home, "__pycache__")
os.Mkdir(pycache, 0755)
os.WriteFile(filepath.Join(pycache, "mod.cpython-311.pyc"), []byte{0}, 0644)

recovered := recoverMisplacedFiles(home, snap, workspace, "test", false)
if recovered != 1 {
t.Fatalf("expected 1 recovered (junk dir deleted), got %d", recovered)
}
if _, err := os.Stat(pycache); err == nil {
t.Error("__pycache__ should be deleted")
}
// Should NOT appear in workspace
if _, err := os.Stat(filepath.Join(workspace, "__pycache__")); err == nil {
t.Error("__pycache__ should not be moved to workspace")
}
}

func TestRecoverMisplacedFiles_MovesNewDirToWorkspace(t *testing.T) {
home := t.TempDir()
workspace := t.TempDir()

snap := snapshotDir(home)

// Simulate agent creating a project directory
projDir := filepath.Join(home, "myproject")
os.Mkdir(projDir, 0755)
os.WriteFile(filepath.Join(projDir, "app.py"), []byte("app"), 0644)

recovered := recoverMisplacedFiles(home, snap, workspace, "test", false)
if recovered != 1 {
t.Fatalf("expected 1 recovered, got %d", recovered)
}
// Directory should exist in workspace
if _, err := os.Stat(filepath.Join(workspace, "myproject", "app.py")); err != nil {
t.Error("myproject/app.py should be in workspace")
}
// Directory should be removed from home
if _, err := os.Stat(projDir); err == nil {
t.Error("myproject should be removed from home")
}
}

func TestRecoverMisplacedFiles_SkipsPreExistingDirs(t *testing.T) {
home := t.TempDir()
workspace := t.TempDir()

// Pre-existing directory
os.Mkdir(filepath.Join(home, "Documents"), 0755)
snap := snapshotDir(home)

recovered := recoverMisplacedFiles(home, snap, workspace, "test", false)
if recovered != 0 {
t.Fatalf("expected 0 recovered, got %d", recovered)
}
// Documents should still exist in home
if _, err := os.Stat(filepath.Join(home, "Documents")); err != nil {
t.Error("Documents should still exist in home")
}
}
