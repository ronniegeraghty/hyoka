# CLI Reference

hyoka provides several commands for evaluating AI-generated code quality.

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--log-level` | `warn` | Log level: `debug`, `info`, `warn`, `error` |
| `--log-file` | (stderr) | Redirect log output to a file |

## Commands

### `hyoka run`

Run evaluations against the prompt library.

```bash
# Run all prompts with a single config
hyoka run --config baseline

# Run all configs (requires explicit flag)
hyoka run --all-configs

# Filter by service, language, etc.
hyoka run --config baseline --service storage --language python

# Run a specific prompt
hyoka run --config baseline --prompt-id storage-dp-python-crud

# Dry run — list matching prompts without running
hyoka run --config baseline --dry-run
```

#### Filter Flags

| Flag | Description |
|------|-------------|
| `--prompts` | Path to prompt library directory (default: `./prompts`) |
| `--service` | Filter by Azure service (e.g., `storage`, `key-vault`) |
| `--language` | Filter by language (e.g., `python`, `java`, `dotnet`, `go`) |
| `--plane` | Filter by `data-plane` or `management-plane` |
| `--category` | Filter by use-case category |
| `--tags` | Filter by tags (comma-separated) |
| `--prompt-id` | Run a single prompt by ID |

#### Config Flags

| Flag | Description |
|------|-------------|
| `--config` | Config name(s) from config file (comma-separated) |
| `--config-file` | Path to a specific YAML config file |
| `--config-dir` | Directory containing config YAML files (default: `./configs`) |
| `--all-configs` | Required when running multiple configs |
| `--model` | Override model for all configs |

#### Execution Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workers` | CPU count (max 8) | Parallel evaluation workers |
| `--max-sessions` | workers × 3 | Maximum concurrent Copilot sessions |
| `--generate-timeout` | 600s | Generation phase timeout |
| `--build-timeout` | 300s | Build verification timeout |
| `--review-timeout` | 300s | Review phase timeout |
| `--output` | `./reports` | Report output directory |
| `--progress` | `auto` | Progress display: `auto`, `live`, `log`, `off` |
| `--stub` | false | Use stub evaluator (no Copilot SDK) |
| `-y`, `--yes` | false | Skip large run confirmation |

#### Feature Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--skip-tests` | false | Skip test generation |
| `--skip-review` | false | Skip code review phase |
| `--skip-trends` | false | Skip trend analysis after run |
| `--verify-build` | false | Run build verification on generated code |
| `--dry-run` | false | List matches without running |
| `--monitor-resources` | false | Track CPU/memory of Copilot sessions |
| `--strict-cleanup` | false | Fail if orphaned processes remain after cleanup |
| `--allow-cloud` | false | Allow real Azure resource provisioning |

#### Guardrail Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--max-turns` | 25 | Maximum assistant turns per generation |
| `--max-files` | 50 | Maximum generated files per evaluation |
| `--max-output-size` | 1MB | Maximum total output size (e.g., `512KB`, `2MB`) |
| `--max-session-actions` | 50 | Maximum actions per Copilot session (reasoning, response, or tool call each count as 1) |

#### Criteria Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--criteria-dir` | (none) | Directory with attribute-matched criteria YAML files |
| `--exclude-dirs` | (none) | Comma-separated directories to exclude from generated_files output |

> **Config Names:** The `--config` flag takes config *names* (the `name:` field inside the YAML config file), not filenames. For example, `configs/azure-mcp-opus.yaml` defines `name: azure-mcp/claude-opus-4.6`. Run it with `--config azure-mcp/claude-opus-4.6`.
>
> **Quoting Rules:** When passing multiple comma-separated values to `--config` or `--tags`, wrap the value in quotes: `--config "baseline/claude-opus-4.6,azure-mcp/claude-opus-4.6"`.

### `hyoka list`

List available prompts from the prompt library.

```bash
hyoka list
hyoka list --service storage
hyoka list --language python --plane data-plane
```

### `hyoka configs`

List and describe available configurations.

```bash
hyoka configs
hyoka configs --config-dir ./my-configs
```

### `hyoka validate`

Validate prompt files and config files for schema compliance.

```bash
hyoka validate
```

### `hyoka check-env`

Verify environment prerequisites (Go, Node.js, Copilot CLI, GitHub CLI).

```bash
hyoka check-env
```

### `hyoka trends`

Analyze trends across evaluation runs.

```bash
hyoka trends
hyoka trends --reports-dir ./reports
```

### `hyoka report`

Re-render reports from a previous run.

```bash
hyoka report 20260327-113302
hyoka report --all
```

### `hyoka serve`

Start a local web server to browse evaluation reports.

```bash
hyoka serve
hyoka serve --port 9090
hyoka serve --output ./my-reports
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | 8080 | Port to serve on |
| `--output` | `./reports` | Reports directory |

### `hyoka new-prompt`

Scaffold a new prompt file interactively.

```bash
hyoka new-prompt
```

### `hyoka version`

Print the hyoka version.
