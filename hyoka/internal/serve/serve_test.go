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

	// Create a nested report.json for eval endpoint
	evalDir := filepath.Join(run1, "results", "identity")
	os.MkdirAll(evalDir, 0755)
	os.WriteFile(filepath.Join(evalDir, "report.json"), []byte(`{"score":85}`), 0644)

	// Create a run without summary (incomplete)
	run2 := filepath.Join(dir, "20260326-103024")
	os.MkdirAll(run2, 0755)

	// Create trends directory (should be skipped)
	os.MkdirAll(filepath.Join(dir, "trends"), 0755)

	return dir
}

func setupTestDocs(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "getting-started.md"), []byte("# Getting Started\n\nWelcome to hyoka."), 0644)
	os.WriteFile(filepath.Join(dir, "architecture.md"), []byte("# Architecture\n\nOverview of the system."), 0644)
	os.WriteFile(filepath.Join(dir, "cleanup-plan.md"), []byte("# Cleanup Plan\n\nInternal doc."), 0644)
	os.WriteFile(filepath.Join(dir, "eval-tool-plan.md"), []byte("# Eval Tool Plan\n\nInternal doc."), 0644)

	return dir
}

func setupTestSite(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!DOCTYPE html><html><body>SPA</body></html>"), 0644)
	assetsDir := filepath.Join(dir, "assets")
	os.MkdirAll(assetsDir, 0755)
	os.WriteFile(filepath.Join(assetsDir, "main.js"), []byte("console.log('app')"), 0644)

	return dir
}

func TestListRunSummaries(t *testing.T) {
	dir := setupTestReports(t)
	runs, err := listRunSummaries(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	// First entry (newest) should have full summary data
	var first map[string]any
	if err := json.Unmarshal(runs[0], &first); err != nil {
		t.Fatalf("failed to unmarshal first run: %v", err)
	}
	if first["run_id"] != "20260327-113302" {
		t.Errorf("expected newest first, got %v", first["run_id"])
	}

	// Second entry (no summary.json) should have minimal data
	var second map[string]any
	if err := json.Unmarshal(runs[1], &second); err != nil {
		t.Fatalf("failed to unmarshal second run: %v", err)
	}
	if second["run_id"] != "20260326-103024" {
		t.Errorf("expected oldest second, got %v", second["run_id"])
	}
}

func TestListRunSummariesEmptyDir(t *testing.T) {
	dir := t.TempDir()
	runs, err := listRunSummaries(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestAPIRunsEndpoint(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	req := httptest.NewRequest("GET", "/api/runs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var runs []json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&runs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestAPIRunDetailEndpoint(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	req := httptest.NewRequest("GET", "/api/runs/20260327-113302", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var summary map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if summary["run_id"] != "20260327-113302" {
		t.Errorf("unexpected run_id: %v", summary["run_id"])
	}
}

func TestAPIRunDetailNotFound(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	req := httptest.NewRequest("GET", "/api/runs/nonexistent", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestAPIEvalEndpoint(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	req := httptest.NewRequest("GET", "/api/runs/20260327-113302/eval?path=results/identity/report.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"score":85`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestAPIEvalMissingPath(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	req := httptest.NewRequest("GET", "/api/runs/20260327-113302/eval", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAPIEvalTraversalBlocked(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	req := httptest.NewRequest("GET", "/api/runs/20260327-113302/eval?path=../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAPIRunIDTraversalBlocked(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

	tests := []struct {
		name string
		path string
	}{
		{"dotdot in runID", "/api/runs/../../etc/passwd"},
		{"dotdot runID with eval", "/api/runs/../../../etc/eval?path=report.json"},
		{"dotdot runID summary", "/api/runs/.."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				t.Errorf("expected non-200 for traversal path %q, got %d", tc.path, rec.Code)
			}
		})
	}
}

func TestAPIDocsEndpoint(t *testing.T) {
	docsDir := setupTestDocs(t)
	mux := buildMux(Options{ReportsDir: t.TempDir(), DocsDir: docsDir})

	req := httptest.NewRequest("GET", "/api/docs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var docs []DocInfo
	if err := json.NewDecoder(rec.Body).Decode(&docs); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	// Should exclude internal docs
	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d: %+v", len(docs), docs)
	}

	slugs := map[string]bool{}
	for _, d := range docs {
		slugs[d.Slug] = true
		if d.Title == "" {
			t.Errorf("expected title for slug %q", d.Slug)
		}
	}
	if slugs["cleanup-plan"] || slugs["eval-tool-plan"] {
		t.Error("internal docs should be filtered out")
	}
}

func TestAPIDocDetailEndpoint(t *testing.T) {
	docsDir := setupTestDocs(t)
	mux := buildMux(Options{ReportsDir: t.TempDir(), DocsDir: docsDir})

	req := httptest.NewRequest("GET", "/api/docs/getting-started", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var doc DocInfo
	if err := json.NewDecoder(rec.Body).Decode(&doc); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if doc.Slug != "getting-started" {
		t.Errorf("unexpected slug: %q", doc.Slug)
	}
	if doc.Title != "Getting Started" {
		t.Errorf("unexpected title: %q", doc.Title)
	}
	if !strings.Contains(doc.Content, "Welcome to hyoka") {
		t.Error("expected content in response")
	}
}

func TestAPIDocDetailInternalFiltered(t *testing.T) {
	docsDir := setupTestDocs(t)
	mux := buildMux(Options{ReportsDir: t.TempDir(), DocsDir: docsDir})

	req := httptest.NewRequest("GET", "/api/docs/cleanup-plan", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for internal doc, got %d", rec.Code)
	}
}

func TestReportsFileServing(t *testing.T) {
	dir := setupTestReports(t)
	mux := buildMux(Options{ReportsDir: dir})

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

func TestSPAFallback(t *testing.T) {
	reportsDir := setupTestReports(t)
	siteDir := setupTestSite(t)
	mux := buildMux(Options{ReportsDir: reportsDir, SiteDir: siteDir})

	// Static file should be served directly
	req := httptest.NewRequest("GET", "/assets/main.js", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for static file, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "console.log") {
		t.Error("expected JS content")
	}

	// Unknown path should fall back to index.html (SPA routing)
	req = httptest.NewRequest("GET", "/runs/20260327", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for SPA fallback, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "SPA") {
		t.Error("expected SPA index.html content")
	}
}

func TestSPANoSiteDir(t *testing.T) {
	mux := buildMux(Options{ReportsDir: t.TempDir()})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 with no site dir, got %d", rec.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	mux := buildMux(Options{ReportsDir: t.TempDir()})
	handler := corsMiddleware(mux)

	req := httptest.NewRequest("GET", "/api/runs", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header")
	}
}

func TestCORSPreflight(t *testing.T) {
	mux := buildMux(Options{ReportsDir: t.TempDir()})
	handler := corsMiddleware(mux)

	req := httptest.NewRequest("OPTIONS", "/api/runs", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rec.Code)
	}
}

func TestExtractMarkdownTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"# Hello World\n\nContent", "Hello World"},
		{"Some text\n# Title\nMore", "Title"},
		{"No heading here", ""},
		{"## Not H1\n# Actual Title", "Actual Title"},
	}
	for _, tt := range tests {
		got := extractMarkdownTitle(tt.input)
		if got != tt.expected {
			t.Errorf("extractMarkdownTitle(%q): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestStartListensAndServes(t *testing.T) {
	dir := setupTestReports(t)

	errCh := make(chan error, 1)
	go func() {
		errCh <- Start(Options{ReportsDir: dir, Port: 0})
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start() returned error: %v", err)
		}
	default:
		// Still running — that's expected.
	}
}

func setupTestPrompts(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	subDir := filepath.Join(dir, "identity", "python")
	os.MkdirAll(subDir, 0755)

	promptContent := `---
id: test-prompt-one
service: identity
plane: data-plane
language: python
category: auth
difficulty: basic
description: "Test prompt"
sdk_package: azure-identity
tags:
  - test
created: "2025-01-01"
author: test
---

## Prompt

This is a test prompt.
`
	os.WriteFile(filepath.Join(subDir, "test-prompt-one.prompt.md"), []byte(promptContent), 0644)

	promptContent2 := `---
id: test-prompt-two
service: storage
plane: management-plane
language: go
category: blobs
difficulty: intermediate
description: "Second test prompt"
sdk_package: azure-storage-blob
tags:
  - storage
  - blob
created: "2025-02-01"
author: tester
---

## Prompt

This is a second test prompt.
`
	os.WriteFile(filepath.Join(subDir, "test-prompt-two.prompt.md"), []byte(promptContent2), 0644)

	return dir
}

func TestAPIPromptsEndpoint(t *testing.T) {
	promptsDir := setupTestPrompts(t)
	mux := buildMux(Options{ReportsDir: t.TempDir(), PromptsDir: promptsDir})

	req := httptest.NewRequest("GET", "/api/prompts", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var prompts []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&prompts); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}

	ids := map[string]bool{}
	for _, p := range prompts {
		ids[p["id"].(string)] = true
	}
	if !ids["test-prompt-one"] || !ids["test-prompt-two"] {
		t.Errorf("expected both test prompts, got %v", ids)
	}
}

func TestAPIPromptDetailEndpoint(t *testing.T) {
	promptsDir := setupTestPrompts(t)
	mux := buildMux(Options{ReportsDir: t.TempDir(), PromptsDir: promptsDir})

	req := httptest.NewRequest("GET", "/api/prompts/test-prompt-one", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var p map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&p); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if p["id"] != "test-prompt-one" {
		t.Errorf("expected id test-prompt-one, got %v", p["id"])
	}
	if p["service"] != "identity" {
		t.Errorf("expected service identity, got %v", p["service"])
	}
}

func TestAPIPromptDetailNotFound(t *testing.T) {
	promptsDir := setupTestPrompts(t)
	mux := buildMux(Options{ReportsDir: t.TempDir(), PromptsDir: promptsDir})

	req := httptest.NewRequest("GET", "/api/prompts/nonexistent-prompt", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
