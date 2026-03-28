package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestReports creates a temporary reports directory with sample run data.
func setupTestReports(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Run 1: has summary.html and summary.json
	run1 := filepath.Join(dir, "20260327-191240")
	if err := os.MkdirAll(run1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run1, "summary.html"), []byte("<html>summary1</html>"), 0644); err != nil {
		t.Fatal(err)
	}
	sj := summaryJSON{TotalEvals: 5}
	data, _ := json.Marshal(sj)
	if err := os.WriteFile(filepath.Join(run1, "summary.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	// Run 2: no summary
	run2 := filepath.Join(dir, "20260324-103151")
	if err := os.MkdirAll(run2, 0755); err != nil {
		t.Fatal(err)
	}

	// Non-run dirs should be skipped
	if err := os.MkdirAll(filepath.Join(dir, "trends"), 0755); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestDiscoverRuns(t *testing.T) {
	dir := setupTestReports(t)

	runs, err := discoverRuns(dir)
	if err != nil {
		t.Fatalf("discoverRuns: %v", err)
	}

	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	// Most recent first
	if runs[0].ID != "20260327-191240" {
		t.Errorf("expected first run ID 20260327-191240, got %s", runs[0].ID)
	}
	if !runs[0].HasSummary {
		t.Error("expected first run to have summary")
	}
	if runs[0].EvalCount != 5 {
		t.Errorf("expected 5 evals, got %d", runs[0].EvalCount)
	}

	if runs[1].ID != "20260324-103151" {
		t.Errorf("expected second run ID 20260324-103151, got %s", runs[1].ID)
	}
	if runs[1].HasSummary {
		t.Error("expected second run to not have summary")
	}
}

func TestFormatRunID(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"20260327-191240", "2026-03-27 19:12:40"},
		{"invalid-id", "invalid-id"},
	}
	for _, tt := range tests {
		got := formatRunID(tt.input)
		if got != tt.want {
			t.Errorf("formatRunID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIndexPage(t *testing.T) {
	dir := setupTestReports(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		runs, _ := discoverRuns(dir)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = indexTmpl.Execute(w, runs)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "20260327-191240") {
		t.Error("index page should contain run ID")
	}
	if !strings.Contains(html, "View Summary") {
		t.Error("index page should contain summary link")
	}
	if !strings.Contains(html, "5 evals") {
		t.Error("index page should contain eval count")
	}
}

func TestStaticFileServing(t *testing.T) {
	dir := setupTestReports(t)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(dir))
	mux.Handle("/reports/", http.StripPrefix("/reports/", fileServer))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/reports/20260327-191240/summary.html")
	if err != nil {
		t.Fatalf("GET summary.html: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>summary1</html>" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestRunInvalidDir(t *testing.T) {
	err := Run(context.Background(), Options{
		ReportsDir: "/nonexistent/path",
		Port:       0,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent dir")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestServerLifecycle(t *testing.T) {
	dir := setupTestReports(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use port 0 to let OS pick a free port.
	// We need to start the server in a modified way to capture the actual port.
	opts := Options{
		ReportsDir: dir,
		Port:       0,
		Open:       false,
	}

	absDir, _ := filepath.Abs(opts.ReportsDir)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		runs, _ := discoverRuns(absDir)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = indexTmpl.Execute(w, runs)
	})

	fileServer := http.FileServer(http.Dir(absDir))
	mux.Handle("/reports/", http.StripPrefix("/reports/", fileServer))

	srv := &http.Server{Handler: mux}
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	port := ln.Addr().(*net.TCPAddr).Port

	// Give server a moment to start.
	time.Sleep(50 * time.Millisecond)

	base := fmt.Sprintf("http://localhost:%d", port)

	// Test the index
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Test static file
	resp2, err := http.Get(base + "/reports/20260327-191240/summary.html")
	if err != nil {
		t.Fatalf("GET summary: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	// Shutdown
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	srvErr := <-errCh
	if srvErr != nil && srvErr != http.ErrServerClosed {
		t.Fatalf("server error: %v", srvErr)
	}
}

func TestDiscoverRunsEmpty(t *testing.T) {
	dir := t.TempDir()
	runs, err := discoverRuns(dir)
	if err != nil {
		t.Fatalf("discoverRuns: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected 0 runs, got %d", len(runs))
	}
}
