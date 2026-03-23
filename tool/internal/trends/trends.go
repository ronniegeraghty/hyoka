// Package trends provides analysis of evaluation trends over time.
// Package trends provides analysis of evaluation trends over time.
package trends

import (
"encoding/json"
"fmt"
"math"
"os"
"path/filepath"
"sort"
"strings"
"time"

"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
)

// TrendClassification indicates the overall trend direction.
type TrendClassification string

const (
TrendStable     TrendClassification = "stable"
TrendImproving  TrendClassification = "improving"
TrendRegressing TrendClassification = "regressing"
TrendFlaky      TrendClassification = "flaky"
TrendNew        TrendClassification = "new"
)

// TrendEntry holds data from a single historical evaluation run.
type TrendEntry struct {
RunID      string   `json:"run_id"`
Timestamp  string   `json:"timestamp"`
ConfigName string   `json:"config_name"`
PromptID   string   `json:"prompt_id"`
Success    bool     `json:"success"`
Duration   float64  `json:"duration_seconds"`
Score      int      `json:"score"`
HasReview  bool     `json:"has_review"`
ToolCalls  []string `json:"tool_calls"`
FileCount  int      `json:"file_count"`
Error      string   `json:"error,omitempty"`
}

// RunResult holds a single run's metrics for a specific prompt+config.
type RunResult struct {
RunID     string   `json:"run_id"`
Timestamp string   `json:"timestamp"`
Success   bool     `json:"success"`
Duration  float64  `json:"duration_seconds"`
FileCount int      `json:"file_count"`
Score     int      `json:"score"`
HasReview bool     `json:"has_review"`
ToolCalls []string `json:"tool_calls,omitempty"`
Error     string   `json:"error,omitempty"`
}

// PromptTrend holds time-series performance data for a single prompt.
type PromptTrend struct {
PromptID string                         `json:"prompt_id"`
Configs  map[string][]RunResult         `json:"configs"`
Trend    map[string]TrendClassification `json:"trend"`
}

// TrendReport summarizes historical trends for a set of prompts.
type TrendReport struct {
PromptID     string        `json:"prompt_id,omitempty"`
Service      string        `json:"service,omitempty"`
Language     string        `json:"language,omitempty"`
TotalRuns    int           `json:"total_runs"`
Entries      []TrendEntry  `json:"entries"`
PromptTrends []PromptTrend `json:"prompt_trends"`
RunIDs       []string      `json:"run_ids"`
Analysis     string        `json:"analysis,omitempty"`
GeneratedAt  string        `json:"generated_at"`
}

// TrendOptions configures trend report generation.
type TrendOptions struct {
ReportsDir string
PromptID   string
Service    string
Language   string
OutputDir  string
Analyze    bool
}

// Generate scans historical reports and produces a trend report.
func Generate(opts TrendOptions) (*TrendReport, error) {
entries, err := scanReports(opts.ReportsDir, opts.PromptID, opts.Service, opts.Language)
if err != nil {
return nil, fmt.Errorf("scanning reports: %w", err)
}

sort.Slice(entries, func(i, j int) bool {
return entries[i].Timestamp < entries[j].Timestamp
})

promptTrends, runIDs := buildPromptTrends(entries)

tr := &TrendReport{
PromptID:     opts.PromptID,
Service:      opts.Service,
Language:      opts.Language,
TotalRuns:    len(entries),
Entries:      entries,
PromptTrends: promptTrends,
RunIDs:       runIDs,
GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
}

return tr, nil
}

// buildPromptTrends groups entries by prompt → config → chronological runs.
func buildPromptTrends(entries []TrendEntry) ([]PromptTrend, []string) {
// Collect unique run IDs in order
runIDSet := map[string]bool{}
var runIDs []string
for _, e := range entries {
if !runIDSet[e.RunID] {
runIDSet[e.RunID] = true
runIDs = append(runIDs, e.RunID)
}
}

// Group by prompt → config → runs
type key struct{ prompt, config string }
grouped := map[key][]RunResult{}
promptSet := map[string]bool{}
var promptOrder []string

for _, e := range entries {
if !promptSet[e.PromptID] {
promptSet[e.PromptID] = true
promptOrder = append(promptOrder, e.PromptID)
}
k := key{e.PromptID, e.ConfigName}
grouped[k] = append(grouped[k], RunResult{
RunID:     e.RunID,
Timestamp: e.Timestamp,
Success:   e.Success,
Duration:  e.Duration,
FileCount: e.FileCount,
Score:     e.Score,
HasReview: e.HasReview,
ToolCalls: e.ToolCalls,
Error:     e.Error,
})
}

var trends []PromptTrend
for _, pid := range promptOrder {
pt := PromptTrend{
PromptID: pid,
Configs:  map[string][]RunResult{},
Trend:    map[string]TrendClassification{},
}
for k, runs := range grouped {
if k.prompt == pid {
pt.Configs[k.config] = runs
pt.Trend[k.config] = classifyTrend(runs)
}
}
trends = append(trends, pt)
}

return trends, runIDs
}

// classifyTrend determines the trend classification for a series of runs.
func classifyTrend(runs []RunResult) TrendClassification {
if len(runs) <= 1 {
return TrendNew
}

passes, fails := 0, 0
for _, r := range runs {
if r.Success {
passes++
} else {
fails++
}
}

if fails == 0 {
// All pass — check duration trend
if len(runs) >= 3 {
mid := len(runs) / 2
firstAvg := avgDuration(runs[:mid])
secondAvg := avgDuration(runs[mid:])
if firstAvg > 0 && secondAvg < firstAvg*0.8 {
return TrendImproving
}
}
return TrendStable
}

if passes == 0 {
return TrendRegressing
}

// Previously passing, now failing → regression
if runs[0].Success && !runs[len(runs)-1].Success {
return TrendRegressing
}

return TrendFlaky
}

func avgDuration(runs []RunResult) float64 {
if len(runs) == 0 {
return 0
}
total := 0.0
for _, r := range runs {
total += r.Duration
}
return total / float64(len(runs))
}

// formatDuration returns a human-readable duration string.
func formatDuration(seconds float64) string {
if seconds <= 0 {
return "—"
}
if seconds < 60 {
return fmt.Sprintf("%.0fs", seconds)
}
if seconds < 3600 {
return fmt.Sprintf("%.1fm", seconds/60)
}
return fmt.Sprintf("%.1fh", seconds/3600)
}

// trendEmoji returns an emoji for the classification.
func trendEmoji(t TrendClassification) string {
switch t {
case TrendStable:
return "✅ stable"
case TrendImproving:
return "📈 improving"
case TrendRegressing:
return "📉 regressing"
case TrendFlaky:
return "⚠️ flaky"
case TrendNew:
return "🆕 new"
default:
return string(t)
}
}

// WriteMarkdown writes the trend report as a Markdown file.
func WriteMarkdown(tr *TrendReport, outputDir string) (string, error) {
if err := os.MkdirAll(outputDir, 0755); err != nil {
return "", fmt.Errorf("creating trends directory: %w", err)
}

filename := trendFilename(tr) + "-trends.md"
outPath := filepath.Join(outputDir, filename)

var b strings.Builder
writeMarkdownReport(&b, tr)

if err := os.WriteFile(outPath, []byte(b.String()), 0644); err != nil {
return "", fmt.Errorf("writing trend markdown: %w", err)
}
return outPath, nil
}

// WriteHTML writes the trend report as an HTML file.
func WriteHTML(tr *TrendReport, outputDir string) (string, error) {
if err := os.MkdirAll(outputDir, 0755); err != nil {
return "", fmt.Errorf("creating trends directory: %w", err)
}

filename := trendFilename(tr) + "-trends.html"
outPath := filepath.Join(outputDir, filename)

var b strings.Builder
writeHTMLReport(&b, tr)

if err := os.WriteFile(outPath, []byte(b.String()), 0644); err != nil {
return "", fmt.Errorf("writing trend HTML: %w", err)
}
return outPath, nil
}

func trendFilename(tr *TrendReport) string {
if tr.PromptID != "" {
return tr.PromptID
}
parts := []string{}
if tr.Service != "" {
parts = append(parts, tr.Service)
}
if tr.Language != "" {
parts = append(parts, tr.Language)
}
if len(parts) == 0 {
return "all"
}
return strings.Join(parts, "-")
}

// scanReports walks the reports directory and extracts trend entries.
func scanReports(reportsDir, promptID, service, language string) ([]TrendEntry, error) {
var entries []TrendEntry

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

if promptID != "" && r.PromptID != promptID {
return nil
}
if service != "" {
svc, _ := r.PromptMeta["service"].(string)
if svc != service {
return nil
}
}
if language != "" {
lang, _ := r.PromptMeta["language"].(string)
if lang != language {
return nil
}
}

rel, _ := filepath.Rel(reportsDir, path)
parts := strings.Split(rel, string(os.PathSeparator))
runID := ""
if len(parts) > 0 {
runID = parts[0]
}

entry := TrendEntry{
RunID:      runID,
Timestamp:  r.Timestamp,
ConfigName: r.ConfigName,
PromptID:   r.PromptID,
Success:    r.Success,
Duration:   r.Duration,
ToolCalls:  r.ToolCalls,
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

// --- Markdown report ---

func writeMarkdownReport(b *strings.Builder, tr *TrendReport) {
title := "Performance Trends"
if tr.PromptID != "" {
title = fmt.Sprintf("Trends: %s", tr.PromptID)
} else if tr.Service != "" || tr.Language != "" {
parts := []string{}
if tr.Service != "" {
parts = append(parts, tr.Service)
}
if tr.Language != "" {
parts = append(parts, tr.Language)
}
title = fmt.Sprintf("Trends: %s", strings.Join(parts, " / "))
}

fmt.Fprintf(b, "# %s\n\n", title)
fmt.Fprintf(b, "**Generated:** %s | **Total Evaluations:** %d\n\n", tr.GeneratedAt, tr.TotalRuns)

if tr.TotalRuns == 0 {
b.WriteString("No historical data found matching the given filters.\n")
return
}

// AI Analysis section
if tr.Analysis != "" {
b.WriteString("## 🤖 AI Analysis\n\n")
b.WriteString(tr.Analysis)
b.WriteString("\n\n---\n\n")
}

// Summary statistics
b.WriteString("## Summary\n\n")
passed, failed, totalScore, scored := 0, 0, 0, 0
configCounts := map[string]int{}
for _, e := range tr.Entries {
if e.Success {
passed++
} else {
failed++
}
if e.HasReview {
totalScore += e.Score
scored++
}
configCounts[e.ConfigName]++
}

b.WriteString("| Metric | Value |\n")
b.WriteString("|--------|-------|\n")
fmt.Fprintf(b, "| Total Evaluations | %d |\n", tr.TotalRuns)
fmt.Fprintf(b, "| Passed | %d (%.0f%%) |\n", passed, pct(passed, tr.TotalRuns))
fmt.Fprintf(b, "| Failed | %d |\n", failed)
if scored > 0 {
fmt.Fprintf(b, "| Avg Score | %.1f/10 |\n", float64(totalScore)/float64(scored))
}
fmt.Fprintf(b, "| Unique Prompts | %d |\n", len(tr.PromptTrends))
fmt.Fprintf(b, "| Configs | %d |\n", len(configCounts))
b.WriteString("\n")

// Regression alerts
var regressions []string
for _, pt := range tr.PromptTrends {
for cfg, trend := range pt.Trend {
if trend == TrendRegressing {
regressions = append(regressions, fmt.Sprintf("- 📉 **%s** / `%s` — previously passing, now failing", pt.PromptID, cfg))
}
}
}
if len(regressions) > 0 {
b.WriteString("## ⚠️ Regression Alerts\n\n")
for _, r := range regressions {
b.WriteString(r + "\n")
}
b.WriteString("\n")
}

// Performance Over Time table
if len(tr.PromptTrends) > 0 && len(tr.RunIDs) > 0 {
b.WriteString("## Performance Over Time\n\n")

// Header
b.WriteString("| Prompt | Config |")
displayIDs := tr.RunIDs
if len(displayIDs) > 8 {
displayIDs = displayIDs[len(displayIDs)-8:]
}
for _, rid := range displayIDs {
short := rid
if len(short) > 10 {
short = short[:10]
}
fmt.Fprintf(b, " %s |", short)
}
b.WriteString(" Trend |\n")

// Separator
b.WriteString("|--------|--------|")
for range displayIDs {
b.WriteString("--------|")
}
b.WriteString("-------|\n")

// Rows
for _, pt := range tr.PromptTrends {
configs := sortedConfigNames(pt.Configs)
for _, cfg := range configs {
runs := pt.Configs[cfg]
runMap := map[string]RunResult{}
for _, r := range runs {
runMap[r.RunID] = r
}

fmt.Fprintf(b, "| %s | %s |", pt.PromptID, cfg)
for _, rid := range displayIDs {
if r, ok := runMap[rid]; ok {
icon := "❌"
if r.Success {
icon = "✅"
}
fmt.Fprintf(b, " %s %s |", icon, formatDuration(r.Duration))
} else {
b.WriteString(" — |")
}
}
fmt.Fprintf(b, " %s |\n", trendEmoji(pt.Trend[cfg]))
}
}
b.WriteString("\n")
}

// Config comparison
if len(configCounts) > 1 {
b.WriteString("## Config Comparison\n\n")
b.WriteString("| Config | Runs | Pass Rate | Avg Duration | Avg Score |\n")
b.WriteString("|--------|------|-----------|--------------|----------|\n")
for cfg, count := range configCounts {
cp, cs, cn := 0, 0, 0
totalDur := 0.0
for _, e := range tr.Entries {
if e.ConfigName == cfg {
if e.Success {
cp++
}
totalDur += e.Duration
if e.HasReview {
cs += e.Score
cn++
}
}
}
avgScore := "—"
if cn > 0 {
avgScore = fmt.Sprintf("%.1f/10", float64(cs)/float64(cn))
}
fmt.Fprintf(b, "| %s | %d | %.0f%% | %s | %s |\n",
cfg, count, pct(cp, count), formatDuration(totalDur/float64(count)), avgScore)
}
b.WriteString("\n")
}
}

// --- HTML report ---

func writeHTMLReport(b *strings.Builder, tr *TrendReport) {
title := "Performance Trends"
if tr.PromptID != "" {
title = fmt.Sprintf("Trends: %s", tr.PromptID)
}

// Find max duration for bar scaling
maxDur := 0.0
for _, e := range tr.Entries {
if e.Duration > maxDur {
maxDur = e.Duration
}
}
if maxDur == 0 {
maxDur = 1
}

fmt.Fprintf(b, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<style>
  :root {
    --green: #22c55e; --red: #ef4444; --yellow: #eab308; --blue: #2563eb;
    --bg: #f8fafc; --text: #0f172a; --text-muted: #64748b; --border: #e2e8f0;
    --green-bg: #f0fdf4; --red-bg: #fef2f2; --yellow-bg: #fefce8; --blue-bg: #eff6ff;
  }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1400px; margin: 0 auto; padding: 2rem 1rem; color: var(--text); background: var(--bg); }
  h1 { margin: 0 0 0.25rem; } h2 { margin: 2rem 0 0.75rem; color: var(--text); }
  .subtitle { color: var(--text-muted); margin-bottom: 1.5rem; font-size: 0.9rem; }
  .stats { display: flex; gap: 1rem; flex-wrap: wrap; margin: 1.25rem 0; }
  .stat { background: #fff; border: 1px solid var(--border); border-radius: 8px; padding: 1rem 1.25rem; text-align: center; min-width: 120px; }
  .stat-value { font-size: 1.5rem; font-weight: 700; } .stat-label { font-size: 0.8rem; color: var(--text-muted); }
  .pass { color: var(--green); } .fail { color: var(--red); }
  .analysis { background: var(--blue-bg); border: 1px solid #bfdbfe; border-radius: 10px; padding: 1.5rem; margin: 1.5rem 0; }
  .analysis h2 { margin: 0 0 0.75rem; font-size: 1.1rem; }
  .analysis-content { white-space: pre-wrap; font-size: 0.9rem; line-height: 1.6; }
  .alerts { background: var(--red-bg); border: 1px solid #fecaca; border-radius: 10px; padding: 1.25rem; margin: 1.5rem 0; }
  .alerts h2 { margin: 0 0 0.5rem; font-size: 1rem; color: var(--red); }
  .alerts ul { margin: 0; padding-left: 1.25rem; } .alerts li { margin: 0.25rem 0; font-size: 0.9rem; }
  table { width: 100%%; border-collapse: collapse; background: #fff; border: 1px solid var(--border); border-radius: 8px; overflow: hidden; margin-bottom: 1.5rem; table-layout: fixed; }
  th { background: #f1f5f9; padding: 0.5rem 0.6rem; text-align: center; font-size: 0.75rem; color: var(--text-muted); border-bottom: 2px solid var(--border); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  th:first-child { text-align: left; width: 200px; }
  th:nth-child(2) { text-align: left; width: 110px; }
  th:last-child { width: 100px; }
  td { padding: 0.4rem 0.6rem; border-bottom: 1px solid #f1f5f9; font-size: 0.82rem; text-align: center; vertical-align: middle; overflow: hidden; text-overflow: ellipsis; }
  td:first-child { text-align: left; font-weight: 600; font-size: 0.8rem; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  td:nth-child(2) { text-align: left; font-family: monospace; font-size: 0.78rem; }
  .cell { position: relative; min-width: 70px; white-space: nowrap; }
  .cell-pass { background: var(--green-bg); }
  .cell-fail { background: var(--red-bg); }
  .cell-icon { font-size: 1rem; }
  .cell-dur { font-size: 0.7rem; color: var(--text-muted); margin-top: 2px; }
  .cell-files { font-size: 0.65rem; color: var(--text-muted); }
  .dur-bar { height: 3px; background: var(--blue); border-radius: 2px; margin-top: 3px; transition: width 0.3s; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; white-space: nowrap; }
  .badge-stable { background: var(--green-bg); color: #15803d; }
  .badge-improving { background: #ecfdf5; color: #047857; }
  .badge-regressing { background: var(--red-bg); color: #b91c1c; }
  .badge-flaky { background: var(--yellow-bg); color: #a16207; }
  .badge-new { background: var(--blue-bg); color: var(--blue); }
  .config-table { margin-top: 1rem; table-layout: auto; }
  .config-table td, .config-table th { text-align: left; padding: 0.5rem 0.75rem; }
  .prompt-group td:first-child { border-top: 2px solid var(--border); }
  .empty-cell { color: #cbd5e1; }
</style>
</head>
<body>
`, title)

fmt.Fprintf(b, "<h1>📈 %s</h1>\n", title)
fmt.Fprintf(b, "<div class=\"subtitle\">Generated: %s · %d evaluations across %d runs</div>\n",
tr.GeneratedAt, tr.TotalRuns, len(tr.RunIDs))

if tr.TotalRuns == 0 {
b.WriteString("<p>No historical data found matching the given filters.</p>\n</body>\n</html>")
return
}

// Summary stats
passed, failed, totalScore, scored := 0, 0, 0, 0
configCounts := map[string]int{}
for _, e := range tr.Entries {
if e.Success {
passed++
} else {
failed++
}
if e.HasReview {
totalScore += e.Score
scored++
}
configCounts[e.ConfigName]++
}

b.WriteString("<div class=\"stats\">\n")
fmt.Fprintf(b, "  <div class=\"stat\"><div class=\"stat-value\">%d</div><div class=\"stat-label\">Total Evals</div></div>\n", tr.TotalRuns)
fmt.Fprintf(b, "  <div class=\"stat\"><div class=\"stat-value pass\">%d</div><div class=\"stat-label\">Passed (%.0f%%)</div></div>\n", passed, pct(passed, tr.TotalRuns))
fmt.Fprintf(b, "  <div class=\"stat\"><div class=\"stat-value fail\">%d</div><div class=\"stat-label\">Failed</div></div>\n", failed)
fmt.Fprintf(b, "  <div class=\"stat\"><div class=\"stat-value\">%d</div><div class=\"stat-label\">Prompts</div></div>\n", len(tr.PromptTrends))
fmt.Fprintf(b, "  <div class=\"stat\"><div class=\"stat-value\">%d</div><div class=\"stat-label\">Configs</div></div>\n", len(configCounts))
if scored > 0 {
fmt.Fprintf(b, "  <div class=\"stat\"><div class=\"stat-value\">%.1f</div><div class=\"stat-label\">Avg Score</div></div>\n", float64(totalScore)/float64(scored))
}
b.WriteString("</div>\n\n")

// AI Analysis
if tr.Analysis != "" {
b.WriteString("<div class=\"analysis\">\n")
b.WriteString("  <h2>🤖 AI Analysis</h2>\n")
fmt.Fprintf(b, "  <div class=\"analysis-content\">%s</div>\n", htmlEscape(tr.Analysis))
b.WriteString("</div>\n\n")
}

// Regression alerts
var regressions []string
for _, pt := range tr.PromptTrends {
for cfg, trend := range pt.Trend {
if trend == TrendRegressing {
regressions = append(regressions, fmt.Sprintf("<li><strong>%s</strong> / <code>%s</code> — previously passing, now failing</li>", pt.PromptID, cfg))
}
}
}
if len(regressions) > 0 {
b.WriteString("<div class=\"alerts\">\n")
b.WriteString("  <h2>⚠️ Regression Alerts</h2>\n  <ul>\n")
for _, r := range regressions {
fmt.Fprintf(b, "    %s\n", r)
}
b.WriteString("  </ul>\n</div>\n\n")
}

// Performance Over Time table
if len(tr.PromptTrends) > 0 && len(tr.RunIDs) > 0 {
b.WriteString("<h2>Performance Over Time</h2>\n")

displayIDs := tr.RunIDs
if len(displayIDs) > 10 {
displayIDs = displayIDs[len(displayIDs)-10:]
}

// Build displayID set for filtering
displayIDSet := make(map[string]bool, len(displayIDs))
for _, id := range displayIDs {
displayIDSet[id] = true
}

b.WriteString("<table>\n<thead><tr>\n")
b.WriteString("  <th>Prompt</th>\n  <th>Config</th>\n")
for i, rid := range displayIDs {
short := rid
if len(short) > 15 {
short = short[:15]
}
fmt.Fprintf(b, "  <th>Run %d<br><small>%s</small></th>\n", i+1, short)
}
b.WriteString("  <th>Trend</th>\n")
b.WriteString("</tr></thead>\n<tbody>\n")

for _, pt := range tr.PromptTrends {
// Skip prompts with no results in the displayed runs
hasDisplayedResult := false
for _, runs := range pt.Configs {
for _, r := range runs {
if displayIDSet[r.RunID] {
hasDisplayedResult = true
break
}
}
if hasDisplayedResult {
break
}
}
if !hasDisplayedResult {
continue
}

configs := sortedConfigNames(pt.Configs)
firstRow := true
for _, cfg := range configs {
runs := pt.Configs[cfg]
runMap := map[string]RunResult{}
for _, r := range runs {
runMap[r.RunID] = r
}

groupClass := ""
if firstRow {
groupClass = " class=\"prompt-group\""
}
fmt.Fprintf(b, "<tr%s>", groupClass)

if firstRow {
fmt.Fprintf(b, "<td rowspan=\"%d\">%s</td>", len(configs), pt.PromptID)
firstRow = false
}

fmt.Fprintf(b, "<td>%s</td>", cfg)

for _, rid := range displayIDs {
if r, ok := runMap[rid]; ok {
icon := "❌"
cellClass := "cell cell-fail"
if r.Success {
icon = "✅"
cellClass = "cell cell-pass"
}
barWidth := 0.0
if maxDur > 0 && r.Duration > 0 {
barWidth = math.Min(r.Duration/maxDur*100, 100)
}
fmt.Fprintf(b, "<td class=\"%s\"><div class=\"cell-icon\">%s</div><div class=\"cell-dur\">%s</div>",
cellClass, icon, formatDuration(r.Duration))
if r.FileCount > 0 {
fmt.Fprintf(b, "<div class=\"cell-files\">%d files</div>", r.FileCount)
}
fmt.Fprintf(b, "<div class=\"dur-bar\" style=\"width:%.0f%%\"></div></td>", barWidth)
} else {
b.WriteString("<td class=\"cell empty-cell\">—</td>")
}
}

trend := pt.Trend[cfg]
badgeClass := "badge badge-" + string(trend)
fmt.Fprintf(b, "<td><span class=\"%s\">%s</span></td>", badgeClass, trendEmoji(trend))
b.WriteString("</tr>\n")
}
}

b.WriteString("</tbody>\n</table>\n\n")
}

// Config Comparison
if len(configCounts) > 1 {
b.WriteString("<h2>Config Comparison</h2>\n")
b.WriteString("<table class=\"config-table\">\n<thead><tr><th>Config</th><th>Runs</th><th>Pass Rate</th><th>Avg Duration</th><th>Avg Score</th></tr></thead>\n<tbody>\n")
for cfg, count := range configCounts {
cp, cs, cn := 0, 0, 0
totalDur := 0.0
for _, e := range tr.Entries {
if e.ConfigName == cfg {
if e.Success {
cp++
}
totalDur += e.Duration
if e.HasReview {
cs += e.Score
cn++
}
}
}
avgScore := "—"
if cn > 0 {
avgScore = fmt.Sprintf("%.1f/10", float64(cs)/float64(cn))
}
passRate := pct(cp, count)
barColor := "var(--green)"
if passRate < 50 {
barColor = "var(--red)"
} else if passRate < 80 {
barColor = "var(--yellow)"
}
fmt.Fprintf(b, "<tr><td><strong>%s</strong></td><td>%d</td><td><span style=\"color:%s\">%.0f%%</span></td><td>%s</td><td>%s</td></tr>\n",
cfg, count, barColor, passRate, formatDuration(totalDur/float64(count)), avgScore)
}
b.WriteString("</tbody>\n</table>\n\n")
}

b.WriteString("</body>\n</html>")
}

func htmlEscape(s string) string {
s = strings.ReplaceAll(s, "&", "&amp;")
s = strings.ReplaceAll(s, "<", "&lt;")
s = strings.ReplaceAll(s, ">", "&gt;")
s = strings.ReplaceAll(s, "\"", "&quot;")
return s
}

func sortedConfigNames(configs map[string][]RunResult) []string {
names := make([]string, 0, len(configs))
for k := range configs {
names = append(names, k)
}
sort.Strings(names)
return names
}

func pct(n, total int) float64 {
if total == 0 {
return 0
}
return float64(n) / float64(total) * 100
}
