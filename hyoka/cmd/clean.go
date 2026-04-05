package cmd

import (
"fmt"
"strings"

"github.com/ronniegeraghty/hyoka/internal/clean"
"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
var dryRun bool
var keepLogs int
var all bool
var yes bool

cmd := &cobra.Command{
Use:   "clean",
Short: "Remove stale Copilot CLI session state and orphaned processes from past eval runs",
Long: "Scans for orphaned hyoka-spawned Copilot processes (tagged with HYOKA_SESSION)\nand stale ~/.copilot/session-state/ directories. Lists any found processes and\nasks for confirmation before killing them. Session state accumulates over many\neval runs and can grow to gigabytes, so periodic cleanup is recommended.\n\nBy default only cleans hyoka-created sessions and processes. Use --all to also\nclean Copilot CLI log files and non-hyoka session state.",
RunE: func(cmd *cobra.Command, args []string) error {
out := cmd.OutOrStdout()

procs, scanErr := clean.ScanHyokaProcesses(out)
if scanErr != nil {
fmt.Fprintf(out, "Warning: process scan: %v\n", scanErr)
}

killOrphans := false
if len(procs) > 0 && !dryRun {
if yes {
killOrphans = true
} else {
fmt.Fprintf(out, "\nKill these %d process(es)? [y/N] ", len(procs))
var answer string
fmt.Fscanln(cmd.InOrStdin(), &answer)
answer = strings.TrimSpace(strings.ToLower(answer))
killOrphans = answer == "y" || answer == "yes"
}
if killOrphans {
killed := clean.KillHyokaProcesses(procs, out)
fmt.Fprintf(out, "Killed %d process(es)\n", killed)
} else {
fmt.Fprintln(out, "Skipped process cleanup.")
}
}

result, err := clean.Run(clean.Options{
DryRun:      dryRun,
KeepLogs:    keepLogs,
All:         all,
KillOrphans: false,
Out:         out,
})
if err != nil {
return err
}

if dryRun {
fmt.Fprintf(out, "\nDry run: found %d session(s) to remove", result.SessionsFound)
if len(procs) > 0 {
fmt.Fprintf(out, ", %d process(es) to kill", len(procs))
}
fmt.Fprintln(out)
} else {
parts := []string{fmt.Sprintf("%d session(s)", result.SessionsRemoved)}
if result.LogsRemoved > 0 {
parts = append(parts, fmt.Sprintf("%d log(s)", result.LogsRemoved))
}
fmt.Fprintf(out, "\nCleaned %s \u2014 freed %s\n",
strings.Join(parts, ", "), humanSize(result.BytesFreed))
}
return nil
},
}

cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleaned without deleting or killing")
cmd.Flags().IntVar(&keepLogs, "keep-logs", 50, "Number of recent log files to keep (only with --all)")
cmd.Flags().BoolVar(&all, "all", false, "Also clean Copilot CLI logs and non-hyoka session state")
cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt and kill orphaned processes automatically")

return cmd
}
