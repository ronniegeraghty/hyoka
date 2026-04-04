package pidfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setTestDir overrides DirFn to point at the given directory and returns
// a cleanup function that restores the original.
func setTestDir(t *testing.T, dir string) {
	t.Helper()
	orig := DirFn
	DirFn = func() string { return dir }
	t.Cleanup(func() { DirFn = orig })
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	info := Info{PID: 12345, PromptID: "test-prompt", Config: "test-config"}
	if err := Write(info); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "12345.json"))
	if err != nil {
		t.Fatalf("reading PID file: %v", err)
	}
	var got Info
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshalling PID file: %v", err)
	}
	if got != info {
		t.Errorf("got %+v, want %+v", got, info)
	}
}

func TestWrite_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "pids")
	setTestDir(t, dir)

	if err := Write(Info{PID: 1}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "1.json")); err != nil {
		t.Fatalf("PID file not created: %v", err)
	}
}

func TestWrite_EmptyDir(t *testing.T) {
	setTestDir(t, "")

	if err := Write(Info{PID: 1}); err == nil {
		t.Fatal("Write() should fail when DirFn returns empty string")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	info := Info{PID: 999}
	if err := Write(info); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	path := filepath.Join(dir, "999.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("PID file should exist before Remove: %v", err)
	}

	Remove(999)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("PID file should be removed, got err = %v", err)
	}
}

func TestRemove_NonExistent(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)
	// Should not panic when file doesn't exist.
	Remove(99999)
}

func TestRemove_EmptyDir(t *testing.T) {
	setTestDir(t, "")
	// Should not panic when DirFn returns empty string.
	Remove(1)
}

func TestReadAlive_CurrentProcess(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	// Use our own PID — guaranteed to be alive.
	self := Info{PID: os.Getpid(), PromptID: "self", Config: "cfg"}
	if err := Write(self); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 1 {
		t.Fatalf("expected 1 alive entry, got %d", len(alive))
	}
	if alive[0] != self {
		t.Errorf("got %+v, want %+v", alive[0], self)
	}
}

func TestReadAlive_StaleProcess(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	// PID that almost certainly doesn't exist.
	stale := Info{PID: 2147483647, PromptID: "stale"}
	if err := Write(stale); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	pidPath := filepath.Join(dir, "2147483647.json")
	if _, err := os.Stat(pidPath); err != nil {
		t.Fatalf("stale PID file should exist before ReadAlive: %v", err)
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 0 {
		t.Errorf("expected 0 alive entries, got %d: %+v", len(alive), alive)
	}

	// Stale file should have been cleaned up.
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("stale PID file should be removed after ReadAlive")
	}
}

func TestReadAlive_MixedAliveAndStale(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	selfInfo := Info{PID: os.Getpid(), PromptID: "alive"}
	staleInfo := Info{PID: 2147483647, PromptID: "stale"}
	for _, info := range []Info{selfInfo, staleInfo} {
		if err := Write(info); err != nil {
			t.Fatalf("Write(%+v) error = %v", info, err)
		}
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 1 {
		t.Fatalf("expected 1 alive, got %d: %+v", len(alive), alive)
	}
	if alive[0].PromptID != "alive" {
		t.Errorf("expected alive entry, got %+v", alive[0])
	}

	// Stale file should be gone.
	stalePath := filepath.Join(dir, "2147483647.json")
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Error("stale PID file should be removed")
	}
}

func TestReadAlive_NonExistentDir(t *testing.T) {
	setTestDir(t, filepath.Join(t.TempDir(), "does-not-exist"))

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() should return nil for non-existent dir, got %v", err)
	}
	if alive != nil {
		t.Errorf("expected nil, got %+v", alive)
	}
}

func TestReadAlive_EmptyDir(t *testing.T) {
	setTestDir(t, "")

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if alive != nil {
		t.Errorf("expected nil for empty DirFn, got %+v", alive)
	}
}

func TestReadAlive_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	// Write garbage JSON into a .json file.
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 0 {
		t.Errorf("expected 0 alive entries for invalid JSON, got %d", len(alive))
	}
}

func TestReadAlive_SkipsNonJSON(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	// Write a non-.json file — should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write a valid alive PID file alongside it.
	selfInfo := Info{PID: os.Getpid(), PromptID: "me"}
	if err := Write(selfInfo); err != nil {
		t.Fatal(err)
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 1 {
		t.Fatalf("expected 1 alive entry, got %d", len(alive))
	}
}

func TestReadAlive_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	if err := os.Mkdir(filepath.Join(dir, "subdir.json"), 0o755); err != nil {
		t.Fatal(err)
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 0 {
		t.Errorf("expected 0, got %d", len(alive))
	}
}

func TestReadAlive_ZeroPIDTreatedAsStale(t *testing.T) {
	dir := t.TempDir()
	setTestDir(t, dir)

	data, _ := json.Marshal(Info{PID: 0, PromptID: "zero"})
	if err := os.WriteFile(filepath.Join(dir, "0.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	alive, err := ReadAlive()
	if err != nil {
		t.Fatalf("ReadAlive() error = %v", err)
	}
	if len(alive) != 0 {
		t.Errorf("PID 0 should not be alive, got %d entries", len(alive))
	}
}

func TestCleanAll(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "hyoka-pids")
	setTestDir(t, dir)

	if err := Write(Info{PID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("dir should exist after Write: %v", err)
	}

	if err := CleanAll(); err != nil {
		t.Fatalf("CleanAll() error = %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("dir should be removed after CleanAll")
	}
}

func TestCleanAll_EmptyDir(t *testing.T) {
	setTestDir(t, "")

	if err := CleanAll(); err != nil {
		t.Fatalf("CleanAll() should succeed with empty DirFn, got %v", err)
	}
}

func TestCleanAll_NonExistentDir(t *testing.T) {
	setTestDir(t, filepath.Join(t.TempDir(), "gone"))

	if err := CleanAll(); err != nil {
		t.Fatalf("CleanAll() should succeed for missing dir, got %v", err)
	}
}

func TestDefaultDir_XDGOverride(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/custom/state")
	got := defaultDir()
	want := filepath.Join("/custom/state", "copilot-cli", "hyoka-pids")
	if got != want {
		t.Errorf("defaultDir() = %q, want %q", got, want)
	}
}

func TestDefaultDir_FallbackHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	got := defaultDir()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	want := filepath.Join(home, ".copilot", "hyoka-pids")
	if got != want {
		t.Errorf("defaultDir() = %q, want %q", got, want)
	}
}
