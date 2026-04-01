package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronniegeraghty/hyoka/internal/prompt"
)

func TestWriteReport(t *testing.T) {
	dir := t.TempDir()

	r := &EvalReport{
		PromptID:       "test-prompt",
		ConfigName:     "baseline",
		Timestamp:      "2024-01-15T10:00:00Z",
		Duration:       12.5,
		PromptMeta:     map[string]any{"service": "storage", "language": "dotnet"},
		ConfigUsed:     map[string]any{"name": "baseline", "model": "gpt-4"},
		GeneratedFiles: []string{"Program.cs", "Storage.csproj"},
		EventCount: 15,
		ToolCalls:  []string{"create_file", "edit_file"},
		Success:    true,
	}

	p := &prompt.Prompt{
		ID:       "test-prompt",
		Service:  "storage",
		Plane:    "data-plane",
		Language: "dotnet",
		Category: "authentication",
	}

	reportPath, err := WriteReport(r, dir, "20240115-100000", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("report file does not exist: %v", err)
	}

	// Verify JSON is valid
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	var parsed EvalReport
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON in report: %v", err)
	}

	if parsed.PromptID != "test-prompt" {
		t.Errorf("expected prompt ID 'test-prompt', got %q", parsed.PromptID)
	}
	if !parsed.Success {
		t.Error("expected success to be true")
	}

	// Verify directory structure — includes prompt ID for workspace isolation
	expectedDir := filepath.Join(dir, "20240115-100000", "results", "storage", "data-plane", "dotnet", "authentication", "test-prompt", "baseline")
	if _, err := os.Stat(expectedDir); err != nil {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestWriteSummary(t *testing.T) {
	dir := t.TempDir()

	s := &RunSummary{
		RunID:        "20240115-100000",
		Timestamp:    "2024-01-15T10:00:00Z",
		TotalPrompts: 5,
		TotalConfigs: 2,
		TotalEvals:   10,
		Passed:       8,
		Failed:       1,
		Errors:       1,
		Duration:     120.5,
		Reports:      []string{"/path/to/report1.json", "/path/to/report2.json"},
	}

	summaryPath, err := WriteSummary(s, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("failed to read summary: %v", err)
	}

	var parsed RunSummary
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON in summary: %v", err)
	}

	if parsed.TotalEvals != 10 {
		t.Errorf("expected 10 total evals, got %d", parsed.TotalEvals)
	}
	if parsed.Passed != 8 {
		t.Errorf("expected 8 passed, got %d", parsed.Passed)
	}
}

func TestWriteReportInvalidDir(t *testing.T) {
	r := &EvalReport{PromptID: "test", ConfigName: "cfg"}
	p := &prompt.Prompt{ID: "test", Service: "svc", Plane: "plane", Language: "lang", Category: "cat"}

	// Use a path containing characters that are invalid on both Unix and Windows.
	// On Windows, /nonexistent is treated as drive-relative and MkdirAll may succeed.
	invalidDir := filepath.Join(t.TempDir(), "not\x00valid")
	_, err := WriteReport(r, invalidDir, "run1", p)
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}
