# Contributing Guide

## Prerequisites

- Go 1.24.5+
- Node.js 18+ (for Copilot CLI)
- GitHub CLI (`gh`)
- Copilot CLI (`copilot`)

## Building

```bash
# From repo root (uses go.work)
cd /path/to/hyoka
go build ./hyoka/...
```

## Running Tests

```bash
# All tests
go test ./hyoka/...

# Specific package
go test ./hyoka/internal/eval/ -v

# With race detection
go test -race ./hyoka/...
```

## Project Structure

```
hyoka/              # Go module root
  main.go           # CLI entry (cobra commands)
  internal/
    build/          # Language-specific build verification
    checkenv/       # Environment prerequisite checks
    config/         # YAML config loading and parsing
    criteria/       # Tiered evaluation criteria system
    eval/           # Evaluation engine, process tracker, resource monitor
    history/        # Run history tracking
    logging/        # Structured logging (slog)
    manifest/       # Dependency manifest
    progress/       # Live progress display
    prompt/         # Prompt loading, parsing, filtering, validation
    report/         # Report generation (JSON, HTML, Markdown)
    rerender/       # Re-rendering past reports
    review/         # Multi-model review panel, rubric
    serve/          # Local web server for report browsing
    skills/         # Skill fetching (local + remote)
    trends/         # Cross-run trend analysis
    utils/          # Shared utilities
    validate/       # Prompt and config validation
configs/            # Evaluation configurations
criteria/           # Attribute-matched criteria (per-language, per-service)
prompts/            # Prompt library
skills/             # Copilot skills (generator + reviewer)
```

## Adding a New Command

1. Create the command function in `main.go`:
   ```go
   func myCmd() *cobra.Command {
       return &cobra.Command{
           Use:   "my-command",
           Short: "Description",
           RunE: func(cmd *cobra.Command, args []string) error {
               // implementation
           },
       }
   }
   ```
2. Register it in `rootCmd()`: `root.AddCommand(myCmd())`
3. Add tests in `main_test.go`

## Adding a New Report Format

1. Add a generation function in `hyoka/internal/report/`
2. Call it from the engine's report-writing section in `engine.go`
3. Add format-specific tests

## Conventions

- Use Go standard library where possible (`log/slog`, `net/http`, `html/template`)
- Return errors up the call stack; don't log-and-return
- User-facing output → stdout/stderr directly
- Diagnostic logging → `log/slog`
- CLI framework: `github.com/spf13/cobra`
- Config format: YAML with `gopkg.in/yaml.v3`

## Git Workflow

- Branch naming: `{user}/issue-{N}-{description}` or `{user}/dev`
- Always include co-author trailer: `Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>`
