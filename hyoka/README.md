# hyoka — CLI Reference

The `hyoka` tool evaluates AI agent code generation quality by running prompts from the `azure-sdk-prompts` library through configurable Copilot sessions, verifying code with Copilot-based verification, reviewing code via a multi-model review panel with criteria-based pass/fail scoring, and generating JSON, HTML, and Markdown reports.

## Prerequisites

- **Go 1.24.5+** — to build and run the tool
- **GitHub Copilot CLI** — the SDK communicates with Copilot via the CLI in server mode. Must be installed and authenticated:
  - Install: follow [GitHub Copilot CLI setup](https://docs.github.com/en/copilot/how-tos/set-up/install-copilot-cli)
  - Authenticate: run `copilot` once to complete OAuth device flow, or set `COPILOT_GITHUB_TOKEN` / `GH_TOKEN` env var
  - Without this, the tool falls back to stub mode (no real evaluations)
- **GitHub CLI (`gh`)** — optional but recommended for auth token management
- **For `azure-mcp` config:** `npx` (Node.js) must be available since the Azure MCP server is launched via `npx -y @azure/mcp@latest`

## Installation

### Run from source (recommended for development)

```bash
cd azure-sdk-prompts
go run ./hyoka <command> [flags]
```

### Install globally

```bash
go install github.com/ronniegeraghty/azure-sdk-prompts/hyoka/cmd/hyoka@latest
hyoka <command> [flags]
```

> **Pinned version:** `go install github.com/ronniegeraghty/azure-sdk-prompts/hyoka/cmd/hyoka@hyoka/v0.3.0`

## Features

### Phase 1 (v0.1.0) ✅
- Prompt library loading, filtering, and validation
- Build verification for 9 languages (dotnet, Python, Go, Java, JS, TS, Rust, C++)
- JSON report generation with directory hierarchy
- Manifest generation and prompt validation

### Phase 2 (v0.2.0) ✅
- **Copilot SDK integration** — Real code generation via `github.com/github/copilot-sdk/go`
- **Multi-model review panel** — Multiple reviewer models evaluate code in parallel; first model consolidates with majority-vote consensus
- **Criteria-based pass/fail scoring** — General criteria (Code Builds, Latest Package Versions, Best Practices, Error Handling, Code Quality) plus prompt-specific criteria from `## Evaluation Criteria` sections
- **Reference answer comparison** — Optional reference code included in review prompt
- **HTML reports** — Per-evaluation reports with criteria pass/fail visualization, review panel table with inline ✅/❌ icons and hover tooltips
- **Summary dashboard** — Cross-config comparison with prompt pass rates, duration analysis by prompt, and prompt comparison section
- **Reviewer action history** — Full event logs (tool calls, build attempts, version checks) captured per reviewer
- **Graceful fallback** — Falls back to stub evaluator if Copilot CLI is unavailable

### Phase 2.1 (v0.3.0) ✅
- **Copilot-based verification** — Separate Copilot session verifies code meets requirements (replaces build-only verification as default)
- **Build verification optional** — Use `--verify-build` to also run language-specific build checks
- **Session transcripts** — Full event capture (tool calls, assistant messages, errors) in JSON + HTML reports
- **Failure diagnostics** — Failed evals show detailed error info, session events, and stub mode indicator
- **Debug mode** — `--debug` streams real-time session events to stderr (tool calls, messages, verification/review status)
- **Flat report structure** — Reports write to `reports/{timestamp}/` instead of `reports/runs/{timestamp}/`
- **Evaluation Criteria** — Parser extracts `## Evaluation Criteria` sections from prompt files for review

## Authentication

The Copilot SDK evaluator requires a running Copilot CLI with valid authentication. The SDK will:
1. Try `GITHUB_TOKEN` environment variable
2. Try the logged-in user's GitHub CLI (`gh`) auth token
3. If neither is available, fall back to the stub evaluator with a warning

Use `--stub` to explicitly skip SDK initialization and use the stub evaluator.

## Commands

### `hyoka run`

Run evaluations against the prompt library.

```bash
hyoka run [flags]
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
| `--config-file` | (auto-detected from `configs/` dir) | Path to configuration YAML |
| `--workers` | `4` | Parallel evaluation workers |
| `--timeout` | `300` | Per-prompt timeout in seconds |
| `--model` | | Override model for all configs |
| `--output` | `./reports` | Report output directory |
| `--skip-tests` | `false` | Skip test generation |
| `--skip-review` | `false` | Skip multi-model code review panel |
| `--verify-build` | `false` | Also run build verification (in addition to Copilot verification) |
| `--stub` | `false` | Force stub evaluator (no Copilot SDK) |
| `--debug` | `false` | Verbose output |
| `--dry-run` | `false` | List matches without executing |

**Examples:**

```bash
# Run all prompts with all configs (real Copilot SDK)
hyoka run

# Run with stub evaluator (no SDK needed)
hyoka run --stub

# Run storage prompts with the baseline config, skip review
hyoka run --service storage --config baseline --skip-review

# Run a single prompt
hyoka run --prompt-id storage-dp-dotnet-auth

# Compare configs
hyoka run --service storage --config baseline,azure-mcp
```

### `hyoka list`

List prompts matching the given filters (no evaluation).

```bash
hyoka list [flags]
```

Takes the same filter flags as `run`. Output shows prompt ID, service/plane/language, category, and description.

### `hyoka manifest`

(Optional) Generate a `manifest.yaml` snapshot from prompt files. The tool discovers prompts directly from the `prompts/` directory at runtime — this command is only needed to produce a static index for external tooling or documentation.

```bash
hyoka manifest [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--prompts` | `./prompts` (auto-detected) | Path to prompt library directory |
| `--output` | `./manifest.yaml` (auto-detected) | Output path for manifest |

### `hyoka validate`

Validate prompt frontmatter against the schema.

```bash
hyoka validate [flags]
```

Checks required fields, enum values, ID naming conventions, and `## Prompt` section presence. Exits with code 1 on validation failure.

### `hyoka configs`

List available tool configurations.

```bash
hyoka configs [--config-file PATH]
```

### `hyoka version`

Print the tool version.

### `hyoka check-env`

Check for required language toolchains and tools.

```bash
hyoka check-env
```

Reports availability of Python, .NET, Go, Node.js, Java, Rust, C/C++, Copilot CLI, gh authentication, and npx (for Azure MCP server). Uses ✅/❌ indicators with version strings.

## Code Review (Multi-Model Panel)

After code generation, `hyoka` runs a **multi-model review panel** — multiple reviewer models evaluate the generated code in parallel, then the first model consolidates results using majority-vote consensus. This avoids self-bias since the reviewers didn't generate the code.

### Evaluation Criteria

Each evaluation uses two sets of criteria, all scored as **pass/fail**:

**General criteria** (always applied, defined in `hyoka/internal/review/rubric.md`):

| Criterion | What it measures |
|-----------|-----------------|
| Code Builds | Does the generated code compile/build without errors? Reviewers actively attempt to build it. |
| Latest Package Versions | Are the Azure SDK packages the latest stable versions? Reviewers verify with available tools. |
| Best Practices | Azure SDK patterns (DefaultAzureCredential, proper disposal, async patterns) |
| Error Handling | Proper error handling, retries, timeouts |
| Code Quality | Clean, readable, well-structured code |

**Prompt-specific criteria** (defined per prompt in the `## Evaluation Criteria` section of each `.prompt.md` file):

Each prompt author lists what the generated code should include. These are evaluated individually as pass/fail alongside the general criteria.

### Scoring

- **Score** = count of passed criteria / total criteria (general + prompt-specific)
- **Pass** = all criteria met
- **Fail** = any criterion not met

### Review Panel

Each config file defines a `reviewer_models` list (e.g., `[claude-opus-4.6, gemini-3-pro-preview, gpt-4.1]`). All reviewers run in parallel, then:

1. The **first model** in the list acts as the **consolidator**, synthesizing all reviews into a consensus result
2. For each criterion, it passes if the **majority** of reviewers marked it passed
3. If consolidation fails, the tool falls back to a Go-based `averageReview()` with majority-vote per criterion

Reviewers actively verify code: they attempt builds, check SDK package versions, and test claims before scoring.

### Reference Answers

If a prompt has a `reference_answer` field pointing to a directory of reference code, that code is included in the review prompt for comparison.

## Report Formats

### JSON (machine-readable)

```
reports/<timestamp>/
├── summary.json          # Aggregate run statistics
└── results/
    └── <service>/<plane>/<language>/<category>/<config>/
        └── report.json   # Individual evaluation result (with criteria pass/fail and review panel)
```

### HTML (human-readable)

```
reports/<timestamp>/
├── summary.html          # Cross-config comparison matrix dashboard
└── results/
    └── <service>/<plane>/<language>/<category>/<config>/
        └── report.html   # Individual report with criteria pass/fail, review panel, and reviewer action history
```

The **summary.html** includes:
- **Prompt Comparison** — pass rates grouped by prompt across configs
- **Config Comparison** — matrix of prompt × config with pass/fail status
- **Duration Analysis** — organized by prompt, with min/max tooltips showing which config produced each result

| Prompt | baseline/sonnet | azure-mcp/sonnet |
|---|---|---|
| storage-dp-dotnet-auth | 8/8 ✅ | 8/8 ✅ |
| storage-dp-python-crud | 6/9 ❌ | 9/9 ✅ |

### Markdown (portable, git-friendly)

```
reports/<timestamp>/
├── summary.md            # Cross-config comparison matrix (Markdown)
└── results/
    └── <service>/<plane>/<language>/<category>/<config>/
        └── report.md     # Individual evaluation report (Markdown)
```

Markdown reports contain the same information as HTML reports (criteria pass/fail, review panel, tool calls, verification) in a clean, readable format suitable for viewing in GitHub, VS Code, or any Markdown renderer.

## Configuration Matrix

Configurations live in the `configs/` directory at the repo root. Each file defines **one generator model** and a shared `reviewer_models` list. All configs are auto-discovered via `LoadDir()`:

| File | Generator Model | Description |
|------|----------------|-------------|
| `configs/baseline-sonnet.yaml` | Claude Sonnet 4.5 | No MCP — raw Copilot |
| `configs/baseline-opus.yaml` | Claude Opus 4.6 | No MCP — raw Copilot |
| `configs/baseline-opus-skills.yaml` | Claude Opus 4.6 | No MCP — raw Copilot + generator skills |
| `configs/baseline-codex.yaml` | GPT Codex | No MCP — raw Copilot |
| `configs/azure-mcp-sonnet.yaml` | Claude Sonnet 4.5 | Azure MCP server attached |
| `configs/azure-mcp-opus.yaml` | Claude Opus 4.6 | Azure MCP server attached |
| `configs/azure-mcp-codex.yaml` | GPT Codex | Azure MCP server attached |

All configs use the same review panel: `reviewer_models: [claude-opus-4.6, gemini-3-pro-preview, gpt-4.1]` (claude-opus-4.6 is the consolidator).

**Sample config file:**

```yaml
configs:
  - name: baseline/claude-sonnet-4.5
    description: "Baseline — raw Copilot with Claude Sonnet 4.5"
    model: "claude-sonnet-4.5"
    reviewer_models:
      - "claude-opus-4.6"
      - "gemini-3-pro-preview"
      - "gpt-4.1"
    mcp_servers: {}
    skill_directories: []
    available_tools: []
    excluded_tools: []
```

### Config Fields

| Field | Type | SDK Mapping | Description |
|-------|------|-------------|-------------|
| `name` | string | — | Unique config identifier |
| `description` | string | — | Human-readable description |
| `model` | string | `SessionConfig.Model` | Generator AI model |
| `reviewer_models` | list | — | Review panel models (first is consolidator) |
| `mcp_servers` | map | `SessionConfig.MCPServers` | MCP server definitions |
| `generator_skill_directories` | list | `SessionConfig.SkillDirectories` | Skill directories for the generator agent (takes priority over `skill_directories`) |
| `reviewer_skill_directories` | list | `SessionConfig.SkillDirectories` | Skill directories for the review panel agents |
| `skill_directories` | list | `SessionConfig.SkillDirectories` | Shared fallback skill directories (used when role-specific fields are not set) |
| `available_tools` | list | `SessionConfig.AvailableTools` | Allowed tool names |
| `excluded_tools` | list | `SessionConfig.ExcludedTools` | Blocked tool names |

#### Skill Directory Resolution

The tool resolves skill directories per role:

- **Generator:** Uses `generator_skill_directories` if set, otherwise falls back to `skill_directories`
- **Reviewers:** Uses `reviewer_skill_directories` if set, otherwise falls back to `skill_directories`

This allows configs to load different skills for generation vs. review. For example, the generator might get SDK-specific coding skills while reviewers get build-verification and version-checking skills.

**Adding skills:** Use `npx skills add microsoft/skills --directory skills/generator` to install skills from the [microsoft/skills](https://github.com/microsoft/skills) registry. See the main [README.md](../README.md#adding-skills-to-configs) for details.

**Example — separate skills per role:**

```yaml
configs:
  - name: full-skills/claude-opus-4.6
    description: "Generator and reviewer skills"
    model: "claude-opus-4.6"
    reviewer_models:
      - "claude-opus-4.6"
      - "gemini-3-pro-preview"
      - "gpt-4.1"
    generator_skill_directories:
      - "./skills/generator"
    reviewer_skill_directories:
      - "./skills/reviewer"
```

## Smart Path Detection

`hyoka` automatically resolves paths when flags aren't explicitly set:

| Flag | Candidates checked |
|------|--------------------|
| `--prompts` | `./prompts` → `../prompts` |
| `--config-file` | `./configs/` → `../configs/` (auto-discovered directory) |
| `--output` (manifest) | `./manifest.yaml` → `../manifest.yaml` (optional snapshot only) |

## Project Structure

```
hyoka/
├── cmd/hyoka/main.go        # CLI entry point (cobra)
├── go.mod / go.sum
├── internal/
│   ├── checkenv/                # Environment check (check-env command)
│   ├── config/                  # Config file parsing
│   ├── prompt/                  # Prompt loading, parsing, filtering
│   ├── eval/                    # Engine, workspace, CopilotSDKEvaluator
│   ├── build/                   # Build verification per language
│   ├── report/                  # JSON + HTML report generation
│   ├── review/                  # Multi-model review panel + criteria scoring
│   │   └── rubric.md            # Criteria-based rubric (embedded via go:embed)
│   ├── verify/                  # Copilot-based code verification
│   ├── manifest/                # Optional manifest snapshot generation
│   └── validate/                # Prompt frontmatter validation
└── testdata/                    # Test fixtures
```

## Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ Done | Prompt library, build verification, JSON reports (stub evaluator) |
| Phase 2 | ✅ Done | Copilot SDK integration, multi-model review panel, criteria-based scoring, HTML reports |
| Phase 2.1 | ✅ Done | Copilot verification, session transcripts, debug mode, failure diagnostics |
| Phase 3 | ✅ Done | Tool matrix, MCP server attachment, skill loading, cross-config comparison |
| Phase 4 | In Progress | Evaluation quality — check-env, expected_tools, reviewer skills |
| Phase 5 | Planned | Polish — report re-rendering, embedded CLI, progress bars |
