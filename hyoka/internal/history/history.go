package history

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/report"
)

// HistoryEntry holds data from one evaluation of a prompt under a specific config.
type HistoryEntry struct {
	RunID      string  `json:"run_id"`
	ConfigName string  `json:"config_name"`
	Success    bool    `json:"success"`
	Duration   float64 `json:"duration_seconds"`
	FileCount  int     `json:"file_count"`
	Score      int     `json:"score,omitempty"`
	HasReview  bool    `json:"has_review"`
	Error      string  `json:"error,omitempty"`
}

// ConfigSummary holds aggregate stats for a single config.
type ConfigSummary struct {
	Config   string  `json:"config"`
	Runs     int     `json:"runs"`
	Passed   int     `json:"passed"`
	PassRate float64 `json:"pass_rate"`
	AvgDur   float64 `json:"avg_duration_seconds"`
}

// HistoryReport holds the full history for a prompt.
type HistoryReport struct {
	PromptID string          `json:"prompt_id"`
	Entries  []HistoryEntry  `json:"entries"`
	Total    int             `json:"total_runs"`
	Passed   int             `json:"passed"`
	PassRate float64         `json:"pass_rate"`
	AvgDur   float64         `json:"avg_duration_seconds"`
	Configs  []ConfigSummary `json:"configs"`
}

// Options configures history report generation.
type Options struct {
	PromptID   string
	ReportsDir string
	JSON       bool
	HTML       bool
	OutputDir  string
}

// RunHistory scans all reports for a given prompt ID and displays its history.
func RunHistory(opts Options) error {
	if opts.PromptID == "" {
		return fmt.Errorf("--prompt-id is required")
	}

	entries, err := scanForPrompt(opts.ReportsDir, opts.PromptID)
	if err != nil {
		return fmt.Errorf("scanning reports: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "No results found for prompt %q in %s\n", opts.PromptID, opts.ReportsDir)
		return nil
	}

	hr := buildHistoryReport(opts.PromptID, entries)

	if opts.JSON {
		return writeJSON(hr)
	}
	if opts.HTML {
		return writeHTML(hr, opts.OutputDir)
	}
	printTable(hr)
	return nil
}

// scanForPrompt walks the reports directory and finds all results for a prompt.
func scanForPrompt(reportsDir, promptID string) ([]HistoryEntry, error) {
	var entries []HistoryEntry

	err := filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || info.Name() != "report.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var r report.EvalReport
		if err := json.Unmarshal(data, &r); err != nil {
			return nil
		}

		if r.PromptID != promptID {
			return nil
		}

		// Extract run ID from path: reports/{runID}/results/...
		rel, _ := filepath.Rel(reportsDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		runID := ""
		if len(parts) > 0 {
			runID = parts[0]
		}

		entry := HistoryEntry{
			RunID:      runID,
			ConfigName: r.ConfigName,
			Success:    r.Success,
			Duration:   r.Duration,
			FileCount:  len(r.GeneratedFiles),
			Error:      r.Error,
		}
		if r.Review != nil {
			entry.Score = r.Review.OverallScore
			entry.HasReview = true
		}

		entries = append(entries, entry)
		return nil
	})

	return entries, err
}

func buildHistoryReport(promptID string, entries []HistoryEntry) *HistoryReport {
	// Sort by run ID (chronological) then config
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].RunID == entries[j].RunID {
			return entries[i].ConfigName < entries[j].ConfigName
		}
		return entries[i].RunID < entries[j].RunID
	})

	passed := 0
	totalDur := 0.0
	configData := map[string]*ConfigSummary{}

	for _, e := range entries {
		if e.Success {
			passed++
		}
		totalDur += e.Duration

		cs, ok := configData[e.ConfigName]
		if !ok {
			cs = &ConfigSummary{Config: e.ConfigName}
			configData[e.ConfigName] = cs
		}
		cs.Runs++
		if e.Success {
			cs.Passed++
		}
		cs.AvgDur += e.Duration
	}

	var configs []ConfigSummary
	for _, cs := range configData {
		if cs.Runs > 0 {
			cs.PassRate = float64(cs.Passed) / float64(cs.Runs) * 100
			cs.AvgDur = cs.AvgDur / float64(cs.Runs)
		}
		configs = append(configs, *cs)
	}
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Config < configs[j].Config
	})

	avgDur := 0.0
	if len(entries) > 0 {
		avgDur = totalDur / float64(len(entries))
	}

	return &HistoryReport{
		PromptID: promptID,
		Entries:  entries,
		Total:    len(entries),
		Passed:   passed,
		PassRate: float64(passed) / float64(len(entries)) * 100,
		AvgDur:   avgDur,
		Configs:  configs,
	}
}

func printTable(hr *HistoryReport) {
	fmt.Printf("\nHistory for: %s\n", hr.PromptID)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Header
	fmt.Printf("%-20s %-14s %-8s %-10s %-6s %-6s\n",
		"Run", "Config", "Result", "Duration", "Files", "Score")
	fmt.Println(strings.Repeat("─", 70))

	for _, e := range hr.Entries {
		icon := "✅"
		if !e.Success {
			icon = "❌"
		}
		score := "—"
		if e.HasReview {
			score = fmt.Sprintf("%d/10", e.Score)
		}
		fmt.Printf("%-20s %-14s %-8s %-10s %-6d %-6s\n",
			truncateStr(e.RunID, 20),
			truncateStr(e.ConfigName, 14),
			icon,
			formatDuration(e.Duration),
			e.FileCount,
			score,
		)
	}

	fmt.Println()
	fmt.Printf("Summary: %d runs, %.0f%% pass rate, avg %s\n",
		hr.Total, hr.PassRate, formatDuration(hr.AvgDur))
	for _, cs := range hr.Configs {
		fmt.Printf("  %-14s %d runs, %.0f%% pass, avg %s\n",
			cs.Config+":", cs.Runs, cs.PassRate, formatDuration(cs.AvgDur))
	}
	fmt.Println()
}

func writeJSON(hr *HistoryReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(hr)
}

func writeHTML(hr *HistoryReport, outputDir string) error {
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	outPath := filepath.Join(outputDir, hr.PromptID+"-history.html")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating HTML file: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("history").Funcs(template.FuncMap{
		"fmtDuration": formatDuration,
		"statusIcon": func(success bool) string {
			if success {
				return "✅"
			}
			return "❌"
		},
		"scoreStr": func(e HistoryEntry) string {
			if e.HasReview {
				return fmt.Sprintf("%d/10", e.Score)
			}
			return "—"
		},
		"passRateColor": func(rate float64) string {
			if rate >= 90 {
				return "#22c55e"
			}
			if rate >= 70 {
				return "#eab308"
			}
			return "#ef4444"
		},
	}).Parse(historyHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err := tmpl.Execute(f, hr); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	fmt.Fprintf(os.Stderr, "HTML history written to %s\n", outPath)
	return nil
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "—"
	}
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	return fmt.Sprintf("%.1fm", seconds/60)
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

const historyHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>History: {{.PromptID}}</title>
<style>
  :root { --green: #22c55e; --red: #ef4444; --yellow: #eab308; --bg: #f8fafc; --text: #0f172a; --text-muted: #64748b; --border: #e2e8f0; --blue: #2563eb; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1100px; margin: 0 auto; padding: 2rem 1rem; color: var(--text); background: var(--bg); line-height: 1.6; }
  h1 { margin: 0 0 0.25rem; }
  .subtitle { color: var(--text-muted); margin-bottom: 1.5rem; }
  .stats { display: flex; gap: 1rem; flex-wrap: wrap; margin: 1.25rem 0; }
  .stat { background: #fff; border: 1px solid var(--border); border-radius: 8px; padding: 1rem 1.25rem; text-align: center; min-width: 120px; }
  .stat-value { font-size: 1.5rem; font-weight: 700; }
  .stat-label { font-size: 0.8rem; color: var(--text-muted); }
  table { width: 100%; border-collapse: collapse; background: #fff; border: 1px solid var(--border); border-radius: 8px; overflow: hidden; margin-bottom: 1.5rem; }
  th { background: #f1f5f9; padding: 0.6rem 0.75rem; text-align: left; font-size: 0.8rem; color: var(--text-muted); border-bottom: 2px solid var(--border); }
  td { padding: 0.5rem 0.75rem; border-bottom: 1px solid #f1f5f9; font-size: 0.85rem; }
  .pass { color: var(--green); }
  .fail { color: var(--red); }
  h2 { margin: 2rem 0 0.75rem; }
  .config-table td:first-child { font-weight: 600; }
</style>
</head>
<body>
<h1>📈 History: {{.PromptID}}</h1>
<div class="subtitle">{{.Total}} runs across {{len .Configs}} configs</div>

<div class="stats">
  <div class="stat"><div class="stat-value">{{.Total}}</div><div class="stat-label">Total Runs</div></div>
  <div class="stat"><div class="stat-value" style="color:{{passRateColor .PassRate}}">{{printf "%.0f%%" .PassRate}}</div><div class="stat-label">Pass Rate</div></div>
  <div class="stat"><div class="stat-value">{{fmtDuration .AvgDur}}</div><div class="stat-label">Avg Duration</div></div>
</div>

<h2>Run History</h2>
<table>
  <thead>
    <tr><th>Run</th><th>Config</th><th>Result</th><th>Duration</th><th>Files</th><th>Score</th></tr>
  </thead>
  <tbody>
    {{range .Entries}}
    <tr>
      <td><code>{{.RunID}}</code></td>
      <td>{{.ConfigName}}</td>
      <td>{{statusIcon .Success}}</td>
      <td>{{fmtDuration .Duration}}</td>
      <td>{{.FileCount}}</td>
      <td>{{scoreStr .}}</td>
    </tr>
    {{end}}
  </tbody>
</table>

{{if gt (len .Configs) 0}}
<h2>Config Breakdown</h2>
<table class="config-table">
  <thead>
    <tr><th>Config</th><th>Runs</th><th>Pass Rate</th><th>Avg Duration</th></tr>
  </thead>
  <tbody>
    {{range .Configs}}
    <tr>
      <td>{{.Config}}</td>
      <td>{{.Runs}}</td>
      <td style="color:{{passRateColor .PassRate}}">{{printf "%.0f%%" .PassRate}}</td>
      <td>{{fmtDuration .AvgDur}}</td>
    </tr>
    {{end}}
  </tbody>
</table>
{{end}}

</body>
</html>`
