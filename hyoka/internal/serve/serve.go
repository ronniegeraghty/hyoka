// Package serve starts a local HTTP server to browse evaluation reports.
package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Options configures the serve command.
type Options struct {
	ReportsDir string // directory containing evaluation report runs
	Port       int    // TCP port to listen on
	Open       bool   // auto-open the browser
}

// runInfo is one discovered run for the index page.
type runInfo struct {
	ID          string
	Timestamp   string
	HasSummary  bool
	SummaryLink string
	EvalCount   int
}

// summaryJSON is a minimal view of summary.json used to extract eval count.
type summaryJSON struct {
	TotalEvals int `json:"total_evaluations"`
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func Run(ctx context.Context, opts Options) error {
	absDir, err := filepath.Abs(opts.ReportsDir)
	if err != nil {
		return fmt.Errorf("resolving reports dir: %w", err)
	}

	if info, err := os.Stat(absDir); err != nil || !info.IsDir() {
		return fmt.Errorf("reports directory does not exist: %s", absDir)
	}

	mux := http.NewServeMux()

	// Serve the index page listing all runs.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		runs, err := discoverRuns(absDir)
		if err != nil {
			slog.Error("discovering runs", "error", err)
			http.Error(w, "failed to list runs", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := indexTmpl.Execute(w, runs); err != nil {
			slog.Error("rendering index", "error", err)
		}
	})

	// Serve static report files under /reports/...
	fileServer := http.FileServer(http.Dir(absDir))
	mux.Handle("/reports/", http.StripPrefix("/reports/", fileServer))

	addr := fmt.Sprintf(":%d", opts.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", addr, err)
	}

	url := fmt.Sprintf("http://localhost:%d", opts.Port)
	slog.Info("serving eval reports", "url", url, "dir", absDir)
	fmt.Printf("🔍 Hyoka report viewer: %s\n", url)
	fmt.Printf("   Reports directory:   %s\n", absDir)
	fmt.Println("   Press Ctrl+C to stop")

	if opts.Open {
		// Opened by the caller (main.go) after starting the server.
	}

	srv := &http.Server{Handler: mux}

	// Graceful shutdown on context cancel.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// discoverRuns scans the reports directory for run subdirectories.
func discoverRuns(dir string) ([]runInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var runs []runInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		// Skip non-run dirs (e.g. "trends").
		if name == "trends" || strings.HasPrefix(name, ".") {
			continue
		}

		ri := runInfo{
			ID:        name,
			Timestamp: formatRunID(name),
		}

		summaryPath := filepath.Join(dir, name, "summary.html")
		if _, err := os.Stat(summaryPath); err == nil {
			ri.HasSummary = true
			ri.SummaryLink = "/reports/" + name + "/summary.html"
		}

		// Try to read eval count from summary.json.
		sjPath := filepath.Join(dir, name, "summary.json")
		if data, err := os.ReadFile(sjPath); err == nil {
			var sj summaryJSON
			if json.Unmarshal(data, &sj) == nil {
				ri.EvalCount = sj.TotalEvals
			}
		}

		runs = append(runs, ri)
	}

	// Most recent first.
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].ID > runs[j].ID
	})

	return runs, nil
}

// formatRunID converts a timestamp-based run ID (e.g. "20260327-191240")
// into a human-readable string.
func formatRunID(id string) string {
	t, err := time.Parse("20060102-150405", id)
	if err != nil {
		return id
	}
	return t.Format("2006-01-02 15:04:05")
}

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Hyoka — Eval Report Viewer</title>
<style>
  :root {
    --bg: #f8fafc; --card-bg: #fff; --border: #e2e8f0;
    --text: #0f172a; --text-muted: #64748b; --blue: #2563eb;
    --green: #22c55e; --purple: #7c3aed;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: var(--bg); color: var(--text);
    max-width: 900px; margin: 0 auto; padding: 2rem;
  }
  h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
  .subtitle { color: var(--text-muted); margin-bottom: 1.5rem; }
  .run-list { list-style: none; }
  .run-item {
    background: var(--card-bg); border: 1px solid var(--border);
    border-radius: 8px; padding: 1rem 1.25rem; margin-bottom: 0.75rem;
    display: flex; justify-content: space-between; align-items: center;
    transition: box-shadow 0.15s;
  }
  .run-item:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
  .run-id { font-family: monospace; font-weight: 600; font-size: 0.95rem; }
  .run-ts { color: var(--text-muted); font-size: 0.85rem; margin-top: 0.15rem; }
  .run-meta { text-align: right; font-size: 0.85rem; color: var(--text-muted); }
  .run-meta a {
    display: inline-block; margin-left: 0.5rem; padding: 0.3rem 0.75rem;
    background: var(--blue); color: #fff; text-decoration: none;
    border-radius: 4px; font-size: 0.8rem; font-weight: 500;
  }
  .run-meta a:hover { opacity: 0.9; }
  .badge {
    display: inline-block; padding: 0.15rem 0.5rem;
    background: #e0e7ff; color: var(--blue); border-radius: 9999px;
    font-size: 0.75rem; font-weight: 600;
  }
  .empty { color: var(--text-muted); text-align: center; padding: 3rem; }
</style>
</head>
<body>
  <h1>📊 Hyoka — Eval Report Viewer</h1>
  <p class="subtitle">Browse evaluation runs</p>
  {{if .}}
  <ul class="run-list">
    {{range .}}
    <li class="run-item">
      <div>
        <div class="run-id">{{.ID}}</div>
        <div class="run-ts">{{.Timestamp}}</div>
      </div>
      <div class="run-meta">
        {{if .EvalCount}}<span class="badge">{{.EvalCount}} evals</span>{{end}}
        {{if .HasSummary}}<a href="{{.SummaryLink}}">View Summary</a>{{end}}
        <a href="/reports/{{.ID}}/">Browse Files</a>
      </div>
    </li>
    {{end}}
  </ul>
  {{else}}
  <div class="empty">No evaluation runs found.</div>
  {{end}}
</body>
</html>
`))
