# history — Prompt History Command

Shows a single prompt's performance across all evaluation runs.

## Usage

```bash
hyoka history --prompt-id key-vault-dp-python-crud
hyoka history --prompt-id key-vault-dp-python-crud --json
hyoka history --prompt-id key-vault-dp-python-crud --html
```

## Wiring into CLI

Add to `cmd/hyoka/main.go` in `rootCmd()`:

```go
root.AddCommand(historyCmd())
```

Add the `historyCmd` function:

```go
func historyCmd() *cobra.Command {
	var promptID, reportsDir, outputDir string
	var jsonOutput, htmlOutput bool

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show a prompt's performance history across all runs",
		Long:  "Scans all report directories for a given prompt ID and shows pass/fail, duration, and score across every run and config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if reportsDir == "" {
				reportsDir = resolveReportsDir() // use existing path resolution
			}
			return history.RunHistory(history.Options{
				PromptID:   promptID,
				ReportsDir: reportsDir,
				JSON:       jsonOutput,
				HTML:       htmlOutput,
				OutputDir:  outputDir,
			})
		},
	}

	cmd.Flags().StringVar(&promptID, "prompt-id", "", "Prompt ID to show history for (required)")
	cmd.Flags().StringVar(&reportsDir, "reports-dir", "", "Reports directory (default: auto-detect)")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory for HTML (default: current dir)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&htmlOutput, "html", false, "Generate HTML history page")
	cmd.MarkFlagRequired("prompt-id")

	return cmd
}
```

Import: `"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/history"`

## Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--prompt-id` | ✅ | The prompt ID to look up |
| `--reports-dir` | | Override reports directory |
| `--json` | | Machine-readable JSON output |
| `--html` | | Generate a single-prompt history HTML page |
| `--output` | | Output directory for HTML file |
