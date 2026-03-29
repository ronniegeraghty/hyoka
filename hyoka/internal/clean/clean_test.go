package clean

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanSessionsDryRun(t *testing.T) {
	// Create fake session-state directory
	dir := t.TempDir()
	sessDir := filepath.Join(dir, "session-state")
	os.MkdirAll(filepath.Join(sessDir, "sess-1"), 0o755)
	os.WriteFile(filepath.Join(sessDir, "sess-1", "workspace.yaml"), []byte("cwd: /tmp/hyoka-gen-abc"), 0o644)

	var buf bytes.Buffer
	// Override state dir
	origFn := copilotStateDirFn
	copilotStateDirFn = func() string { return dir }
	defer func() { copilotStateDirFn = origFn }()

	result, err := Run(Options{DryRun: true, All: false, Out: &buf})
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionsFound != 1 {
		t.Errorf("expected 1 session found, got %d", result.SessionsFound)
	}
	if result.SessionsRemoved != 0 {
		t.Errorf("expected 0 removed in dry run, got %d", result.SessionsRemoved)
	}
	// Directory should still exist
	if _, err := os.Stat(filepath.Join(sessDir, "sess-1")); err != nil {
		t.Error("session dir should still exist in dry run")
	}
}

func TestCleanSessionsRemovesHyokaSessions(t *testing.T) {
	dir := t.TempDir()
	sessDir := filepath.Join(dir, "session-state")

	// Hyoka session
	os.MkdirAll(filepath.Join(sessDir, "hyoka-sess"), 0o755)
	os.WriteFile(filepath.Join(sessDir, "hyoka-sess", "workspace.yaml"),
		[]byte("cwd: /home/user/projects/hyoka/reports/20240101"), 0o644)
	os.WriteFile(filepath.Join(sessDir, "hyoka-sess", "events.json"),
		[]byte(`[{"type":"test"}]`), 0o644)

	// Non-hyoka session
	os.MkdirAll(filepath.Join(sessDir, "other-sess"), 0o755)
	os.WriteFile(filepath.Join(sessDir, "other-sess", "workspace.yaml"),
		[]byte("cwd: /home/user/myproject"), 0o644)

	origFn := copilotStateDirFn
	copilotStateDirFn = func() string { return dir }
	defer func() { copilotStateDirFn = origFn }()

	var buf bytes.Buffer
	result, err := Run(Options{All: false, Out: &buf})
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionsFound != 1 {
		t.Errorf("expected 1 hyoka session, got %d", result.SessionsFound)
	}
	if result.SessionsRemoved != 1 {
		t.Errorf("expected 1 removed, got %d", result.SessionsRemoved)
	}
	// Hyoka session should be gone
	if _, err := os.Stat(filepath.Join(sessDir, "hyoka-sess")); !os.IsNotExist(err) {
		t.Error("hyoka session should have been removed")
	}
	// Non-hyoka session should remain
	if _, err := os.Stat(filepath.Join(sessDir, "other-sess")); err != nil {
		t.Error("non-hyoka session should remain")
	}
}

func TestCleanSessionsAllFlag(t *testing.T) {
	dir := t.TempDir()
	sessDir := filepath.Join(dir, "session-state")

	os.MkdirAll(filepath.Join(sessDir, "sess-a"), 0o755)
	os.WriteFile(filepath.Join(sessDir, "sess-a", "workspace.yaml"),
		[]byte("cwd: /home/user/myproject"), 0o644)

	os.MkdirAll(filepath.Join(sessDir, "sess-b"), 0o755)
	os.WriteFile(filepath.Join(sessDir, "sess-b", "data.txt"),
		[]byte("data"), 0o644)

	origFn := copilotStateDirFn
	copilotStateDirFn = func() string { return dir }
	defer func() { copilotStateDirFn = origFn }()

	var buf bytes.Buffer
	result, err := Run(Options{All: true, Out: &buf})
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionsRemoved != 2 {
		t.Errorf("expected 2 removed with --all, got %d", result.SessionsRemoved)
	}
}

func TestCleanLogsTrimming(t *testing.T) {
	dir := t.TempDir()
	logsDir := filepath.Join(dir, "logs")
	os.MkdirAll(logsDir, 0o755)

	// Create 5 log files with different mod times
	for i := 0; i < 5; i++ {
		path := filepath.Join(logsDir, filepath.Base(t.Name())+"-log-"+string(rune('a'+i))+".log")
		os.WriteFile(path, []byte("log data"), 0o644)
		// Stagger modification times
		modTime := time.Now().Add(-time.Duration(5-i) * time.Hour)
		os.Chtimes(path, modTime, modTime)
	}

	origFn := copilotStateDirFn
	copilotStateDirFn = func() string { return dir }
	defer func() { copilotStateDirFn = origFn }()

	var buf bytes.Buffer
	result, err := Run(Options{KeepLogs: 2, Out: &buf})
	if err != nil {
		t.Fatal(err)
	}
	if result.LogsRemoved != 3 {
		t.Errorf("expected 3 logs removed (keep 2 of 5), got %d", result.LogsRemoved)
	}

	// Count remaining log files
	remaining, _ := os.ReadDir(logsDir)
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining logs, got %d", len(remaining))
	}
}

func TestCleanNoStateDir(t *testing.T) {
	origFn := copilotStateDirFn
	copilotStateDirFn = func() string { return "/nonexistent/path" }
	defer func() { copilotStateDirFn = origFn }()

	var buf bytes.Buffer
	result, err := Run(Options{Out: &buf})
	if err != nil {
		t.Fatal(err)
	}
	// Should succeed with zero results
	if result.SessionsFound != 0 || result.LogsRemoved != 0 {
		t.Errorf("expected zero results for missing dir, got sessions=%d logs=%d",
			result.SessionsFound, result.LogsRemoved)
	}
}

func TestIsHyokaSession(t *testing.T) {
	tests := []struct {
		name    string
		content string
		expect  bool
	}{
		{"hyoka-gen workspace", "cwd: /tmp/hyoka-gen-abc123", true},
		{"reports dir", "cwd: /home/user/projects/hyoka/reports/20240101", true},
		{"hyoka-config", "configDir: /tmp/hyoka-config-xyz", true},
		{"random project", "cwd: /home/user/myproject", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.content != "" {
				os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(tc.content), 0o644)
			}
			got := isHyokaSession(dir)
			if got != tc.expect {
				t.Errorf("isHyokaSession(%q) = %v, want %v", tc.content, got, tc.expect)
			}
		})
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		input  int64
		expect string
	}{
		{0, "0B"},
		{500, "500B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
		{4831838208, "4.5GB"},
	}
	for _, tc := range tests {
		got := humanBytes(tc.input)
		if got != tc.expect {
			t.Errorf("humanBytes(%d) = %q, want %q", tc.input, got, tc.expect)
		}
	}
}
