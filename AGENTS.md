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

## Running Evaluations

### Config Naming Convention

Config YAML files live in `configs/`. The `--config` flag takes the `name:` field from **inside** the YAML file, **NOT** the filename.

Example: `configs/azure-mcp-opus.yaml` contains `name: azure-mcp/claude-opus-4.6` → use `--config azure-mcp/claude-opus-4.6`

**All current config names:**

| Config name (`--config` value)       | File                          |
|--------------------------------------|-------------------------------|
| `baseline/claude-sonnet-4.5`        | `baseline-sonnet.yaml`        |
| `baseline/claude-opus-4.6`          | `baseline-opus.yaml`          |
| `baseline-skills/claude-opus-4.6`   | `baseline-opus-skills.yaml`   |
| `baseline/gpt-5.3-codex`            | `baseline-codex.yaml`         |
| `azure-mcp/claude-opus-4.6`         | `azure-mcp-opus.yaml`         |
| `azure-mcp/claude-sonnet-4.5`       | `azure-mcp-sonnet.yaml`       |
| `azure-mcp/gpt-5.3-codex`           | `azure-mcp-codex.yaml`        |

### Prompt ID Patterns

- `--prompt-id` accepts a **single** prompt ID (not multiple, not comma-separated)
- Prompt IDs follow the pattern: `{service}-{plane-abbrev}-{language}-{short-name}`
  - e.g., `identity-dp-python-default-credential`, `key-vault-dp-python-crud-secrets`
  - `dp` = data-plane, `mp` = management-plane
- To run multiple prompts, use filter flags: `--service`, `--language`, `--plane`, `--category`

### Command Examples

```bash
# Single prompt, single config:
go run ./hyoka run --prompt-id identity-dp-python-default-credential \
  --config baseline/claude-opus-4.6

# Single prompt, multiple configs (MUST quote comma-separated values):
go run ./hyoka run --prompt-id identity-dp-python-default-credential \
  --config "baseline/claude-opus-4.6,azure-mcp/claude-opus-4.6"

# Filter by service + language (runs ALL matching prompts):
go run ./hyoka run --service key-vault --language python \
  --config "baseline/claude-opus-4.6,azure-mcp/claude-opus-4.6"

# Full debug logging with log file:
go run ./hyoka run --service identity --language python \
  --config azure-mcp/claude-opus-4.6 \
  --log-level debug --log-file hyoka-debug.log

# Dry run (list matching prompts without executing):
go run ./hyoka run --service storage --language dotnet --dry-run

# All configs (requires explicit --all-configs flag):
go run ./hyoka run --prompt-id identity-dp-python-default-credential --all-configs

# With resource monitoring:
go run ./hyoka run --service key-vault --language python \
  --config azure-mcp/claude-opus-4.6 --monitor-resources
```

### Important Flag Rules

- `--config` values with commas **MUST** be quoted: `--config "config1,config2"`
- `--prompt-id` is singular — pass **ONE** ID only
- `--tags` is also comma-separated and must be quoted: `--tags "auth,crud"`
- Without `--config` or `--all-configs`, the run will fail
- `--log-level debug` enables verbose logging; pair with `--log-file` to capture to file
- `--max-session-actions` (default: 50) limits actions per Copilot session

### Available Filter Flags

```
--service        Azure service (e.g., identity, key-vault, storage, cosmos-db)
--language       Programming language (e.g., python, dotnet, java, js-ts, go, rust, cpp)
--plane          data-plane or management-plane
--category       Use-case category (e.g., auth, crud, pagination)
--tags           Comma-separated tags (must quote)
--prompt-id      Single prompt ID
```

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
