# sdk-eval — CLI Reference

The `sdk-eval` tool evaluates AI agent code generation quality by running prompts from the `azure-sdk-prompts` library through configurable Copilot sessions, verifying builds, and generating scored JSON reports.

## Installation

### Run from source (recommended for development)

```bash
cd azure-sdk-prompts
go run ./tool/cmd/sdk-eval <command> [flags]
```

### Install globally

```bash
go install github.com/ronniegeraghty/azure-sdk-prompts/tool/cmd/sdk-eval@latest
sdk-eval <command> [flags]
```

## Commands

### `sdk-eval run`

Run evaluations against the prompt library.

```bash
sdk-eval run [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--prompts` | `./prompts` (auto-detected) | Path to prompt library directory |
| `--service` | | Filter by Azure service |
| `--language` | | Filter by programming language |
| `--plane` | | Filter by data-plane / management-plane |
| `--category` | | Filter by use-case category |
| `--tags` | | Filter by tags (comma-separated) |
| `--prompt-id` | | Run a single prompt by ID |
| `--config` | all | Config name(s) (comma-separated) |
| `--config-file` | `./configs.yaml` (auto-detected) | Path to configuration YAML |
| `--workers` | `4` | Parallel evaluation workers |
| `--timeout` | `300` | Per-prompt timeout in seconds |
| `--model` | | Override model for all configs |
| `--output` | `./reports` | Report output directory |
| `--skip-tests` | `false` | Skip test generation |
| `--skip-review` | `false` | Skip code review |
| `--debug` | `false` | Verbose output |
| `--dry-run` | `false` | List matches without executing |

**Examples:**

```bash
# Run all prompts with all configs
sdk-eval run

# Run storage prompts with the baseline config
sdk-eval run --service storage --config baseline

# Run a single prompt
sdk-eval run --prompt-id storage-dp-dotnet-auth

# Compare configs
sdk-eval run --service storage --config baseline,azure-mcp
```

### `sdk-eval list`

List prompts matching the given filters (no evaluation).

```bash
sdk-eval list [flags]
```

Takes the same filter flags as `run`. Output shows prompt ID, service/plane/language, category, and description.

**Examples:**

```bash
sdk-eval list
sdk-eval list --service storage --language dotnet
```

### `sdk-eval manifest`

Regenerate `manifest.yaml` from prompt files.

```bash
sdk-eval manifest [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--prompts` | `./prompts` (auto-detected) | Path to prompt library directory |
| `--output` | `./manifest.yaml` (auto-detected) | Output path for manifest |

**Examples:**

```bash
# From repo root
sdk-eval manifest

# Explicit paths
sdk-eval manifest --prompts ./prompts --output ./manifest.yaml
```

### `sdk-eval validate`

Validate prompt frontmatter against the schema.

```bash
sdk-eval validate [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--prompts` | `./prompts` (auto-detected) | Path to prompt library directory |

Checks:
- Required fields present (id, service, plane, language, category, difficulty, description, created, author)
- Enum values valid (service, plane, language, category, difficulty)
- ID naming convention (`{service}-{dp|mp}-{language}-...`)
- `## Prompt` section present with content

Exits with code 1 on validation failure.

**Examples:**

```bash
sdk-eval validate
sdk-eval validate --prompts ./prompts
```

### `sdk-eval configs`

List available tool configurations.

```bash
sdk-eval configs [--config-file PATH]
```

### `sdk-eval version`

Print the tool version.

## Configuration Matrix

Configurations in `configs.yaml` define the Copilot environment for each evaluation:

```yaml
configs:
  - name: baseline
    description: "No MCP servers, no skills — just base Copilot"
    model: "claude-sonnet-4.5"
    mcp_servers: {}
    skill_directories: []
    available_tools: []
    excluded_tools: []

  - name: azure-mcp
    description: "Azure MCP server attached"
    model: "claude-sonnet-4.5"
    mcp_servers:
      azure:
        type: local
        command: npx
        args: ["-y", "@azure/mcp@latest"]
        tools: ["*"]
    skill_directories: []
    available_tools: []
    excluded_tools: []
```

### Config Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique config identifier |
| `description` | string | Human-readable description |
| `model` | string | AI model to use |
| `mcp_servers` | map | MCP server definitions |
| `skill_directories` | list | Paths to skill directories |
| `available_tools` | list | Allowed tool names |
| `excluded_tools` | list | Blocked tool names |

## Smart Path Detection

`sdk-eval` automatically resolves paths when flags aren't explicitly set:

| Flag | Candidates checked |
|------|--------------------|
| `--prompts` | `./prompts` → `../prompts` |
| `--config-file` | `./configs.yaml` → `../configs.yaml` |
| `--output` (manifest) | `./manifest.yaml` → `../manifest.yaml` |

This means running from the repo root or the `tool/` directory both work without extra flags.

## Report Format

Evaluations produce JSON reports:

```
reports/runs/<timestamp>/
├── summary.json          # Aggregate run statistics
└── results/
    └── <service>/<plane>/<language>/<category>/<config>/
        └── report.json   # Individual evaluation result
```

### `report.json` fields

| Field | Description |
|-------|-------------|
| `prompt_id` | Prompt identifier |
| `config_name` | Configuration used |
| `timestamp` | Evaluation start time |
| `duration_seconds` | Wall-clock time |
| `generated_files` | Files created by the agent |
| `build` | Build verification result |
| `success` | Overall pass/fail |
| `error` | Error message (if any) |

## Project Structure

```
tool/
├── cmd/sdk-eval/main.go        # CLI entry point (cobra)
├── configs.yaml                 # Default configs
├── go.mod / go.sum
├── internal/
│   ├── config/                  # Config file parsing
│   ├── prompt/                  # Prompt loading, parsing, filtering
│   ├── eval/                    # Evaluation engine + workspace
│   ├── build/                   # Build verification per language
│   ├── report/                  # JSON report generation
│   ├── manifest/                # Manifest generation from prompts
│   └── validate/                # Prompt frontmatter validation
└── testdata/                    # Test fixtures
```

## Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ Current | Prompt library, build verification, report generation (stub evaluator) |
| Phase 2 | Planned | Copilot SDK integration — live agent evaluation, code generation, LLM-as-judge scoring |
