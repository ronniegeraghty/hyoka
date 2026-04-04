// Package report handles generation of JSON, HTML, and Markdown reports.
package report

import (
"encoding/json"
"fmt"
"log/slog"
"os"
"path/filepath"

"github.com/ronniegeraghty/hyoka/internal/prompt"
)

// ReportDir returns the directory path for a specific evaluation report.
// The path includes the prompt ID so that different prompts sharing the same
// service/plane/language/category get isolated workspace directories.
func ReportDir(outputDir string, runID string, p *prompt.Prompt) string {
	return filepath.Join(
		outputDir, runID, "results",
		p.Service(), p.Plane(), p.Language(), p.Category(), p.ID,
	)
}

// WriteReport writes an EvalReport as JSON to the appropriate directory.
func WriteReport(r *EvalReport, outputDir string, runID string, p *prompt.Prompt) (string, error) {
reportDir := filepath.Join(
ReportDir(outputDir, runID, p), r.ConfigName,
)

if err := os.MkdirAll(reportDir, 0755); err != nil {
return "", fmt.Errorf("creating report directory: %w", err)
}

reportPath := filepath.Join(reportDir, "report.json")

data, err := json.MarshalIndent(r, "", "  ")
if err != nil {
return "", fmt.Errorf("marshaling report: %w", err)
}

if err := os.WriteFile(reportPath, data, 0644); err != nil {
return "", fmt.Errorf("writing report: %w", err)
}

slog.Debug("Report written", "path", reportPath, "size", len(data))
return reportPath, nil
}

// WriteSummary writes a RunSummary as JSON.
func WriteSummary(s *RunSummary, outputDir string) (string, error) {
summaryDir := filepath.Join(outputDir, s.RunID)
if err := os.MkdirAll(summaryDir, 0755); err != nil {
return "", fmt.Errorf("creating summary directory: %w", err)
}

summaryPath := filepath.Join(summaryDir, "summary.json")

data, err := json.MarshalIndent(s, "", "  ")
if err != nil {
return "", fmt.Errorf("marshaling summary: %w", err)
}

if err := os.WriteFile(summaryPath, data, 0644); err != nil {
return "", fmt.Errorf("writing summary: %w", err)
}

slog.Debug("Summary written", "path", summaryPath, "size", len(data))
return summaryPath, nil
}
