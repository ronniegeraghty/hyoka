// Package serve provides a local web server for browsing evaluation reports.
package serve

import (
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
)

// RunInfo holds summary metadata about a single evaluation run.
type RunInfo struct {
	RunID     string `json:"run_id"`
	Timestamp string `json:"timestamp"`
	Total     int    `json:"total_evaluations"`
	Passed    int    `json:"passed"`
	Failed    int    `json:"failed"`
	Errors    int    `json:"errors"`
	Duration  float64 `json:"duration_seconds"`
	HasHTML   bool   `json:"-"`
}

// Options configures the serve command.
type Options struct {
	ReportsDir string
	Port       int
}

// Start launches a local HTTP server for browsing reports.
func Start(opts Options) error {
	if opts.Port == 0 {
		opts.Port = 8080
	}

	abs, err := filepath.Abs(opts.ReportsDir)
	if err != nil {
		return fmt.Errorf("resolving reports dir: %w", err)
	}
	opts.ReportsDir = abs

	mux := http.NewServeMux()

	// Index page listing all runs
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		runs, err := listRuns(opts.ReportsDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing runs: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := indexTemplate.Execute(w, runs); err != nil {
			slog.Error("Template execution failed", "error", err)
		}
	})

	// API endpoint for run list
	mux.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		runs, err := listRuns(opts.ReportsDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runs)
	})

	// Serve static report files (HTML, JSON, MD, etc.)
	fileServer := http.FileServer(http.Dir(opts.ReportsDir))
	mux.Handle("/reports/", http.StripPrefix("/reports/", fileServer))

	addr := fmt.Sprintf(":%d", opts.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)
	fmt.Printf("🌐 Serving evaluation reports at %s\n", url)
	fmt.Printf("   Reports directory: %s\n", opts.ReportsDir)
	fmt.Printf("   Press Ctrl+C to stop\n\n")

	return http.Serve(listener, mux)
}

func listRuns(reportsDir string) ([]RunInfo, error) {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("reading reports dir: %w", err)
	}

	var runs []RunInfo
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "trends" {
			continue
		}
		run := RunInfo{RunID: e.Name()}

		// Try to load summary.json for richer metadata
		summaryPath := filepath.Join(reportsDir, e.Name(), "summary.json")
		if data, err := os.ReadFile(summaryPath); err == nil {
			var summary struct {
				Timestamp string  `json:"timestamp"`
				Total     int     `json:"total_evaluations"`
				Passed    int     `json:"passed"`
				Failed    int     `json:"failed"`
				Errors    int     `json:"errors"`
				Duration  float64 `json:"duration_seconds"`
			}
			if err := json.Unmarshal(data, &summary); err == nil {
				run.Timestamp = summary.Timestamp
				run.Total = summary.Total
				run.Passed = summary.Passed
				run.Failed = summary.Failed
				run.Errors = summary.Errors
				run.Duration = summary.Duration
			}
		}

		// Check if HTML summary exists
		htmlPath := filepath.Join(reportsDir, e.Name(), "summary.html")
		if _, err := os.Stat(htmlPath); err == nil {
			run.HasHTML = true
		}

		runs = append(runs, run)
	}

	// Sort newest first
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].RunID > runs[j].RunID
	})

	return runs, nil
}

// formatDuration formats seconds into a human-readable string.
func formatDuration(seconds float64) string {
	if seconds == 0 {
		return "-"
	}
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%dm%ds", m, s)
}

var funcMap = template.FuncMap{
	"formatDuration": formatDuration,
	"passRate": func(passed, total int) string {
		if total == 0 {
			return "-"
		}
		return fmt.Sprintf("%.0f%%", float64(passed)/float64(total)*100)
	},
	"statusClass": func(passed, failed, total int) string {
		if total == 0 {
			return "pending"
		}
		if failed == 0 {
			return "success"
		}
		if passed == 0 {
			return "failure"
		}
		return "partial"
	},
	"trimPrefix": strings.TrimPrefix,
}

var indexTemplate = template.Must(template.New("index").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="30">
<title>hyoka — Evaluation Reports</title>
<style>
  :root { --bg: #0d1117; --surface: #161b22; --border: #30363d; --text: #e6edf3; --muted: #8b949e; --green: #3fb950; --red: #f85149; --yellow: #d29922; --blue: #58a6ff; }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif; background: var(--bg); color: var(--text); line-height: 1.5; padding: 2rem; }
  h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
  .subtitle { color: var(--muted); margin-bottom: 2rem; }
  table { width: 100%; border-collapse: collapse; background: var(--surface); border-radius: 6px; overflow: hidden; }
  th { text-align: left; padding: 0.75rem 1rem; color: var(--muted); font-size: 0.85rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; border-bottom: 1px solid var(--border); }
  td { padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); }
  tr:last-child td { border-bottom: none; }
  tr:hover td { background: rgba(88, 166, 255, 0.04); }
  a { color: var(--blue); text-decoration: none; }
  a:hover { text-decoration: underline; }
  .badge { display: inline-block; padding: 0.15rem 0.5rem; border-radius: 999px; font-size: 0.8rem; font-weight: 600; }
  .badge.success { background: rgba(63, 185, 80, 0.15); color: var(--green); }
  .badge.failure { background: rgba(248, 81, 73, 0.15); color: var(--red); }
  .badge.partial { background: rgba(210, 153, 34, 0.15); color: var(--yellow); }
  .badge.pending { background: rgba(139, 148, 158, 0.15); color: var(--muted); }
  .score { font-variant-numeric: tabular-nums; }
  .empty { text-align: center; padding: 3rem; color: var(--muted); }
</style>
</head>
<body>
<h1>hyoka — Evaluation Reports</h1>
<p class="subtitle">{{len .}} evaluation run{{if ne (len .) 1}}s{{end}}</p>
{{if .}}
<table>
  <thead>
    <tr>
      <th>Run ID</th>
      <th>Timestamp</th>
      <th>Evals</th>
      <th>Passed</th>
      <th>Failed</th>
      <th>Pass Rate</th>
      <th>Duration</th>
      <th>Report</th>
    </tr>
  </thead>
  <tbody>
    {{range .}}
    <tr>
      <td><code>{{.RunID}}</code></td>
      <td>{{if .Timestamp}}{{.Timestamp}}{{else}}<span style="color:var(--muted)">-</span>{{end}}</td>
      <td class="score">{{if .Total}}{{.Total}}{{else}}-{{end}}</td>
      <td class="score" style="color:var(--green)">{{if .Total}}{{.Passed}}{{else}}-{{end}}</td>
      <td class="score" style="color:var(--red)">{{if .Total}}{{.Failed}}{{else}}-{{end}}</td>
      <td><span class="badge {{statusClass .Passed .Failed .Total}}">{{passRate .Passed .Total}}</span></td>
      <td>{{formatDuration .Duration}}</td>
      <td>{{if .HasHTML}}<a href="/reports/{{.RunID}}/summary.html">View</a>{{else}}<span style="color:var(--muted)">—</span>{{end}}</td>
    </tr>
    {{end}}
  </tbody>
</table>
{{else}}
<div class="empty">No evaluation runs found.</div>
{{end}}
</body>
</html>
`))
