# azure-sdk-prompts

A curated library of prompts for evaluating how well AI agents generate Azure SDK code, paired with a Go evaluation tool (`azsdk-prompt-eval`) that runs prompts through the Copilot SDK, reviews code via a multi-model panel, and produces criteria-based pass/fail reports.

## Quick Start

### Prerequisites

- **Go 1.24.5+** — to build and run the tool
- **GitHub Copilot CLI** — the SDK communicates with Copilot via the CLI in server mode. Must be installed and authenticated:
  - Install: follow [GitHub Copilot CLI setup](https://docs.github.com/en/copilot/how-tos/set-up/install-copilot-cli)
  - Authenticate: run `copilot` once to complete OAuth device flow, or set `COPILOT_GITHUB_TOKEN` / `GH_TOKEN` env var
  - Without this, the tool falls back to stub mode (no real evaluations)
- **GitHub CLI (`gh`)** — optional but recommended for auth token management
- **For `azure-mcp` config:** `npx` (Node.js) must be available since the Azure MCP server is launched via `npx -y @azure/mcp@latest`

### Run from the repo (recommended)

The repo root has a `go.work` file, so all commands run from the repo root:

```bash
git clone https://github.com/ronniegeraghty/azure-sdk-prompts.git
cd azure-sdk-prompts

# List prompts
go run ./tool/cmd/azsdk-prompt-eval list

# Run all evaluations (auto-generates trend analysis after)
go run ./tool/cmd/azsdk-prompt-eval run

# Filter by service and language
go run ./tool/cmd/azsdk-prompt-eval run --service storage --language dotnet
```

### Install as a CLI

```bash
go install github.com/ronniegeraghty/azure-sdk-prompts/tool/cmd/azsdk-prompt-eval@latest

# When run from the repo root, prompts are auto-detected
cd azure-sdk-prompts
azsdk-prompt-eval run --service storage

# Or specify the prompts path explicitly
azsdk-prompt-eval run --prompts ~/projects/azure-sdk-prompts/prompts
```

> **Smart path detection:** `azsdk-prompt-eval` checks `./prompts` then `../prompts` automatically. Running from the repo root or the `tool/` directory both work without extra flags.

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `azsdk-prompt-eval run` | | Run evaluations against prompts |
| `azsdk-prompt-eval list` | `ls` | List prompts matching filters |
| `azsdk-prompt-eval configs` | | Show available tool configurations |
| `azsdk-prompt-eval validate` | | Validate prompt frontmatter against schema |
| `azsdk-prompt-eval check-env` | `env` | Check for required language toolchains and tools |
| `azsdk-prompt-eval trends` | | Generate historical trend reports with AI analysis |
| `azsdk-prompt-eval report` | | Re-render HTML/MD reports from existing JSON data |
| `azsdk-prompt-eval new-prompt` | | Scaffold a new prompt file interactively |
| `azsdk-prompt-eval version` | | Print version |

### Filtering

All filter flags work with `run`, `list`, and other prompt-aware commands:

```bash
# By service
azsdk-prompt-eval run --service storage

# By language
azsdk-prompt-eval run --language dotnet

# Combine filters (AND logic)
azsdk-prompt-eval run --service storage --language dotnet --plane data-plane

# By category
azsdk-prompt-eval run --category authentication

# By tags
azsdk-prompt-eval run --tags identity

# Single prompt by ID
azsdk-prompt-eval run --prompt-id storage-dp-dotnet-auth

# Dry run — list matches without executing
azsdk-prompt-eval run --service storage --dry-run

# JSON output for scripting
azsdk-prompt-eval list --json
```

### Run Command Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--analyze` | `true` | AI-powered trend analysis after run |
| `--skip-trends` | `false` | Skip automatic trend analysis after run |
| `--progress` | `auto` | Progress display mode: `auto`, `live`, `log`, `off` |
| `--skip-tests` | `false` | Skip test generation |
| `--skip-review` | `false` | Skip code review |
| `--verify-build` | `false` | Run build verification (in addition to Copilot verification) |
| `--stub` | `false` | Use stub evaluator (no Copilot SDK) |
| `--dry-run` | `false` | List matching prompts without running |
| `--workers` | `4` | Parallel evaluation workers |
| `--timeout` | `300` | Per-prompt timeout in seconds |

### Validating Prompts

```bash
# Validate all prompts
azsdk-prompt-eval validate
```

### Tool Configurations

Each config file defines **one generator model** and a **multi-model review panel**. The `configs/` directory contains 6 configs (3 baseline + 3 azure-mcp), auto-discovered via `LoadDir()`:

```bash
# List configs
azsdk-prompt-eval configs

# Run with a specific config file
azsdk-prompt-eval run --config-file configs/baseline-sonnet.yaml --prompt-id storage-dp-dotnet-auth

# Run all configs (default — auto-discovers configs/ directory)
azsdk-prompt-eval run --prompt-id storage-dp-dotnet-auth

# Run with a specific config name
azsdk-prompt-eval run --config baseline/claude-sonnet-4.5

# Run multiple configs (produces comparison data)
azsdk-prompt-eval run --config baseline/claude-sonnet-4.5,azure-mcp/claude-sonnet-4.5
```

#### Custom Configs

Create your own config YAML in the `configs/` directory. Each file defines one generator model and its review panel:

```yaml
configs:
  - name: my-custom-config
    description: "My custom evaluation config"
    model: "claude-sonnet-4.5"
    reviewer_models:
      - "claude-opus-4.6"        # first model acts as consolidator
      - "gemini-3-pro-preview"
      - "gpt-4.1"
    mcp_servers: {}
    skill_directories: []
    available_tools: []
    excluded_tools: []
```

Then run with: `azsdk-prompt-eval run --config-file configs/my-custom-config.yaml`

## Adding a New Prompt

Add a `.prompt.md` file to `prompts/` and run the tool — it discovers prompts automatically.

```bash
# 1. Copy the template
cp templates/prompt-template.prompt.md \
   prompts/<service>/<plane>/<language>/<use-case>.prompt.md

# 2. Edit the file — fill in frontmatter and write your prompt

# 3. Validate
go run ./tool/cmd/azsdk-prompt-eval validate

# 4. Commit
git add prompts/
git commit -m "prompt: add <service> <plane> <language> <category>"
```

## Repo Structure

```
azure-sdk-prompts/
├── README.md
├── go.work                            # Go workspace (run commands from repo root)
├── configs/                           # Evaluation configs (one generator per file)
│   ├── baseline-sonnet.yaml           # Baseline + Claude Sonnet 4.5
│   ├── baseline-opus.yaml             # Baseline + Claude Opus 4.6
│   ├── baseline-codex.yaml            # Baseline + GPT Codex
│   ├── azure-mcp-sonnet.yaml          # Azure MCP + Claude Sonnet 4.5
│   ├── azure-mcp-opus.yaml            # Azure MCP + Claude Opus 4.6
│   └── azure-mcp-codex.yaml           # Azure MCP + GPT Codex
├── prompts/                           # Prompt library
│   ├── storage/
│   │   ├── data-plane/
│   │   │   ├── dotnet/
│   │   │   │   └── authentication.prompt.md
│   │   │   └── python/
│   │   │       └── pagination-list-blobs.prompt.md
│   │   └── management-plane/
│   │       └── ...
│   └── key-vault/
│       └── ...
├── skills/                            # Copilot skills for eval sessions
│   ├── generator/                     # Skills loaded only for the generator agent
│   └── reviewer/                      # Skills loaded only for the review agent
│       ├── code-review-comments/
│       ├── reviewer-build/
│       └── sdk-version-check/
├── tool/                              # Go eval tool (azsdk-prompt-eval)
│   ├── cmd/azsdk-prompt-eval/main.go
│   ├── go.mod / go.sum
│   └── internal/                      # config, prompt, eval, build, report,
│       │                              #   validate, trends, verify, review
│       ├── config/
│       ├── prompt/
│       ├── eval/
│       ├── build/
│       ├── report/
│       ├── trends/
│       ├── verify/
│       ├── review/
│       │   └── rubric.md              # Criteria-based scoring rubric (embedded)
│       └── validate/
├── reports/                           # Evaluation output
│   └── <run-id>/
│       ├── summary.{json,html,md}
│       └── results/<service>/<plane>/<language>/<category>/<config>/
│           └── report.{json,html,md}
├── docs/                              # Documentation
│   ├── getting-started.md
│   └── cleanup-plan.md
└── templates/
    └── prompt-template.prompt.md
```

## Tagging System

Every prompt uses YAML frontmatter:

| Field | Required | Values |
|---|---|---|
| `id` | ✅ | `{service}-{dp\|mp}-{lang}-{category-slug}` |
| `service` | ✅ | `storage`, `key-vault`, `cosmos-db`, `event-hubs`, `app-configuration`, `purview`, `digital-twins`, `identity`, `resource-manager`, `service-bus` |
| `plane` | ✅ | `data-plane`, `management-plane` |
| `language` | ✅ | `dotnet`, `java`, `js-ts`, `python`, `go`, `rust`, `cpp` |
| `category` | ✅ | `authentication`, `pagination`, `polling`, `retries`, `error-handling`, `crud`, `batch`, `streaming`, `auth`, `provisioning` |
| `difficulty` | ✅ | `basic`, `intermediate`, `advanced` |
| `description` | ✅ | What this prompt tests (1-3 sentences) |
| `created` | ✅ | Date (YYYY-MM-DD) |
| `author` | ✅ | GitHub username |
| `sdk_package` | ❌ | SDK package name |
| `doc_url` | ❌ | Library reference docs (API overview, pkg.go.dev, docs.rs) |
| `tags` | ❌ | Free-form tags for additional filtering |

## Roadmap

- **Phase 1:** ✅ Prompt library, build verification, report generation with stub evaluator
- **Phase 2:** ✅ Copilot SDK integration — live agent evaluation with code generation and criteria-based review panel
- **Phase 3:** ✅ Tool matrix, MCP server attachment, skill loading, cross-config comparison
- **Phase 4:** In progress — Evaluation quality (check-env, expected_tools, reviewer skills)
- **Phase 5:** Planned — Polish (report re-rendering, embedded CLI, progress bars)

See [`tool/README.md`](tool/README.md) for full CLI reference and configuration docs.

## License

[MIT](LICENSE)
