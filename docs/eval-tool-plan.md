# Plan: Go-Based SDK Code Evaluation Tool

**Date:** 2026-07-27
**Status:** DRAFT
**Requested by:** Ronnie Geraghty

> **⚠️ UPDATE 2026-07-28:** Implementation complete. Python scripts removed. The Go
> `hyoka` tool is the sole evaluation approach. Tool lives in `hyoka/` subdirectory
> with configs at `configs/` (repo root) and reports at `reports/` (repo root). The
> standalone `hyoka-tool` repo has been deleted. Sections referencing Python scripts,
> `doc-agent evaluate`, `scripts/`, or the separate repo are historical context only.

---

## Executive Summary

This plan describes the Go-based evaluation tool for testing how well AI agents write and update Azure SDK code. The tool lives in the `ronniegeraghty/azure-sdk-prompts` monorepo alongside the prompt library. Python scripts and the `doc-agent evaluate` workflow have been removed — `hyoka` is the sole evaluation approach.

The new tool uses the **GitHub Copilot SDK for Go** (`github.com/github/copilot-sdk/go`) to programmatically create Copilot sessions, send prompts, capture generated code, verify builds, optionally generate and run tests, and score results via LLM-as-judge review. A key differentiator is the **tool configuration matrix**: each prompt can be tested against multiple configurations (different MCP servers, skills, tool sets) to measure how tooling affects code quality.

### Why Go?

- The Copilot SDK's Go package is mature and well-documented
- Go compiles to a single static binary — no Python/Node runtime dependency
- Native concurrency primitives (goroutines, channels) map perfectly to parallel eval execution
- The SDK manages the Copilot CLI process lifecycle automatically and supports embedding the CLI binary

### Why Not doc-review-agent?

| Issue | Detail |
|---|---|
| Shell-probing loops | Inner Copilot session gets stuck probing the shell environment and never executes the actual prompt |
| Architectural mismatch | doc-review-agent is designed for testing documentation walkthrough execution, not evaluating code generation quality |
| No build verification | doc-review-agent doesn't attempt to compile generated code |
| No tool matrix | No way to compare results across different MCP/skill configurations |
| No reference comparison | No structured comparison against human-written reference answers |

---

## 1. Prompt System

### 1.1 Prompt Library (Same Repo)

This repo (`ronniegeraghty/azure-sdk-prompts`) already has 57+ prompts organized by service/language/plane/category with YAML frontmatter. The Go eval tool lives in the same repo and reads prompts directly from the `prompts/` directory — no separate clone or config path required.

By default the tool looks for `./prompts` relative to the repo root. An override flag (`--prompts`) is available for pointing at alternative prompt directories during development.

### 1.2 Extended Frontmatter Schema

Add new fields to the existing prompt template. Existing prompts without the new fields use defaults (blank project, no reference answer).

```yaml
---
id: storage-dp-dotnet-auth
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: intermediate
description: >
  Test writing authentication code for Azure Blob Storage using DefaultAzureCredential.

# --- Existing fields (unchanged) ---
sdk_package: Azure.Storage.Blobs
api_version: "2024-11-04"
doc_url: https://learn.microsoft.com/...
tags: [identity, default-credential]
created: 2026-07-01
author: ronniegeraghty

# --- New fields for eval tool ---
project_context: blank          # "blank" | "existing"
starter_project: ""             # Path relative to prompt file, or empty
reference_answer: ""            # Path to reference implementation, relative to prompt file
timeout: 300                    # Per-prompt timeout override (seconds)
expected_packages: []           # SDK packages the code should use
---
```

### 1.3 Project Context: Blank vs. Existing

**Blank project** (`project_context: blank`): The eval tool creates an empty workspace directory and sends the prompt as-is. The agent starts from scratch.

**Existing project** (`project_context: existing`): The eval tool sets up a starter project before sending the prompt. The `starter_project` field points to a directory (relative to the prompt file) containing the initial code.

```
prompts/
└── storage/
    └── data-plane/
        └── dotnet/
            ├── add-error-handling.prompt.md
            └── add-error-handling.starter/       # Starter project
                ├── Program.cs
                ├── storage-app.csproj
                └── appsettings.json
```

**Implementation detail:** Before each eval run, the tool copies the starter project into the workspace directory. The Copilot session's `WorkingDirectory` is set to that workspace. The prompt then instructs the agent to modify existing code.

### 1.4 Reference Answers

Each prompt can include a human-written reference implementation for comparison:

```
prompts/
└── storage/
    └── data-plane/
        └── dotnet/
            ├── authentication.prompt.md
            └── authentication.reference/         # Reference answer
                ├── Program.cs
                └── storage-auth.csproj
```

The reference answer is used in the Code Review phase (Section 6) where a separate Copilot session compares the agent's output against the reference.

---

## 2. Evaluation Engine (Go + Copilot SDK)

### 2.1 SDK Architecture

The Copilot SDK communicates with the Copilot CLI via JSON-RPC:

```
Eval Tool (Go binary)
       ↓
  copilot.Client (SDK)
       ↓ JSON-RPC (stdio)
  Copilot CLI (server mode, auto-managed)
       ↓
  GitHub Copilot API
```

The SDK manages the CLI process lifecycle automatically. Key SDK types:

| Type | Purpose |
|---|---|
| `copilot.Client` | Manages the CLI server process and creates sessions |
| `copilot.ClientOptions` | Configures auth, working directory, log level, environment |
| `copilot.SessionConfig` | Configures model, MCP servers, skills, tools, permissions |
| `copilot.Session` | Represents a conversation; send messages, receive events |

### 2.2 Core Evaluation Flow

```
For each (prompt, config) pair:

1. Create temp workspace directory
2. If existing project: copy starter files into workspace
3. Create Copilot client (with auth)
4. Create session with:
   - Model from config
   - MCPServers from config
   - SkillDirectories from config
   - AvailableTools / ExcludedTools from config
   - WorkingDirectory = workspace
   - OnPermissionRequest = ApproveAll (eval context)
   - SystemMessage = eval-specific instructions
5. Subscribe to all session events (capture everything)
6. Send the prompt via session.SendAndWait()
7. Collect: generated files, tool calls, assistant messages, errors
8. Run build verification (Section 4)
9. Optionally run test generation (Section 5)
10. Run code review in a separate session (Section 6)
11. Generate report artifacts (Section 7)
12. Clean up workspace and disconnect session
```

### 2.3 Event Capture

Subscribe to all session events to capture a complete trace:

```go
var events []copilot.SessionEvent
unsubscribe := session.On(func(event copilot.SessionEvent) {
    events = append(events, event)
})
defer unsubscribe()
```

Key event types to capture:
- `assistant.message` — The agent's text responses
- `tool.call` / `tool.result` — Every tool invocation and result
- `session.idle` — Session completed processing
- `session.error` — Errors during processing
- `assistant.message_delta` — Streaming chunks (if enabled)

### 2.4 System Message

Use `SystemMessage` with append mode to add eval-specific instructions without replacing the default Copilot system prompt:

```go
session, err := client.CreateSession(ctx, &copilot.SessionConfig{
    SystemMessage: &copilot.SystemMessageConfig{
        Mode: "append",
        Content: `You are being evaluated on code generation quality.
Write complete, working code. Use the specified SDK packages.
Do not ask clarifying questions — make reasonable assumptions.
Write all code to the current working directory.`,
    },
    // ...
})
```

---

## 3. Tool Configuration Matrix

### 3.1 Matrix Definition (YAML)

Define configurations in a YAML file (`configs/default.yaml`):

```yaml
# configs/default.yaml — Tool configuration matrix
configs:
  - name: baseline
    description: "No MCP servers, no skills — just base Copilot"
    model: "claude-sonnet-4.5"
    mcp_servers: {}
    skill_directories: []
    available_tools: []    # empty = all built-in tools
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

  - name: azure-mcp-plus-skills
    description: "Azure MCP + Azure SDK best-practices skills"
    model: "claude-sonnet-4.5"
    mcp_servers:
      azure:
        type: local
        command: npx
        args: ["-y", "@azure/mcp@latest"]
        tools: ["*"]
    skill_directories:
      - ./skills/azure-sdk
    available_tools: []
    excluded_tools: []

  - name: no-web-search
    description: "All tools except web search"
    model: "claude-sonnet-4.5"
    mcp_servers: {}
    skill_directories: []
    available_tools: []
    excluded_tools: ["web_search"]
```

### 3.2 Mapping to SDK SessionConfig

Each config maps directly to `copilot.SessionConfig` fields:

| Config YAML field | SessionConfig field | Type |
|---|---|---|
| `model` | `Model` | `string` |
| `mcp_servers` | `MCPServers` | `map[string]MCPServerConfig` |
| `skill_directories` | `SkillDirectories` | `[]string` |
| `available_tools` | `AvailableTools` | `[]string` |
| `excluded_tools` | `ExcludedTools` | `[]string` |

### 3.3 Cross-Product Execution

The eval runner computes `len(prompts) × len(configs)` evaluation tasks. Each task is the combination of one prompt and one configuration. Results are organized by both dimensions.

---

## 4. Build Verification

### 4.1 Language-Specific Build Commands

After the agent produces code, attempt to build/compile it:

| Language | Detection | Build Command | Notes |
|---|---|---|---|
| .NET | `*.csproj` or `*.sln` | `dotnet build` | Check for `<TargetFramework>` |
| Java | `pom.xml` | `mvn compile -q` | Maven quiet mode |
| Java | `build.gradle` | `gradle compileJava` | Gradle alternative |
| Python | `*.py` | `python -m py_compile <file>` | Per-file syntax check |
| Go | `go.mod` | `go build ./...` | Build all packages |
| JS/TS | `tsconfig.json` | `npx tsc --noEmit` | Type-check only |
| JS | `package.json` (no TS) | `node --check <file>` | Syntax check |
| Rust | `Cargo.toml` | `cargo build` | |
| C++ | `CMakeLists.txt` | `cmake -B build && cmake --build build` | |

### 4.2 Build Result Capture

```go
type BuildResult struct {
    Language    string        `json:"language"`
    Command     string        `json:"command"`
    ExitCode    int           `json:"exit_code"`
    Stdout      string        `json:"stdout"`
    Stderr      string        `json:"stderr"`
    Duration    time.Duration `json:"duration_ms"`
    Success     bool          `json:"success"`
}
```

### 4.3 Dependency Installation

Before building, the tool should handle dependency installation:

- .NET: `dotnet restore` before `dotnet build`
- Go: `go mod tidy` before `go build`
- Python: `pip install -r requirements.txt` if present
- JS/TS: `npm install` if `package.json` exists
- Java: Maven/Gradle handle dependencies during build
- Rust: `cargo build` handles dependencies

Use a configurable timeout (default: 120 seconds) for dependency installation, separate from the build timeout.

---

## 5. Stretch Goal: Auto-Generated Tests

### 5.1 Flow

After code generation and build verification:

1. Create a **new** Copilot session (separate from the code generation session)
2. Set `WorkingDirectory` to the same workspace
3. Send a test-generation prompt:

```
You are a test engineer. The code in this workspace was generated by an AI agent.
Write comprehensive unit tests for this code.

Requirements:
- Mock all Azure service interactions (do not call real Azure services)
- Test error handling paths
- Test edge cases
- Use the standard test framework for this language:
  - .NET: xunit or MSTest
  - Python: pytest
  - Go: testing package
  - Java: JUnit 5
  - JS/TS: jest or vitest
```

4. After tests are generated, run them
5. Capture test results (pass/fail counts, output)

### 5.2 Test Result Capture

```go
type TestResult struct {
    Framework   string `json:"framework"`
    Command     string `json:"command"`
    TotalTests  int    `json:"total_tests"`
    Passed      int    `json:"passed"`
    Failed      int    `json:"failed"`
    Skipped     int    `json:"skipped"`
    ExitCode    int    `json:"exit_code"`
    Stdout      string `json:"stdout"`
    Stderr      string `json:"stderr"`
    Duration    time.Duration `json:"duration_ms"`
}
```

### 5.3 Why Separate Session?

Using a separate session for test generation prevents the code-generation agent from seeing its own test expectations, which would create a feedback loop. It also simulates a more realistic workflow where tests are written by a different developer.

---

## 6. Code Review (LLM-as-Judge)

### 6.1 Separate Review Session

After code is produced, launch a **separate** Copilot session for review. This avoids self-bias — the reviewer didn't generate the code.

```go
reviewSession, err := client.CreateSession(ctx, &copilot.SessionConfig{
    Model: "claude-sonnet-4.5",  // or configurable
    SystemMessage: &copilot.SystemMessageConfig{
        Mode: "append",
        Content: reviewSystemPrompt,
    },
    WorkingDirectory: workspaceDir,
    OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
})
```

### 6.2 Review Prompt Construction

The review prompt includes:
1. The original prompt that was sent to the code-generation agent
2. The generated code (files in workspace)
3. The human-written reference answer (if available)
4. A structured scoring rubric

```
You are a senior Azure SDK code reviewer. Review the generated code in this workspace.

## Original Prompt
{prompt_text}

## Reference Answer
{reference_code_or_"No reference answer provided."}

## Scoring Rubric
Score each dimension from 1-10:

1. **Correctness** — Does the code correctly implement what was asked?
2. **Completeness** — Are all requirements addressed? Missing features?
3. **Best Practices** — Does it follow Azure SDK best practices?
   (DefaultAzureCredential, proper disposal, async patterns, etc.)
4. **Error Handling** — Are errors handled properly? Retries? Timeouts?
5. **Package Usage** — Are the correct and latest SDK packages used?
6. **Code Quality** — Clean, readable, well-structured code?
7. **Reference Similarity** — How similar is it to the reference? (skip if no reference)

## Output Format (JSON)
{
  "scores": {
    "correctness": N,
    "completeness": N,
    "best_practices": N,
    "error_handling": N,
    "package_usage": N,
    "code_quality": N,
    "reference_similarity": N
  },
  "overall_score": N,
  "summary": "...",
  "issues": ["issue 1", "issue 2"],
  "strengths": ["strength 1", "strength 2"]
}
```

### 6.3 Review Result Parsing

Parse the reviewer's JSON output into a structured Go type:

```go
type ReviewResult struct {
    Scores struct {
        Correctness         int `json:"correctness"`
        Completeness        int `json:"completeness"`
        BestPractices       int `json:"best_practices"`
        ErrorHandling       int `json:"error_handling"`
        PackageUsage        int `json:"package_usage"`
        CodeQuality         int `json:"code_quality"`
        ReferenceSimilarity int `json:"reference_similarity"`
    } `json:"scores"`
    OverallScore int      `json:"overall_score"`
    Summary      string   `json:"summary"`
    Issues       []string `json:"issues"`
    Strengths    []string `json:"strengths"`
}
```

---

## 7. Report Generation

### 7.1 Per-Evaluation Report

Each (prompt, config) pair produces a report directory:

```
reports/
└── runs/
    └── 2026-07-27T14-30-00/
        ├── run-metadata.yaml           # Run config, timing, summary stats
        ├── summary.html                # Cross-prompt comparison dashboard
        └── results/
            └── storage/
                └── data-plane/
                    └── dotnet/
                        └── authentication/
                            ├── baseline/
                            │   ├── report.html
                            │   ├── report.json
                            │   ├── events.jsonl
                            │   ├── generated-code/    # Snapshot of workspace
                            │   ├── build-result.json
                            │   ├── test-result.json   # If tests were run
                            │   └── review-result.json
                            ├── azure-mcp/
                            │   └── ...
                            └── azure-mcp-plus-skills/
                                └── ...
```

### 7.2 Report Formats

**JSON** (machine-readable, primary):
```json
{
  "prompt_id": "storage-dp-dotnet-auth",
  "config_name": "azure-mcp",
  "timestamp": "2026-07-27T14:30:00Z",
  "duration_seconds": 45,
  "prompt_metadata": { ... },
  "config_used": { ... },
  "generated_files": ["Program.cs", "storage-auth.csproj"],
  "build": { "success": true, "exit_code": 0, ... },
  "tests": { "passed": 5, "failed": 0, ... },
  "review": { "overall_score": 8, "scores": { ... } },
  "event_count": 42,
  "tool_calls": ["read_file", "edit_file", "bash"]
}
```

**HTML** (human-readable dashboard):
- Use Go's `html/template` package
- Side-by-side comparison of generated code vs. reference answer
- Score visualization (bar charts or color-coded scores)
- Collapsible sections for build output, test output, event trace

**Markdown** (lightweight alternative):
- For quick terminal review via `cat` or `less`
- Include scores, summary, and key issues

### 7.3 Cross-Config Comparison

The summary report (`summary.html`) shows a matrix:

| Prompt | baseline | azure-mcp | azure-mcp-plus-skills |
|---|---|---|---|
| storage-dp-dotnet-auth | 6/10 ✅ | 8/10 ✅ | 9/10 ✅ |
| storage-dp-python-crud | 5/10 ❌ | 7/10 ✅ | 8/10 ✅ |
| keyvault-dp-go-secrets | 4/10 ❌ | 6/10 ✅ | 7/10 ✅ |

Score = overall review score. ✅/❌ = build success/failure.

### 7.4 Historical Tracking

Each run is timestamped. A `latest` symlink points to the most recent run. Over time, this enables tracking whether changes to MCP servers, skills, or models improve code generation quality.

Future enhancement: A `trends.html` page showing score trends over time per prompt/config.

---

## 8. CLI Interface

### 8.1 Command Structure

```bash
# Binary name: hyoka
hyoka run [flags]          # Run evaluations
hyoka list [flags]         # List matching prompts (dry run)
hyoka report [run-id]      # Open/regenerate report for a past run
hyoka configs              # List available configurations
hyoka version              # Print version
```

### 8.2 Filter Flags

```bash
hyoka run \
  --prompts ./prompts \
  --service storage \
  --language dotnet \
  --plane data-plane \
  --category authentication \
  --tags identity \
  --prompt-id storage-dp-dotnet-auth \
  --config baseline,azure-mcp \
  --config-file ./configs/default.yaml \
  --workers 4 \
  --timeout 300 \
  --model claude-sonnet-4.5 \
  --output ./reports \
  --skip-tests \
  --skip-review \
  --debug \
  --dry-run
```

| Flag | Description | Default |
|---|---|---|
| `--prompts` | Path to prompt library directory | `./prompts` |
| `--service` | Filter by Azure service | (all) |
| `--language` | Filter by programming language | (all) |
| `--plane` | Filter by data-plane/management-plane | (all) |
| `--category` | Filter by use-case category | (all) |
| `--tags` | Filter by tags (comma-separated) | (all) |
| `--prompt-id` | Run a single prompt by ID | (all) |
| `--config` | Config name(s) from config file | (all configs) |
| `--config-file` | Path to configuration YAML | `./configs/default.yaml` |
| `--workers` | Parallel workers | `4` |
| `--timeout` | Per-prompt timeout (seconds) | `300` |
| `--model` | Override model for all configs | (per-config) |
| `--output` | Report output directory | `./reports` |
| `--skip-tests` | Skip test generation (Section 5) | `false` |
| `--skip-review` | Skip code review (Section 6) | `false` |
| `--debug` | Verbose output (log all events) | `false` |
| `--dry-run` | List matching prompts without running | `false` |

### 8.3 Progress Output

```
hyoka run --service storage --config baseline,azure-mcp --config-file ./configs/default.yaml

Running 8 prompts × 2 configs = 16 evaluations (4 workers)

[1/16] storage-dp-dotnet-auth (baseline)     ✅ build ✅ review: 7/10  [32s]
[2/16] storage-dp-dotnet-auth (azure-mcp)    ✅ build ✅ review: 9/10  [28s]
[3/16] storage-dp-dotnet-crud (baseline)     ❌ build ⚠ review: 4/10  [45s]
...
[16/16] storage-mp-python-provision (azure-mcp) ✅ build ✅ review: 8/10 [38s]

Summary:
  Build pass rate: baseline=62% azure-mcp=87%
  Avg review score: baseline=5.8 azure-mcp=7.4
  Reports: ./reports/runs/2026-07-27T14-30-00/
```

---

## 9. Authentication

### 9.1 Copilot SDK Auth Chain

The Go SDK supports these auth methods in priority order:

1. **Explicit `GitHubToken`** — Token passed to `ClientOptions.GitHubToken`
2. **Environment variables** — `COPILOT_GITHUB_TOKEN` → `GH_TOKEN` → `GITHUB_TOKEN`
3. **Stored OAuth credentials** — From previous `copilot` CLI login
4. **GitHub CLI** — `gh auth` credentials

### 9.2 Recommended Approach

For local development and evaluation runs, use the **stored OAuth credentials** approach (the default). This requires:

1. Install the Copilot CLI: follow [installation guide](https://docs.github.com/en/copilot/how-tos/set-up/install-copilot-cli)
2. Run `copilot` once to authenticate via GitHub OAuth device flow
3. The SDK automatically uses stored credentials — no additional configuration

For CI/CD or automation, set `COPILOT_GITHUB_TOKEN` or `GH_TOKEN` in the environment.

### 9.3 Implementation

```go
func newCopilotClient(debug bool) *copilot.Client {
    opts := &copilot.ClientOptions{
        LogLevel: "error",
    }
    if debug {
        opts.LogLevel = "debug"
    }
    // Auth is handled automatically:
    // 1. Check GitHubToken (empty, skip)
    // 2. Check env vars (COPILOT_GITHUB_TOKEN, GH_TOKEN, GITHUB_TOKEN)
    // 3. Use stored OAuth credentials from `copilot` CLI login
    // 4. Fall back to `gh auth` credentials
    return copilot.NewClient(opts)
}
```

### 9.4 Note on doc-review-agent Auth

The doc-review-agent uses `gh auth login` credentials (via the `GH_TOKEN` env var or the `gh` CLI's stored auth). Its SKILL.md lists `gh auth login` as a prerequisite. The SDK approach is equivalent but more flexible — it checks multiple token sources automatically. Users who already have `gh auth` set up need zero additional configuration.

---

## 10. Repo Structure

### 10.1 Monorepo Layout

The Go eval tool lives in `ronniegeraghty/azure-sdk-prompts` alongside the prompt library and reports. Everything ships from one repo. Python scripts have been removed — `hyoka` is the sole evaluation tool.

```
azure-sdk-prompts/                     # ronniegeraghty/azure-sdk-prompts
├── README.md
├── LICENSE
├── manifest.yaml                      # Auto-generated prompt index
├── configs/                           # Tool configuration matrix (repo root)
│   ├── all.yaml                       # Both configs (default for matrix runs)
│   ├── baseline.yaml                  # No MCP, no skills — raw Copilot
│   └── azure-mcp.yaml                # Azure MCP server attached
├── prompts/                           # Prompt library (79+ prompts)
│   └── storage/
│       └── data-plane/
│           └── dotnet/
│               └── authentication.prompt.md
├── hyoka/                              # Go eval tool (hyoka)
│   ├── README.md                      # CLI reference
│   ├── cmd/hyoka/main.go           # CLI entry point (cobra)
│   ├── go.mod / go.sum
│   ├── internal/                      # Go packages
│   │   ├── config/                    # Config file parsing
│   │   ├── prompt/                    # Load, filter, parse prompts
│   │   ├── eval/                      # Evaluation engine + workspace
│   │   ├── build/                     # Language-specific build verification
│   │   ├── report/                    # JSON report generation
│   │   ├── manifest/                  # Manifest generation from prompts
│   │   └── validate/                  # Prompt frontmatter validation
│   └── testdata/                      # Test fixtures
├── reports/                           # Evaluation reports
│   └── runs/<timestamp>/
├── docs/                              # Documentation
│   └── eval-tool-plan.md              # This plan
└── templates/
    └── prompt-template.prompt.md
```

### 10.2 Dependencies

```
hyoka/go.mod:
  module github.com/ronniegeraghty/azure-sdk-prompts/tool

  go 1.26.1

  require (
      github.com/spf13/cobra            v1.10.2  // CLI framework
      gopkg.in/yaml.v3                  v3.0.1   // YAML parsing
  )
```

### 10.3 Tool as Sole Evaluation Approach

Python scripts (`run-evals.py`, `generate-manifest.py`, `validate-prompts.py`) have been removed. The Go `hyoka` tool replaces all their functionality:

| Removed Python Script | Go Replacement |
|---|---|
| `scripts/run-evals.py` | `hyoka run` |
| `scripts/generate-manifest.py` | `hyoka manifest` |
| `scripts/validate-prompts.py` | `hyoka validate` |

The tool uses smart path detection — running from the repo root or the `hyoka/` directory both work without extra flags.

---

## 11. Error Handling and Timeout Management

### 11.1 Timeout Hierarchy

```
Global timeout (--timeout flag, default 300s)
  └── Per-prompt timeout (frontmatter `timeout` field, overrides global)
       ├── Code generation session: 80% of timeout
       ├── Build verification: 120s fixed
       ├── Test generation: remainder of timeout
       └── Code review session: 60s fixed
```

### 11.2 Context Cancellation

Use Go's `context.Context` for cancellation throughout:

```go
ctx, cancel := context.WithTimeout(parentCtx, timeout)
defer cancel()

response, err := session.SendAndWait(ctx, copilot.MessageOptions{
    Prompt: promptText,
})
```

The SDK's `SendAndWait` respects context deadlines and returns `context.DeadlineExceeded` on timeout.

### 11.3 Error Categories

| Error | Handling |
|---|---|
| SDK/CLI startup failure | Fatal — abort the entire run |
| Session creation failure | Skip this eval, log error, continue others |
| Prompt execution timeout | Record as timeout, capture partial results |
| Build failure | Expected — record result, continue to review |
| Review session failure | Record as review error, still report build results |
| MCP server startup failure | Skip this config, log warning |

### 11.4 Graceful Shutdown

Handle SIGINT/SIGTERM to:
1. Cancel all in-progress evaluations
2. Disconnect all sessions
3. Stop the Copilot client
4. Write partial results for completed evaluations
5. Generate partial report

---

## 12. Parallel Execution

### 12.1 Worker Pool

Use a semaphore-based worker pool. Each worker gets its own Copilot client to avoid session conflicts:

```go
type WorkerPool struct {
    workers   int
    semaphore chan struct{}
}

// Each eval task runs in its own goroutine with its own client
func (p *WorkerPool) RunEval(ctx context.Context, task EvalTask) EvalResult {
    p.semaphore <- struct{}{}
    defer func() { <-p.semaphore }()

    client := copilot.NewClient(opts)
    client.Start(ctx)
    defer client.Stop()

    // Run eval with this client...
}
```

### 12.2 Resource Considerations

- Each worker spawns a Copilot CLI process (stdio mode)
- Memory: ~50-100MB per CLI process
- Recommended: `--workers 4` on a standard dev machine
- The CLI handles API rate limiting internally

---

## 13. Implementation Phases

All work happens in the `ronniegeraghty/azure-sdk-prompts` repo.

### Phase 1: Foundation (MVP) ✅
- [x] Project scaffolding (`go.mod`, `cmd/hyoka/`, `internal/` packages, CLI framework)
- [x] Prompt loader (read and filter prompts from `./prompts`)
- [x] Basic evaluation engine (create session, send prompt, capture events)
- [x] Build verification (all languages)
- [x] JSON report generation to `./reports`
- [x] CLI with filter flags, `--workers`, `--dry-run`

### Phase 2: Quality Signals ✅
- [x] Code review via separate Copilot session (LLM-as-judge)
- [x] Reference answer comparison
- [x] HTML report generation with comparison dashboard
- [x] Cross-config comparison summary

### Phase 3: Tool Matrix ✅
- [x] Configuration matrix YAML parsing (`configs/` directory)
- [x] MCP server attachment per config
- [x] Skill loading per config
- [x] Tool filtering per config
- [x] Matrix execution (prompt × config cross-product)

### Phase 4: Evaluation Quality ✅
- [x] Rename CLI to `hyoka`
- [x] `check-env` command — tests if language toolchains are installed (dotnet, python, go, node, java, rust, cargo, cmake, etc.) and reports availability
- [x] `expected_tools` field in prompt frontmatter — reviewer checks if generation session used those tools
- [x] Reviewer build skill — a skill the reviewing Copilot session loads that knows how to set up build environments and attempt builds without modifying generated code
- [x] SDK version checking skill — skill for reviewer that checks if generated code uses latest Azure SDK package versions
- [x] Tool usage evaluation criteria — reviewer checks generation session tool calls against `expected_tools`
- [x] Historical trend command — `hyoka trends --prompt-id X` uses Copilot SDK to analyze past runs and generate trend reports

> **Status:** Complete (v0.5.0).

### Phase 5: Polish ✅
- [x] Review comments on generated code — new `code-review-comments` skill instructs reviewer to add inline `REVIEW:` comments. Annotated files saved to `reviewed-code/` alongside `generated-code/`. HTML report highlights review comments in amber; Markdown report shows annotated code in fenced blocks.
- [x] `hyoka report` command — re-render HTML/MD from existing report.json (no Copilot SDK, purely template-based). Supports `report <run-id>` and `report --all`.
- [x] Progress bars and color output for terminal UX — ANSI progress bar with pass/fail icons, duration, worker count. Auto-disabled in `--debug` mode to avoid conflict with verbose output.
- [x] Starter project support — `project_context: existing` + `starter_project:` in prompt frontmatter. Already implemented in evaluator; documented in prompt template.
- [ ] Embedded CLI binary with `auth login`/`auth status` commands — deferred. Use SDK bundler to embed Copilot CLI inside Go binary so users never need to install the Copilot CLI separately.

> **Status:** Complete (v0.2.0). Embedded CLI binary deferred to future work.

---

## 14. Open Questions

1. **Prompt schema migration:** The new eval-specific frontmatter fields (`project_context`, `reference_answer`, `timeout`, `expected_packages`) must be optional with sensible defaults so existing prompts work unchanged with the Go eval tool.

2. **Shared vs. separate Copilot client:** Should all workers share one client (one CLI process, multiple sessions) or each get their own? The SDK supports multiple sessions per client, but isolated clients per worker is simpler and avoids shared state. Start with isolated clients; optimize later if resource usage is a concern.

3. **MCP server lifecycle:** When using the same MCP server config across multiple evals, should the MCP server process be shared or restarted per eval? The SDK manages MCP processes per session, so they're naturally isolated. This is correct behavior for eval — each eval should start clean.

4. **Model selection for review:** Should the review model be the same as the generation model, or always use a specific "reviewer" model? Recommendation: default to the same model but allow override in the config file via a `review_model` field.

5. **Cost management:** Each eval consumes premium requests. With 57 prompts × 3 configs × (generation + review) = ~342 API calls per full run. Add `--budget N` flag to cap total evaluations?

6. ~~**Go module vs. existing repo tooling:**~~ Resolved — Python scripts removed. The repo is Go (`hyoka/`) + prompt files + configs. No polyglot complexity.

---

## 15. Key Risks

| Risk | Mitigation |
|---|---|
| Copilot SDK is in technical preview and may change | Pin SDK version in `go.mod`; wrap SDK calls in an adapter layer |
| API rate limits with parallel execution | Start with `--workers 4`; the SDK handles rate limiting internally |
| MCP server startup failures | Timeout + retry; skip config on repeated failure |
| Generated code has dependencies we can't install | Timeout on `dotnet restore` / `npm install`; record as build failure |
| Review scoring inconsistency | Use structured JSON output format; validate against schema |
| Large workspace directories from agent | Set workspace size limit; clean up after each eval |
| ~~Polyglot repo complexity~~ | Resolved — Python scripts removed; repo is Go + prompts only |

---

## Appendix A: Go Copilot SDK Quick Reference

```go
// Create client
client := copilot.NewClient(&copilot.ClientOptions{
    LogLevel: "error",
})
client.Start(ctx)
defer client.Stop()

// Create session with MCP and skills
session, _ := client.CreateSession(ctx, &copilot.SessionConfig{
    Model: "claude-sonnet-4.5",
    WorkingDirectory: "/tmp/eval-workspace",
    MCPServers: map[string]copilot.MCPServerConfig{
        "azure": {
            "type": "local",
            "command": "npx",
            "args": []string{"-y", "@azure/mcp@latest"},
            "tools": []string{"*"},
        },
    },
    SkillDirectories: []string{"./skills/azure-sdk"},
    OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
    SystemMessage: &copilot.SystemMessageConfig{
        Mode: "append",
        Content: "...",
    },
})
defer session.Disconnect()

// Send and wait
response, _ := session.SendAndWait(ctx, copilot.MessageOptions{
    Prompt: "Write a Go program that lists Azure Storage blobs...",
})
fmt.Println(*response.Data.Content)
```

## Appendix B: Prompt Frontmatter Example

```yaml
---
id: storage-dp-dotnet-auth
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: intermediate
description: >
  Write a .NET console app that authenticates to Azure Blob Storage
  using DefaultAzureCredential and lists all containers.
sdk_package: Azure.Storage.Blobs
tags: [identity, default-credential, blob-storage]
created: 2026-07-01
author: ronniegeraghty
project_context: blank
reference_answer: authentication.reference/
timeout: 300
expected_packages:
  - Azure.Storage.Blobs
  - Azure.Identity
---

# Authentication: Azure Blob Storage (.NET)

## Prompt

Write a .NET 8 console application that:
1. Authenticates to Azure Blob Storage using `DefaultAzureCredential`
2. Lists all containers in the storage account
3. Prints each container's name and last modified date
4. Handles authentication errors gracefully

The storage account URL should come from an environment variable `AZURE_STORAGE_URL`.

## Expected Coverage

- Azure.Identity package for DefaultAzureCredential
- Azure.Storage.Blobs package for BlobServiceClient
- Proper async/await usage
- Environment variable configuration
- Error handling with try/catch
```
