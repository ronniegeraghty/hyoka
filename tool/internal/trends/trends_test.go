package trends

import (
"encoding/json"
"os"
"path/filepath"
"strings"
"testing"

"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/review"
)

func TestScanReportsEmpty(t *testing.T) {
dir := t.TempDir()
entries, err := scanReports(dir, "", "", "")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(entries) != 0 {
t.Fatalf("expected 0 entries, got %d", len(entries))
}
}

func TestScanReportsFindsReports(t *testing.T) {
dir := t.TempDir()

runDir := filepath.Join(dir, "20250101-120000", "results", "key-vault", "data-plane", "python", "crud", "baseline")
if err := os.MkdirAll(runDir, 0755); err != nil {
t.Fatal(err)
}

r := report.EvalReport{
PromptID:   "key-vault-dp-python-crud",
ConfigName: "baseline",
Timestamp:  "2025-01-01T12:00:00Z",
Duration:   45.2,
Success:    true,
ToolCalls:  []string{"create_file", "bash"},
PromptMeta: map[string]any{
"service":  "key-vault",
"language": "python",
},
GeneratedFiles: []string{"main.py"},
Review: &review.ReviewResult{
OverallScore: 8,
},
}
data, _ := json.MarshalIndent(r, "", "  ")
if err := os.WriteFile(filepath.Join(runDir, "report.json"), data, 0644); err != nil {
t.Fatal(err)
}

entries, err := scanReports(dir, "", "", "")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(entries) != 1 {
t.Fatalf("expected 1 entry, got %d", len(entries))
}
if entries[0].PromptID != "key-vault-dp-python-crud" {
t.Errorf("unexpected prompt ID: %s", entries[0].PromptID)
}
if entries[0].Score != 8 {
t.Errorf("expected score 8, got %d", entries[0].Score)
}
if !entries[0].HasReview {
t.Error("expected HasReview to be true")
}
}

func TestScanReportsFilterByPromptID(t *testing.T) {
dir := t.TempDir()

for _, pid := range []string{"prompt-a", "prompt-b"} {
runDir := filepath.Join(dir, "run1", "results", "svc", "dp", "py", "crud", "baseline")
os.MkdirAll(runDir, 0755)
r := report.EvalReport{PromptID: pid, ConfigName: "baseline", PromptMeta: map[string]any{}}
data, _ := json.Marshal(r)
os.WriteFile(filepath.Join(runDir, pid+"-report.json"), data, 0644)
}
for _, pid := range []string{"prompt-a", "prompt-b"} {
runDir := filepath.Join(dir, "run1", pid)
os.MkdirAll(runDir, 0755)
r := report.EvalReport{PromptID: pid, ConfigName: "baseline", PromptMeta: map[string]any{}}
data, _ := json.Marshal(r)
os.WriteFile(filepath.Join(runDir, "report.json"), data, 0644)
}

entries, _ := scanReports(dir, "prompt-a", "", "")
if len(entries) != 1 {
t.Fatalf("expected 1 entry for prompt-a, got %d", len(entries))
}
}

func TestGenerateAndWriteMarkdown(t *testing.T) {
dir := t.TempDir()

runDir := filepath.Join(dir, "reports", "run1", "results", "svc", "dp", "py", "crud", "baseline")
os.MkdirAll(runDir, 0755)
r := report.EvalReport{
PromptID:   "test-prompt",
ConfigName: "baseline",
Timestamp:  "2025-01-01T12:00:00Z",
Success:    true,
PromptMeta: map[string]any{},
ToolCalls:  []string{"bash"},
}
data, _ := json.Marshal(r)
os.WriteFile(filepath.Join(runDir, "report.json"), data, 0644)

tr, err := Generate(TrendOptions{
ReportsDir: filepath.Join(dir, "reports"),
PromptID:   "test-prompt",
})
if err != nil {
t.Fatalf("Generate failed: %v", err)
}
if tr.TotalRuns != 1 {
t.Fatalf("expected 1 run, got %d", tr.TotalRuns)
}

outDir := filepath.Join(dir, "trends")
mdPath, err := WriteMarkdown(tr, outDir)
if err != nil {
t.Fatalf("WriteMarkdown failed: %v", err)
}
content, _ := os.ReadFile(mdPath)
if !strings.Contains(string(content), "test-prompt") {
t.Error("markdown should contain prompt ID")
}

htmlPath, err := WriteHTML(tr, outDir)
if err != nil {
t.Fatalf("WriteHTML failed: %v", err)
}
htmlContent, _ := os.ReadFile(htmlPath)
if !strings.Contains(string(htmlContent), "test-prompt") {
t.Error("HTML should contain prompt ID")
}
}

func TestEvaluateToolUsage(t *testing.T) {
result := &report.ToolUsageResult{
ExpectedTools: []string{"azure-mcp", "bash"},
ActualTools:   []string{"bash", "create_file", "read_file"},
MatchedTools:  []string{"bash"},
MissingTools:  []string{"azure-mcp"},
ExtraTools:    []string{"create_file", "read_file"},
Match:         false,
}

if result.Match {
t.Error("expected Match to be false")
}
if len(result.MissingTools) != 1 || result.MissingTools[0] != "azure-mcp" {
t.Errorf("unexpected missing tools: %v", result.MissingTools)
}
}

func TestPct(t *testing.T) {
if pct(0, 0) != 0 {
t.Error("pct(0,0) should be 0")
}
if pct(1, 2) != 50 {
t.Error("pct(1,2) should be 50")
}
if pct(3, 3) != 100 {
t.Error("pct(3,3) should be 100")
}
}

func TestClassifyTrend(t *testing.T) {
tests := []struct {
name     string
runs     []RunResult
expected TrendClassification
}{
{"single run", []RunResult{{Success: true}}, TrendNew},
{"all pass", []RunResult{{Success: true}, {Success: true}, {Success: true}}, TrendStable},
{"all fail", []RunResult{{Success: false}, {Success: false}}, TrendRegressing},
{"pass then fail", []RunResult{{Success: true}, {Success: true}, {Success: false}}, TrendRegressing},
{"mixed", []RunResult{{Success: false}, {Success: true}, {Success: false}, {Success: true}}, TrendFlaky},
{"improving duration", []RunResult{
{Success: true, Duration: 100}, {Success: true, Duration: 95},
{Success: true, Duration: 50}, {Success: true, Duration: 45},
}, TrendImproving},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := classifyTrend(tt.runs)
if got != tt.expected {
t.Errorf("classifyTrend() = %s, want %s", got, tt.expected)
}
})
}
}

func TestBuildPromptTrends(t *testing.T) {
entries := []TrendEntry{
{RunID: "run1", PromptID: "prompt-a", ConfigName: "baseline", Success: true, Duration: 45, Timestamp: "2025-01-01T12:00:00Z"},
{RunID: "run1", PromptID: "prompt-a", ConfigName: "azure-mcp", Success: true, Duration: 60, Timestamp: "2025-01-01T12:00:00Z"},
{RunID: "run2", PromptID: "prompt-a", ConfigName: "baseline", Success: true, Duration: 40, Timestamp: "2025-01-02T12:00:00Z"},
{RunID: "run2", PromptID: "prompt-a", ConfigName: "azure-mcp", Success: false, Duration: 0, Timestamp: "2025-01-02T12:00:00Z"},
{RunID: "run1", PromptID: "prompt-b", ConfigName: "baseline", Success: true, Duration: 30, Timestamp: "2025-01-01T12:00:00Z"},
}

trends, runIDs := buildPromptTrends(entries)

if len(runIDs) != 2 {
t.Fatalf("expected 2 run IDs, got %d", len(runIDs))
}
if runIDs[0] != "run1" || runIDs[1] != "run2" {
t.Errorf("unexpected run IDs: %v", runIDs)
}

if len(trends) != 2 {
t.Fatalf("expected 2 prompt trends, got %d", len(trends))
}

// Find prompt-a
var promptA *PromptTrend
for i := range trends {
if trends[i].PromptID == "prompt-a" {
promptA = &trends[i]
break
}
}
if promptA == nil {
t.Fatal("prompt-a not found")
}
if len(promptA.Configs) != 2 {
t.Errorf("expected 2 configs for prompt-a, got %d", len(promptA.Configs))
}
if len(promptA.Configs["baseline"]) != 2 {
t.Errorf("expected 2 runs for baseline, got %d", len(promptA.Configs["baseline"]))
}
}

func TestFormatDuration(t *testing.T) {
tests := []struct {
input    float64
expected string
}{
{0, "—"},
{45, "45s"},
{90, "1.5m"},
{3600, "1.0h"},
}
for _, tt := range tests {
got := formatDuration(tt.input)
if got != tt.expected {
t.Errorf("formatDuration(%v) = %s, want %s", tt.input, got, tt.expected)
}
}
}

func TestTrendEmoji(t *testing.T) {
if !strings.Contains(trendEmoji(TrendStable), "stable") {
t.Error("stable trend should contain 'stable'")
}
if !strings.Contains(trendEmoji(TrendRegressing), "regressing") {
t.Error("regressing trend should contain 'regressing'")
}
}

func TestGenerateWithMultipleRunsHTML(t *testing.T) {
dir := t.TempDir()
reportsDir := filepath.Join(dir, "reports")

// Create two runs with two configs
for _, run := range []struct{ id, ts string }{
{"20250101-120000", "2025-01-01T12:00:00Z"},
{"20250102-120000", "2025-01-02T12:00:00Z"},
} {
for _, cfg := range []string{"baseline", "azure-mcp"} {
runDir := filepath.Join(reportsDir, run.id, "results", "kv", "dp", "py", "crud", cfg)
os.MkdirAll(runDir, 0755)
r := report.EvalReport{
PromptID:       "kv-dp-python-crud",
ConfigName:     cfg,
Timestamp:      run.ts,
Duration:       45.0,
Success:        cfg == "baseline", // azure-mcp fails
PromptMeta:     map[string]any{},
ToolCalls:      []string{"bash"},
GeneratedFiles: []string{"main.py"},
}
data, _ := json.Marshal(r)
os.WriteFile(filepath.Join(runDir, "report.json"), data, 0644)
}
}

tr, err := Generate(TrendOptions{ReportsDir: reportsDir})
if err != nil {
t.Fatalf("Generate failed: %v", err)
}
if tr.TotalRuns != 4 {
t.Fatalf("expected 4 runs, got %d", tr.TotalRuns)
}
if len(tr.PromptTrends) != 1 {
t.Fatalf("expected 1 prompt trend, got %d", len(tr.PromptTrends))
}
if len(tr.RunIDs) != 2 {
t.Fatalf("expected 2 run IDs, got %d", len(tr.RunIDs))
}

outDir := filepath.Join(dir, "trends")
htmlPath, err := WriteHTML(tr, outDir)
if err != nil {
t.Fatalf("WriteHTML failed: %v", err)
}
htmlContent, _ := os.ReadFile(htmlPath)
html := string(htmlContent)

// Should contain time-series elements
if !strings.Contains(html, "Performance Over Time") {
t.Error("HTML should contain Performance Over Time section")
}
if !strings.Contains(html, "kv-dp-python-crud") {
t.Error("HTML should contain prompt ID")
}
if !strings.Contains(html, "baseline") {
t.Error("HTML should contain config name")
}
if !strings.Contains(html, "Config Comparison") {
t.Error("HTML should contain Config Comparison section")
}
// Should have trend badges
if !strings.Contains(html, "badge") {
t.Error("HTML should contain trend badges")
}
}

func TestGenerateWithAnalysis(t *testing.T) {
dir := t.TempDir()
reportsDir := filepath.Join(dir, "reports")

runDir := filepath.Join(reportsDir, "run1", "results", "svc", "dp", "py", "crud", "baseline")
os.MkdirAll(runDir, 0755)
r := report.EvalReport{
PromptID:   "test-prompt",
ConfigName: "baseline",
Timestamp:  "2025-01-01T12:00:00Z",
Success:    true,
PromptMeta: map[string]any{},
}
data, _ := json.Marshal(r)
os.WriteFile(filepath.Join(runDir, "report.json"), data, 0644)

tr, _ := Generate(TrendOptions{ReportsDir: reportsDir})
tr.Analysis = "Test analysis: baseline config is performing well."

outDir := filepath.Join(dir, "trends")
htmlPath, _ := WriteHTML(tr, outDir)
htmlContent, _ := os.ReadFile(htmlPath)
html := string(htmlContent)

if !strings.Contains(html, "AI Analysis") {
t.Error("HTML should contain AI Analysis section")
}
if !strings.Contains(html, "Test analysis") {
t.Error("HTML should contain analysis text")
}

mdPath, _ := WriteMarkdown(tr, outDir)
mdContent, _ := os.ReadFile(mdPath)
md := string(mdContent)

if !strings.Contains(md, "AI Analysis") {
t.Error("markdown should contain AI Analysis section")
}
}
