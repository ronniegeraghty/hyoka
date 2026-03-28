# Hyoka — Project Architecture

## Overview

Hyoka is a Go CLI tool that evaluates AI agents generating Azure SDK code. It sends prompts through the Copilot SDK, collects generated code, optionally builds it, then runs a multi-model review panel to score the output.

## Directory Layout

```
hyoka/                         # Go module (github.com/ronniegeraghty/hyoka)
├── main.go                    # CLI entry point (cobra). Parses flags, wires pipeline.
└── internal/
    ├── config/config.go       # Loads YAML config files (models, skills, MCP servers)
    ├── prompt/                # Loads prompt .md files, parses YAML frontmatter
    │   ├── loader.go          #   Discovers prompts from prompts/ directory tree
    │   ├── parser.go          #   Parses frontmatter + body from markdown
    │   └── types.go           #   Prompt struct definition
    ├── skills/fetcher.go      # Fetches remote skills from GitHub before eval starts
    │
    ├── eval/                  # CORE — orchestrates the entire eval pipeline
    │   ├── engine.go          #   Runs: generate → build → review → report
    │   ├── copilot.go         #   Talks to Copilot SDK — sends prompts, handles events
    │   ├── workspace.go       #   Temp dirs, file recovery, cleanup
    │   └── proctracker.go     #   Tracks child processes, kills orphans
    │
    ├── build/                 # Runs real build commands per language
    │   ├── verifier.go        #   Executes build (go build, dotnet build, mvn, npm, etc.)
    │   └── languages.go       #   Language → build command mapping
    │
    ├── review/                # Multi-model code review panel
    │   ├── reviewer.go        #   Creates reviewer sessions, runs panel, consolidates scores
    │   ├── rubric.go          #   Scoring rubric template injected into review prompts
    │   └── types.go           #   ReviewResult, CriterionResult structs
    │
    ├── report/                # Writes evaluation results to disk
    │   ├── generator.go       #   Writes JSON reports
    │   ├── html.go            #   Generates HTML dashboard
    │   ├── markdown.go        #   Generates Markdown summary
    │   ├── summary_stats.go   #   Aggregate statistics
    │   └── types.go           #   EvalReport struct
    │
    ├── logging/logging.go     # slog setup, EvalLogger helper, CLI flag integration
    ├── progress/              # Live terminal display (TUI during eval runs)
    ├── trends/                # Cross-run trend analysis
    ├── history/               # Historical run tracking
    ├── validate/              # Prompt frontmatter validation
    ├── rerender/              # Re-render reports from existing JSON
    ├── manifest/              # Prompt manifest generation
    ├── checkenv/              # Environment prerequisite checks
    └── utils/                 # Shared helpers

configs/                       # Evaluation config YAML files
prompts/                       # Prompt library (organized by language/service)
skills/                        # Copilot skills (generator/ and reviewer/)
reports/                       # Generated output (gitignored)
docs/                          # Documentation
```

## Eval Pipeline

A single evaluation flows through these stages:

```
 ┌─────────────┐
 │ Prompt + Config │
 └──────┬──────┘
        ▼
 ┌─────────────┐     Copilot SDK session generates code
 │  eval/copilot │──→ files land in isolated workspace
 └──────┬──────┘
        ▼
 ┌─────────────┐     Optional: runs real build commands
 │  build/       │──→ go build, dotnet build, npm install, etc.
 └──────┬──────┘     (enabled with --verify-build)
        ▼
 ┌─────────────┐     Multiple LLMs score code against rubric
 │  review/      │──→ each model reviews independently,
 └──────┬──────┘     then scores are consolidated
        ▼
 ┌─────────────┐
 │  report/      │──→ JSON + HTML + Markdown output
 └─────────────┘
```

## Key Concepts

### Configs
YAML files in `configs/` define an evaluation environment: which model generates code, which models review it, what skills and MCP servers to attach. Multiple configs can run against the same prompts for comparison.

### Prompts
Markdown files in `prompts/` organized by `{language}/{service}/`. Each has YAML frontmatter (`id`, `service`, `language`, `plane`, `category`, `difficulty`) and a body containing the task description plus optional evaluation criteria.

### Multi-Model Review Panel
The review phase sends generated code to multiple LLMs (e.g., Claude, Gemini, GPT) independently. Each scores against a rubric of criteria (pass/fail per criterion). A consolidator model merges the individual reviews into a final consensus score.

### Skills
Copilot skills (SKILL.md files) provide domain knowledge to the generator and reviewer sessions. Skills can be local (in `skills/`) or remote (fetched from GitHub repos before eval starts).

### Guardrails
- **Turn limit**: 25 assistant turns per generation (prevents runaway sessions)
- **File limit**: 50 generated files max
- **Output size limit**: 1 MB total
- **Process tracking**: Child processes are registered and killed on timeout/abort
- **Workspace isolation**: Generated code lands in temp dirs, misplaced files are recovered

## CLI

```bash
hyoka run           # Run evaluations (main command)
hyoka list          # List available prompts
hyoka validate      # Validate prompt frontmatter
hyoka check-env     # Verify prerequisites
hyoka trends        # Analyze trends across runs
hyoka report        # Re-generate reports from JSON
hyoka version       # Print version
```

Key flags for `hyoka run`:
- `--prompt-id` — Run a single prompt
- `--language` / `--service` — Filter prompts
- `--config-file` — Use a specific config
- `--log-level debug` — Structured debug logging
- `--log-file path` — Redirect logs to file
- `--verify-build` — Run real build verification
- `--skip-review` — Skip the review phase
