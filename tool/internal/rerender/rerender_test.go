package rerender

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
)

func TestRerenderRun(t *testing.T) {
	dir := t.TempDir()
	runID := "20240115-100000"

	// Create a report.json
	r := &report.EvalReport{
		PromptID:       "test-prompt",
		ConfigName:     "baseline",
		Timestamp:      "2024-01-15T10:00:00Z",
		Duration:       12.5,
		PromptMeta:     map[string]any{"service": "storage", "plane": "data-plane", "language": "go", "category": "auth"},
		ConfigUsed:     map[string]any{"name": "baseline", "model": "gpt-4"},
		GeneratedFiles: []string{"main.go"},
		EventCount:     5,
		ToolCalls:      []string{"create_file"},
		Success:        true,
	}

	reportDir := filepath.Join(dir, runID, "results", "storage", "data-plane", "go", "auth", "baseline")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(r, "", "  ")
	if err := os.WriteFile(filepath.Join(reportDir, "report.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	// Run rerender
	err := Run(Options{ReportsDir: dir, RunID: runID})
	if err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	// Check HTML report was generated
	htmlPath := filepath.Join(reportDir, "report.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("expected report.html to exist")
	}

	// Check MD report was generated
	mdPath := filepath.Join(reportDir, "report.md")
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Error("expected report.md to exist")
	}

	// Check summary HTML was generated
	summaryHTML := filepath.Join(dir, runID, "summary.html")
	if _, err := os.Stat(summaryHTML); os.IsNotExist(err) {
		t.Error("expected summary.html to exist")
	}
}

func TestRerenderAll(t *testing.T) {
	dir := t.TempDir()

	// Create two runs with report.json files
	for _, runID := range []string{"20240101-100000", "20240102-100000"} {
		reportDir := filepath.Join(dir, runID, "results", "storage", "data-plane", "python", "crud", "baseline")
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			t.Fatal(err)
		}
		r := &report.EvalReport{
			PromptID:   "test-prompt",
			ConfigName: "baseline",
			Timestamp:  "2024-01-01T10:00:00Z",
			PromptMeta: map[string]any{"service": "storage", "plane": "data-plane", "language": "python", "category": "crud"},
			ConfigUsed: map[string]any{"name": "baseline"},
			Success:    true,
		}
		data, _ := json.MarshalIndent(r, "", "  ")
		os.WriteFile(filepath.Join(reportDir, "report.json"), data, 0644)
	}

	err := Run(Options{ReportsDir: dir, All: true})
	if err != nil {
		t.Fatalf("rerender all failed: %v", err)
	}

	// Check both runs got re-rendered
	for _, runID := range []string{"20240101-100000", "20240102-100000"} {
		summaryHTML := filepath.Join(dir, runID, "summary.html")
		if _, err := os.Stat(summaryHTML); os.IsNotExist(err) {
			t.Errorf("expected summary.html for run %s", runID)
		}
	}
}

func TestRerenderNoRunFound(t *testing.T) {
	dir := t.TempDir()
	err := Run(Options{ReportsDir: dir, RunID: "nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

func TestBuildSummaryFromReports(t *testing.T) {
	reports := []*report.EvalReport{
		{PromptID: "p1", ConfigName: "c1", Success: true, Duration: 10},
		{PromptID: "p1", ConfigName: "c2", Success: false, Duration: 20},
		{PromptID: "p2", ConfigName: "c1", Success: true, Error: "", Duration: 15},
	}

	s := buildSummaryFromReports("run1", reports)
	if s.TotalEvals != 3 {
		t.Errorf("expected 3 evals, got %d", s.TotalEvals)
	}
	if s.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", s.Passed)
	}
	if s.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", s.Failed)
	}
	if s.TotalPrompts != 2 {
		t.Errorf("expected 2 prompts, got %d", s.TotalPrompts)
	}
	if s.TotalConfigs != 2 {
		t.Errorf("expected 2 configs, got %d", s.TotalConfigs)
	}
}
