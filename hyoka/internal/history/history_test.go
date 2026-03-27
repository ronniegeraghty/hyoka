package history

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestReports(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	type testReport struct {
		runID    string
		service  string
		plane    string
		language string
		category string
		config   string
		report   map[string]any
	}

	reports := []testReport{
		{
			runID: "20260322-142536", service: "key-vault", plane: "data-plane",
			language: "python", category: "crud", config: "baseline",
			report: map[string]any{
				"prompt_id":       "key-vault-dp-python-crud",
				"config_name":     "baseline",
				"timestamp":       "2026-03-22T14:25:36Z",
				"duration_seconds": 50.0,
				"generated_files": []string{"main.py", "requirements.txt", "README.md"},
				"success":         true,
			},
		},
		{
			runID: "20260322-142536", service: "key-vault", plane: "data-plane",
			language: "python", category: "crud", config: "azure-mcp",
			report: map[string]any{
				"prompt_id":       "key-vault-dp-python-crud",
				"config_name":     "azure-mcp",
				"timestamp":       "2026-03-22T14:26:00Z",
				"duration_seconds": 84.0,
				"generated_files": []string{"main.py", "requirements.txt", "README.md"},
				"success":         true,
				"review":          map[string]any{"overall_score": 8},
			},
		},
		{
			runID: "20260322-151232", service: "key-vault", plane: "data-plane",
			language: "python", category: "crud", config: "baseline",
			report: map[string]any{
				"prompt_id":       "key-vault-dp-python-crud",
				"config_name":     "baseline",
				"timestamp":       "2026-03-22T15:12:32Z",
				"duration_seconds": 108.0,
				"generated_files": []string{"main.py", "requirements.txt", "README.md"},
				"success":         true,
			},
		},
		{
			runID: "20260322-151232", service: "key-vault", plane: "data-plane",
			language: "python", category: "crud", config: "azure-mcp",
			report: map[string]any{
				"prompt_id":       "key-vault-dp-python-crud",
				"config_name":     "azure-mcp",
				"timestamp":       "2026-03-22T15:13:00Z",
				"duration_seconds": 84.0,
				"generated_files": []string{"main.py", "requirements.txt"},
				"success":         false,
				"error":           "timeout",
			},
		},
		// A different prompt that should NOT appear
		{
			runID: "20260322-142536", service: "storage", plane: "data-plane",
			language: "python", category: "auth", config: "baseline",
			report: map[string]any{
				"prompt_id":       "storage-dp-python-auth",
				"config_name":     "baseline",
				"timestamp":       "2026-03-22T14:30:00Z",
				"duration_seconds": 30.0,
				"generated_files": []string{"auth.py"},
				"success":         true,
			},
		},
	}

	for _, r := range reports {
		reportDir := filepath.Join(dir, r.runID, "results", r.service, r.plane, r.language, r.category, r.config)
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			t.Fatal(err)
		}
		data, err := json.MarshalIndent(r.report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(reportDir, "report.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestScanForPrompt(t *testing.T) {
	dir := setupTestReports(t)

	entries, err := scanForPrompt(dir, "key-vault-dp-python-crud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Should not include the storage prompt
	for _, e := range entries {
		if e.ConfigName != "baseline" && e.ConfigName != "azure-mcp" {
			t.Errorf("unexpected config: %s", e.ConfigName)
		}
	}
}

func TestScanForPromptNotFound(t *testing.T) {
	dir := setupTestReports(t)

	entries, err := scanForPrompt(dir, "nonexistent-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestBuildHistoryReport(t *testing.T) {
	entries := []HistoryEntry{
		{RunID: "run1", ConfigName: "baseline", Success: true, Duration: 50},
		{RunID: "run1", ConfigName: "azure-mcp", Success: true, Duration: 84, HasReview: true, Score: 8},
		{RunID: "run2", ConfigName: "baseline", Success: true, Duration: 108},
		{RunID: "run2", ConfigName: "azure-mcp", Success: false, Duration: 84},
	}

	hr := buildHistoryReport("test-prompt", entries)

	if hr.Total != 4 {
		t.Errorf("expected 4 total, got %d", hr.Total)
	}
	if hr.Passed != 3 {
		t.Errorf("expected 3 passed, got %d", hr.Passed)
	}
	if hr.PassRate != 75 {
		t.Errorf("expected 75%% pass rate, got %.0f%%", hr.PassRate)
	}
	if len(hr.Configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(hr.Configs))
	}

	// Check entries are sorted
	if hr.Entries[0].RunID != "run1" || hr.Entries[0].ConfigName != "azure-mcp" {
		t.Errorf("expected first entry to be run1/azure-mcp, got %s/%s", hr.Entries[0].RunID, hr.Entries[0].ConfigName)
	}
}

func TestPrintTable(t *testing.T) {
	hr := &HistoryReport{
		PromptID: "test-prompt",
		Total:    2,
		Passed:   1,
		PassRate: 50,
		AvgDur:   67,
		Entries: []HistoryEntry{
			{RunID: "run1", ConfigName: "baseline", Success: true, Duration: 50, FileCount: 3},
			{RunID: "run1", ConfigName: "mcp", Success: false, Duration: 84, FileCount: 2, Error: "timeout"},
		},
		Configs: []ConfigSummary{
			{Config: "baseline", Runs: 1, Passed: 1, PassRate: 100, AvgDur: 50},
			{Config: "mcp", Runs: 1, Passed: 0, PassRate: 0, AvgDur: 84},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable(hr)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	checks := []string{
		"test-prompt",
		"baseline",
		"mcp",
		"50%",
		"Summary:",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q", check)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	hr := &HistoryReport{
		PromptID: "test-prompt",
		Total:    1,
		Passed:   1,
		PassRate: 100,
		AvgDur:   50,
		Entries: []HistoryEntry{
			{RunID: "run1", ConfigName: "baseline", Success: true, Duration: 50, FileCount: 3},
		},
		Configs: []ConfigSummary{
			{Config: "baseline", Runs: 1, Passed: 1, PassRate: 100, AvgDur: 50},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writeJSON(hr)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should be valid JSON
	var parsed HistoryReport
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed.PromptID != "test-prompt" {
		t.Errorf("expected prompt_id test-prompt, got %s", parsed.PromptID)
	}
}

func TestWriteHTML(t *testing.T) {
	dir := t.TempDir()

	hr := &HistoryReport{
		PromptID: "test-prompt",
		Total:    2,
		Passed:   2,
		PassRate: 100,
		AvgDur:   67,
		Entries: []HistoryEntry{
			{RunID: "run1", ConfigName: "baseline", Success: true, Duration: 50, FileCount: 3},
			{RunID: "run1", ConfigName: "mcp", Success: true, Duration: 84, FileCount: 2, HasReview: true, Score: 8},
		},
		Configs: []ConfigSummary{
			{Config: "baseline", Runs: 1, Passed: 1, PassRate: 100, AvgDur: 50},
			{Config: "mcp", Runs: 1, Passed: 1, PassRate: 100, AvgDur: 84},
		},
	}

	err := writeHTML(hr, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outPath := filepath.Join(dir, "test-prompt-history.html")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}

	content := string(data)
	checks := []string{
		"test-prompt",
		"Run History",
		"Config Breakdown",
		"baseline",
		"mcp",
		"100%",
		"8/10",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("HTML missing %q", check)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "—"},
		{-1, "—"},
		{30, "30s"},
		{59.9, "60s"},
		{90, "1.5m"},
		{3600, "60.0m"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRunHistoryEndToEnd(t *testing.T) {
	dir := setupTestReports(t)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := RunHistory(Options{
		PromptID:   "key-vault-dp-python-crud",
		ReportsDir: dir,
	})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "key-vault-dp-python-crud") {
		t.Error("expected prompt ID in output")
	}
	if !strings.Contains(output, "4 runs") {
		t.Error("expected '4 runs' in output")
	}
	if !strings.Contains(output, "75%") {
		t.Error("expected '75%' pass rate in output")
	}
}

func TestRunHistoryNoPromptID(t *testing.T) {
	err := RunHistory(Options{ReportsDir: "."})
	if err == nil {
		t.Error("expected error for missing prompt-id")
	}
}
