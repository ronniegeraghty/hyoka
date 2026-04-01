# Getting Started with hyoka

This guide walks you through cloning the repo, running your first evaluation, and viewing results.

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Go | 1.24.5+ | `go version` |
| GitHub Copilot CLI | Latest | `copilot --version` |
| Git | Any | `git --version` |
| Node.js (for Azure MCP config) | 18+ | `node --version` |

### Copilot Authentication

The tool uses the Copilot SDK which requires an authenticated Copilot CLI:

```bash
# Option A: OAuth device flow (interactive)
copilot

# Option B: Environment variable
export COPILOT_GITHUB_TOKEN="your-token"
```

Without Copilot auth, the tool falls back to **stub mode** (no real agent evaluations).

## 1. Clone and Verify

```bash
git clone https://github.com/ronniegeraghty/hyoka.git
cd hyoka
```

The repo uses a `go.work` file, so all commands run from the repo root — no need to `cd hyoka/`.

Verify the setup:

```bash
go run ./hyoka version
```

Expected output:
```
hyoka version 0.2.0
```

Check your environment:

```bash
go run ./hyoka check-env
```

This reports which language toolchains and tools are available.

## 2. Explore Available Prompts

```bash
# List all prompts
go run ./hyoka list

# Filter by service
go run ./hyoka list --service storage

# JSON output (for scripting)
go run ./hyoka list --json
```

Expected output:
```
Found 79 prompt(s):

  storage-dp-dotnet-auth         storage/data-plane/dotnet [authentication]
                                 Can the docs help a developer authenticate to Azure Blob Storage...
  ...
```

## 3. Run Your First Evaluation

Start with a single prompt to keep it quick:

```bash
go run ./hyoka run \
  --prompt-id storage-dp-dotnet-auth \
  --config baseline
```

Or use **stub mode** to test the pipeline without Copilot:

```bash
go run ./hyoka run \
  --prompt-id storage-dp-dotnet-auth \
  --stub
```

> **Confirmation prompt:** If a run would execute more than 10 evaluations, hyoka asks for confirmation before proceeding. Use `-y` to skip in CI or scripted runs. If you run without a `--config` filter and multiple configs exist, add `--all-configs` to confirm you intend to run all of them.

Expected output:
```
Found 1 prompt(s), 1 config(s) → 1 evaluation(s)
Using Copilot SDK evaluator

Run Summary:
  Run ID:      20250728-143022
  Evaluations: 1
  Passed:      1
  Failed:      0
  Errors:      0
  Duration:    45.20s

────────────────────────────────────────────────────────
📊 Generating trend analysis...
...
```

## 4. View Results

Reports are generated in `reports/<run-id>/`:

```bash
# Open the summary HTML in your browser
open reports/<run-id>/summary.html    # macOS
xdg-open reports/<run-id>/summary.html  # Linux
```

The summary includes:
- **Prompt × Config Matrix** — pass/fail status with scores
- **Duration Analysis** — min/avg/max per config and prompt
- **Config Comparison** — side-by-side pass rates
- **Tool Usage** — aggregate tool call statistics
- **Detailed Results** — individual eval links

Individual reports at `reports/<run-id>/results/.../report.html` show the full agent session: prompt, reasoning, tool calls, generated code, verification, and review scores.

## 5. Run Trend Analysis

After multiple runs, generate trend reports:

```bash
go run ./hyoka trends
```

This scans all past runs and produces:
- Pass rate timelines
- Duration trends
- Config comparisons
- AI-powered insights (enabled by default)

Open the trend report:

```bash
go run ./hyoka trends --open
```

## 6. Create a New Prompt

Use the interactive scaffolder:

```bash
go run ./hyoka new-prompt
```

Or copy the template manually:

```bash
cp templates/prompt-template.prompt.md \
   prompts/<service>/<plane>/<language>/<slug>.prompt.md
```

Validate after editing:

```bash
go run ./hyoka validate
```

## Common Workflows

### Run a full evaluation matrix

```bash
# All prompts × all configs (requires --all-configs since multiple configs exist)
go run ./hyoka run --all-configs

# Skip confirmation prompt for CI
go run ./hyoka run --all-configs -y
```

### Run with specific configs

```bash
# Just baseline
go run ./hyoka run --config baseline

# Both configs for one service
go run ./hyoka run --service storage
```

### Adjust guardrails

```bash
# Tighter limits for faster iteration
go run ./hyoka run --max-turns 10 --max-files 20 --max-output-size 512KB

# Allow real Azure resource provisioning
go run ./hyoka run --allow-cloud

# Limit concurrent sessions on a shared machine
go run ./hyoka run --max-sessions 4 --workers 2
```

### Re-render reports after template changes

```bash
go run ./hyoka report --all
```

### Skip AI analysis for faster iteration

```bash
go run ./hyoka run --skip-trends
go run ./hyoka trends --no-analyze
```

## Guardrails & Safety

hyoka applies sensible defaults to keep evaluation runs safe and bounded. All limits are configurable via CLI flags.

### Generator Limits

Every code-generation session is automatically aborted if it exceeds:

| Limit | Default | Flag |
|-------|---------|------|
| Conversation turns | 25 | `--max-turns` |
| Generated files | 50 | `--max-files` |
| Total output size | 1 MB | `--max-output-size` |

When a limit is hit, the evaluation stops and the report shows the specific guardrail that triggered (e.g., `guardrail: file count 51 exceeded limit of 50`).

### Safety Boundaries

By default, generators are instructed **not to provision real Azure resources**. They'll use:
- Local emulators (Azurite, CosmosDB emulator)
- Environment variable placeholders (`os.Getenv("AZURE_STORAGE_CONNECTION_STRING")`)
- Bicep/ARM/Terraform templates instead of live `az` CLI commands

To opt out: `--allow-cloud`.

### Process Cleanup

All spawned Copilot processes are tracked and terminated on run completion or Ctrl+C. The cleanup sends SIGTERM, waits up to 5 seconds, then escalates to SIGKILL.

### Prompt Discovery

If `validate` or `run` finds zero prompts, it scans for near-miss filenames and suggests fixes:
- `auth-prompt.md` → `auth.prompt.md` (hyphen instead of dot)
- `crud.prompt.txt` → `crud.prompt.md` (wrong extension)

## Browse Results with Serve

Start the built-in report viewer:

```bash
go run ./hyoka serve
```

This launches a local web server at `http://localhost:8080` with an index of all evaluation runs, linking to individual HTML reports.

## Next Steps

- [CLI Reference](cli-reference.md) — Full command and flag documentation
- [Configuration Guide](configuration.md) — Config YAML format and options
- [Prompt Authoring Guide](prompt-authoring.md) — How to write evaluation prompts
- [Guardrails and Safety](guardrails.md) — Limits, process cleanup, and safety boundaries
- [Architecture Overview](architecture.md) — How hyoka works end-to-end
- [Contributing Guide](contributing.md) — Building, testing, and adding features
