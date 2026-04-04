package cmd

import (
"context"
"fmt"
"log/slog"
"path/filepath"

"github.com/ronniegeraghty/hyoka/internal/trends"
"github.com/spf13/cobra"
)

func trendsCmd() *cobra.Command {
var promptID, service, language, reportsDir, output string
var analyze bool
var openBrowser bool

cmd := &cobra.Command{
Use:   "trends",
Short: "Generate historical trend reports with time-series performance data",
Long:  "Scans all past runs in reports/ directory and generates a trend report with pass-rate timelines, duration trends, config comparisons, and regression detection. AI-powered insights are included by default; use --no-analyze to skip.",
RunE: func(cmd *cobra.Command, args []string) error {
reportsDir = resolvePathFlag(cmd, "reports-dir", []string{"../reports", "./reports"})
if !cmd.Flags().Changed("output") {
output = filepath.Join(reportsDir, "trends")
}

tr, err := trends.Generate(trends.TrendOptions{
ReportsDir: reportsDir,
PromptID:   promptID,
Service:    service,
Language:   language,
OutputDir:  output,
Analyze:    analyze,
})
if err != nil {
return fmt.Errorf("generating trends: %w", err)
}

if tr.TotalRuns == 0 {
fmt.Println("No historical data found matching the given filters.")
return nil
}

// Run Copilot-powered analysis if requested
if analyze {
fmt.Println("\U0001f916 Running AI-powered trend analysis...")
analysis, err := trends.AnalyzeTrends(context.Background(), tr)
if err != nil {
slog.Warn("AI trend analysis failed", "error", err)
fmt.Printf("\u26a0\ufe0f  AI analysis failed: %v (continuing without analysis)\n", err)
} else {
tr.Analysis = analysis
fmt.Println("\n--- AI Analysis ---")
fmt.Println(analysis)
fmt.Println("-------------------")
}
}

mdPath, err := trends.WriteMarkdown(tr, output)
if err != nil {
return fmt.Errorf("writing markdown trends: %w", err)
}
fmt.Printf("Markdown trend report: %s\n", mdPath)

htmlPath, err := trends.WriteHTML(tr, output)
if err != nil {
return fmt.Errorf("writing HTML trends: %w", err)
}
fmt.Printf("HTML trend report:     %s\n", htmlPath)
fmt.Printf("\nAnalyzed %d historical evaluation(s) across %d prompt(s)\n", tr.TotalRuns, len(tr.PromptTrends))

if openBrowser && htmlPath != "" {
openInBrowser(htmlPath)
}

return nil
},
}

cmd.Flags().StringVar(&promptID, "prompt-id", "", "Filter trends by prompt ID")
cmd.Flags().StringVar(&service, "service", "", "Filter trends by Azure service")
cmd.Flags().StringVar(&language, "language", "", "Filter trends by programming language")
cmd.Flags().StringVar(&reportsDir, "reports-dir", "./reports", "Directory containing past evaluation reports")
cmd.Flags().StringVar(&output, "output", "./reports/trends", "Output directory for trend reports")
cmd.Flags().BoolVar(&analyze, "analyze", true, "Run AI-powered analysis of trends (enabled by default)")
cmd.Flags().BoolVar(&openBrowser, "open", false, "Auto-open the HTML trend report in the browser")

// --no-analyze opt-out: cobra doesn't auto-generate negation flags,
// so we register a separate bool and reconcile in RunE.
var noAnalyze bool
cmd.Flags().BoolVar(&noAnalyze, "no-analyze", false, "Skip AI-powered trend analysis")
// Wire no-analyze into analyze before RunE executes
origRunE := cmd.RunE
cmd.RunE = func(cmd *cobra.Command, args []string) error {
if noAnalyze {
analyze = false
}
return origRunE(cmd, args)
}

return cmd
}
