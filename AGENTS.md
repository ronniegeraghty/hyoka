# Agent Instructions for hyoka

## Overview

hyoka is a Go CLI tool that evaluates AI agents generating Azure SDK code. It uses GitHub Copilot sessions to generate code from prompts, then runs a multi-model review panel to score the output.

## Repository Structure

```
hyoka/              # Go source (module: github.com/ronniegeraghty/hyoka)
  main.go           # CLI entry point (cobra)
  internal/         # All packages
    build/          # Language-specific build verification
    checkenv/       # Environment prerequisite validation
    clean/          # Session state & orphan process cleanup (#62, #70)
    config/         # Config loading & parsing
    criteria/       # Tiered evaluation criteria system (#30)
    eval/           # Evaluation engine (generation + review orchestration)
    history/        # Run history tracking
    logging/        # Structured slog logging utilities
    manifest/       # Dependency manifest
    plugin/         # Composable plugin system (#50)
    progress/       # Progress display (live, log, off)
    prompt/         # Prompt loading, filtering, validation
    rerender/       # Report re-rendering from JSON
    report/         # Report generation (JSON, HTML, Markdown)
    review/         # Multi-model review panel + rubric
    serve/          # Local web server for report browsing (#20)
    skills/         # Skill fetching (local + remote)
    trends/         # Cross-run trend analysis
    utils/          # Shared utility functions
    validate/       # Prompt schema validation
configs/            # Evaluation config YAML files
prompts/            # Prompt library (organized by language/service)
skills/             # Copilot skills (generator/ and reviewer/)
reports/            # Generated evaluation output (gitignored)
docs/               # Design docs and getting started guide
```

## Build & Test

```bash
# Build (from repo root — uses go.work)
cd /home/rgeraghty/projects/hyoka
go build ./hyoka/...

# Run tests
go test ./hyoka/...

# Run the CLI
go run ./hyoka <command>

# Common commands
go run ./hyoka list
go run ./hyoka run --all-configs
go run ./hyoka validate
go run ./hyoka check-env
go run ./hyoka clean
```

Go version: 1.24.5+ required. Module path: `github.com/ronniegeraghty/hyoka`.

## Git Workflow

- **Branch naming**: `{username}/issue-{N}-{short-description}`
- **Commit trailers**: Always include `Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>`
- **Git identity**: Use `ronniegeraghty` account (not EMU). Set:
  ```
  git config user.name "ronniegeraghty"
  git config user.email "ronniegeraghty@users.noreply.github.com"
  ```
- **Push auth**: Run `gh auth switch --user ronniegeraghty` before pushing

## Coding Conventions

- **Go standard library preferred** — use `log/slog` for logging, `net/http` for HTTP, etc.
- **CLI framework**: `github.com/spf13/cobra`
- **Config format**: YAML with `gopkg.in/yaml.v3`
- **No third-party logging** — use `log/slog` (Go 1.21+)
- **User-facing output** goes to stdout/stderr directly (progress bars, results)
- **Diagnostic logging** goes through slog
- **Error handling**: Return errors up the call stack; don't log-and-return

## Key Architectural Patterns

- **Multi-model review panel**: Multiple LLMs review generated code independently, then a consolidator merges scores
- **Config-driven evaluation**: Each YAML config defines a generator model, reviewer models, skills, and MCP servers
- **Prompt frontmatter**: Prompts have YAML frontmatter with `id`, `service`, `language`, `plane`, `category`, `difficulty`
- **Guardrails**: Turn limits (25), file limits (50), output size limits (1 MB)

## Board Integration

Issues are tracked on Azure/projects/424. When starting work:
1. Set Status → In Progress
2. Set Squad Member → Copilot
3. Create branch, implement, push, open PR
4. Set Status → Done when merged
