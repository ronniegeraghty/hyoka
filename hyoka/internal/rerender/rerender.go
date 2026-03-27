package rerender

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ronniegeraghty/hyoka/internal/report"
)

// Options configures report re-rendering.
type Options struct {
	ReportsDir string
	RunID      string // Specific run ID, or empty for --all
	All        bool
}

// Run re-renders HTML/MD reports from existing report.json files.
func Run(opts Options) error {
	if opts.All {
		return rerenderAll(opts.ReportsDir)
	}
	if opts.RunID == "" {
		return fmt.Errorf("specify a run ID or use --all")
	}
	return rerenderRun(opts.ReportsDir, opts.RunID)
}

func rerenderRun(reportsDir, runID string) error {
	runDir := filepath.Join(reportsDir, runID)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return fmt.Errorf("run directory not found: %s", runDir)
	}

	// Find all report.json files under this run
	var reportFiles []string
	if err := filepath.Walk(runDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "report.json" && !info.IsDir() {
			reportFiles = append(reportFiles, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walking run directory %s: %w", runDir, err)
	}

	if len(reportFiles) == 0 {
		return fmt.Errorf("no report.json files found in %s", runDir)
	}

	// Re-render each individual report
	var allReports []*report.EvalReport
	for _, rf := range reportFiles {
		evalReport, err := loadReport(rf)
		if err != nil {
			fmt.Printf("  ⚠️  skipping %s: %v\n", rf, err)
			continue
		}
		allReports = append(allReports, evalReport)

		// Determine report directory components from the path
		service, plane, language, category := extractMeta(evalReport)

		// Write HTML
		if _, err := report.WriteHTMLReport(evalReport, reportsDir, runID, service, plane, language, category); err != nil {
			fmt.Printf("  ⚠️  HTML failed for %s: %v\n", evalReport.PromptID, err)
		}

		// Write Markdown
		if _, err := report.WriteMarkdownReport(evalReport, reportsDir, runID, service, plane, language, category); err != nil {
			fmt.Printf("  ⚠️  MD failed for %s: %v\n", evalReport.PromptID, err)
		}
	}

	// Re-render summary if we have reports
	if len(allReports) > 0 {
		summary := buildSummaryFromReports(runID, allReports)

		if _, err := report.WriteSummaryHTML(summary, reportsDir); err != nil {
			fmt.Printf("  ⚠️  summary HTML failed: %v\n", err)
		}
		if _, err := report.WriteSummaryMarkdown(summary, reportsDir); err != nil {
			fmt.Printf("  ⚠️  summary MD failed: %v\n", err)
		}
	}

	fmt.Printf("✅ Re-rendered %d report(s) for run %s\n", len(reportFiles), runID)
	return nil
}

func rerenderAll(reportsDir string) error {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return fmt.Errorf("reading reports directory: %w", err)
	}

	var runIDs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && e.Name() != "trends" {
			runIDs = append(runIDs, e.Name())
		}
	}
	sort.Strings(runIDs)

	if len(runIDs) == 0 {
		fmt.Println("No runs found in", reportsDir)
		return nil
	}

	fmt.Printf("Re-rendering %d run(s)...\n", len(runIDs))
	for _, id := range runIDs {
		fmt.Printf("  → %s\n", id)
		if err := rerenderRun(reportsDir, id); err != nil {
			fmt.Printf("  ⚠️  %s: %v\n", id, err)
		}
	}

	return nil
}

func loadReport(path string) (*report.EvalReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r report.EvalReport
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func extractMeta(r *report.EvalReport) (service, plane, language, category string) {
	service, _ = r.PromptMeta["service"].(string)
	plane, _ = r.PromptMeta["plane"].(string)
	language, _ = r.PromptMeta["language"].(string)
	category, _ = r.PromptMeta["category"].(string)
	return
}

func buildSummaryFromReports(runID string, reports []*report.EvalReport) *report.RunSummary {
	s := &report.RunSummary{
		RunID:      runID,
		TotalEvals: len(reports),
		Results:    reports,
	}

	promptSet := make(map[string]bool)
	configSet := make(map[string]bool)

	for _, r := range reports {
		promptSet[r.PromptID] = true
		configSet[r.ConfigName] = true
		s.Duration += r.Duration

		if r.Timestamp != "" && s.Timestamp == "" {
			s.Timestamp = r.Timestamp
		}

		if r.Success {
			s.Passed++
		} else if r.Error != "" {
			s.Errors++
		} else {
			s.Failed++
		}
	}

	s.TotalPrompts = len(promptSet)
	s.TotalConfigs = len(configSet)
	return s
}
