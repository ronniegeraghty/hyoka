package cmd

import (
"fmt"

"github.com/ronniegeraghty/hyoka/internal/rerender"
"github.com/spf13/cobra"
)

func reportCmd() *cobra.Command {
var reportsDir string
var all bool

cmd := &cobra.Command{
Use:   "report [run-id]",
Short: "Re-render HTML/MD reports from existing report.json files",
Long:  "Re-generates report.html, report.md, summary.html, and summary.md using current templates without re-running evaluations. Useful after template improvements.",
Args:  cobra.MaximumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
reportsDir = resolvePathFlag(cmd, "reports-dir", []string{"../reports", "./reports"})

var runID string
if len(args) > 0 {
runID = args[0]
}

if !all && runID == "" {
return fmt.Errorf("specify a run ID or use --all to re-render all runs")
}

return rerender.Run(rerender.Options{
ReportsDir: reportsDir,
RunID:      runID,
All:        all,
})
},
}

cmd.Flags().StringVar(&reportsDir, "reports-dir", "./reports", "Directory containing evaluation reports")
cmd.Flags().BoolVar(&all, "all", false, "Re-render all runs")

return cmd
}
