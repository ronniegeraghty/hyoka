# Getting Started with azure-sdk-prompts

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
git clone https://github.com/ronniegeraghty/azure-sdk-prompts.git
cd azure-sdk-prompts
```

The repo uses a `go.work` file, so all commands run from the repo root — no need to `cd tool/`.

Verify the setup:

```bash
go run ./tool/cmd/azsdk-prompt-eval version
```

Expected output:
```
azsdk-prompt-eval version 0.6.0
```

Check your environment:

```bash
go run ./tool/cmd/azsdk-prompt-eval check-env
```

This reports which language toolchains and tools are available.

## 2. Explore Available Prompts

```bash
# List all prompts
go run ./tool/cmd/azsdk-prompt-eval list

# Filter by service
go run ./tool/cmd/azsdk-prompt-eval list --service storage

# JSON output (for scripting)
go run ./tool/cmd/azsdk-prompt-eval list --json
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
go run ./tool/cmd/azsdk-prompt-eval run \
  --prompt-id storage-dp-dotnet-auth \
  --config baseline
```

Or use **stub mode** to test the pipeline without Copilot:

```bash
go run ./tool/cmd/azsdk-prompt-eval run \
  --prompt-id storage-dp-dotnet-auth \
  --stub
```

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
go run ./tool/cmd/azsdk-prompt-eval trends
```

This scans all past runs and produces:
- Pass rate timelines
- Duration trends
- Config comparisons
- AI-powered insights (enabled by default)

Open the trend report:

```bash
go run ./tool/cmd/azsdk-prompt-eval trends --open
```

## 6. Create a New Prompt

Use the interactive scaffolder:

```bash
go run ./tool/cmd/azsdk-prompt-eval new-prompt
```

Or copy the template manually:

```bash
cp templates/prompt-template.prompt.md \
   prompts/<service>/<plane>/<language>/<slug>.prompt.md
```

Validate after editing:

```bash
go run ./tool/cmd/azsdk-prompt-eval validate
```

## Common Workflows

### Run a full evaluation matrix

```bash
# All prompts × all configs (baseline + azure-mcp)
go run ./tool/cmd/azsdk-prompt-eval run
```

### Run with specific configs

```bash
# Just baseline
go run ./tool/cmd/azsdk-prompt-eval run --config baseline

# Both configs for one service
go run ./tool/cmd/azsdk-prompt-eval run --service storage
```

### Re-render reports after template changes

```bash
go run ./tool/cmd/azsdk-prompt-eval report --all
```

### Skip AI analysis for faster iteration

```bash
go run ./tool/cmd/azsdk-prompt-eval run --skip-trends
go run ./tool/cmd/azsdk-prompt-eval trends --no-analyze
```

## Next Steps

- Read the [root README](../README.md) for full command reference
- Check out `skills/prompt-authoring/SKILL.md` for prompt writing best practices
- See `docs/cleanup-plan.md` for the project roadmap
