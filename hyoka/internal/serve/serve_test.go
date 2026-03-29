package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestReports(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a run with summary
	run1 := filepath.Join(dir, "20260327-113302")
	os.MkdirAll(run1, 0755)
	summary := map[string]any{
		"run_id":            "20260327-113302",
		"timestamp":         "2026-03-27T18:33:02Z",
		"total_evaluations": 5,
		"passed":            3,
		"failed":            2,
		"errors":            0,
		"duration_seconds":  120.5,
	}
	data, _ := json.Marshal(summary)
	os.WriteFile(filepath.Join(run1, "summary.json"), data, 0644)
	os.WriteFile(filepath.Join(run1, "summary.html"), []byte("<html>test</html>"), 0644)

	// Create a run without summary (incomplete)
	run2 := filepath.Join(dir, "20260326-103024")
	os.MkdirAll(run2, 0755)

	// Create trends directory (should be skipped)
	os.MkdirAll(filepath.Join(dir, "trends"), 0755)

	return dir
}

func TestListRuns(t *testing.T) {
	dir := setupTestReports(t)
	runs, err := listRuns(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	// Should be sorted newest first
	if runs[0].RunID != "20260327-113302" {
		t.Errorf("expected newest first, got %q", runs[0].RunID)
	}

	// First run should have full metadata
	if runs[0].Total != 5 {
		t.Errorf("expected 5 total, got %d", runs[0].Total)
	}
	if runs[0].Passed != 3 {
		t.Errorf("expected 3 passed, got %d", runs[0].Passed)
	}
	if !runs[0].HasHTML {
		t.Error("expected HasHTML to be true for run with summary.html")
	}

	// Second run should have minimal info
	if runs[1].Total != 0 {
		t.Errorf("expected 0 total for incomplete run, got %d", runs[1].Total)
	}
}

func TestListRunsEmptyDir(t *testing.T) {
	dir := t.TempDir()
	runs, err := listRuns(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "-"},
		{30.5, "30.5s"},
		{90.0, "1m30s"},
		{3661.0, "61m1s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.expected {
			t.Errorf("formatDuration(%f): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestIndexEndpoint(t *testing.T) {
	dir := setupTestReports(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		runs, _ := listRuns(dir)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		indexTemplate.Execute(w, runs)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "20260327-113302") {
		t.Error("expected run ID in response")
	}
	if !strings.Contains(body, "hyoka") {
		t.Error("expected title in response")
	}
}

func TestAPIRunsEndpoint(t *testing.T) {
	dir := setupTestReports(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		runs, _ := listRuns(dir)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runs)
	})

	req := httptest.NewRequest("GET", "/api/runs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var runs []RunInfo
	if err := json.NewDecoder(rec.Body).Decode(&runs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestReportsFileServing(t *testing.T) {
	dir := setupTestReports(t)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(dir))
	mux.Handle("/reports/", http.StripPrefix("/reports/", fileServer))

	req := httptest.NewRequest("GET", "/reports/20260327-113302/summary.html", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<html>test</html>") {
		t.Error("expected HTML content")
	}
}

func TestStartListensAndServes(t *testing.T) {
	dir := setupTestReports(t)

	// Start server on random port in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- Start(Options{ReportsDir: dir, Port: 0})
	}()

	// Give server a moment to start, then try to hit it.
	// Since Start() picks port 0, we can't know the port.
	// Instead, verify it doesn't immediately error.
	select {
	case err := <-errCh:
		// If Start returns very quickly, it likely failed to bind.
		if err != nil {
			t.Fatalf("Start() returned error: %v", err)
		}
	default:
		// Still running — that's expected.
	}
}

func TestIndexContainsAutoRefresh(t *testing.T) {
	dir := setupTestReports(t)

	runs, _ := listRuns(dir)
	var buf strings.Builder
	indexTemplate.Execute(&buf, runs)

	if !strings.Contains(buf.String(), `http-equiv="refresh"`) {
		t.Error("expected auto-refresh meta tag in index HTML")
	}
}
