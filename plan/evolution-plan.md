# Hyoka Evolution Plan

**Version:** 1.1  
**Date:** 2026-04-04  
**Author:** Morpheus (Lead/Architect)  
**Status:** Approved by Ronnie Geraghty  
**Updated:** Design meeting decisions (D-AUTO-DM1–DM20) incorporated

---

## Issue Tracking

**All 72 evolution plan tasks are tracked as GitHub issues #91–#162.**

Phase 0 (Foundation): #91–#99 + new tasks 0.10 (9+ issues)  
Phase 1 (Core Model): #100–#119 + new tasks 1.6d, 1.8 (20+ issues)  
Phase 2 (Evaluation Engine): #120–#137 (18 issues)  
Phase 3 (Transparency): #138–#143 (6 issues)  
Phase 4 (Insights & Comparison): #144–#151 (8 issues)  
Phase 5 (Ecosystem): #152–#162 (11 issues)  

---

## Vision

Transform hyoka from an Azure SDK–focused evaluation tool into a **general-purpose AI agent benchmarking platform**. Any team can bring their own prompts, criteria, tools, and system prompts to measure how different tool configurations affect agent code generation quality.

The core question hyoka answers: **Which tools help agents write better code, and which ones hurt?**

---

## Current State Assessment

**Codebase:** ~20K lines Go across 22 packages + 1329-line `main.go` monolith. 264 test functions. 87 prompts. 8 configs. Go 1.26.1.

**Architecture:** Clean layered pipeline — `eval → review → report`. Good use of interfaces (`CopilotEvaluator`, `Reviewer`). Dependency graph is acyclic. Config normalization handles legacy→new format migration. 3 direct dependencies (Copilot SDK, Cobra, YAML).

**What works well:**
- Error handling is excellent — `%w` wrapping, no log-and-return, no panics
- Safety boundaries properly implemented (`--allow-cloud` opt-in)
- Build passes clean, go vet clean, all 264 tests pass
- Dependency footprint is minimal

**What needs fixing:**
- **No CI pipeline** — any PR can merge broken code (P0)
- **Reviewer model bug** — multi-config evals silently use wrong reviewer panel (`main.go:469-473`) (P0)
- **main.go monolith** — 1329 lines, all commands in one file (P1)
- **Hardcoded prompt fields** — `Prompt` struct has 12+ fixed fields, blocks non-Azure use (architectural)
- **Biased system prompt** — 15 rules injected into every agent session, many are behavioral bias not operational necessity (architectural)
- **No pairwise testing** — can't measure individual tool impact (feature gap)
- **No comparison engine** — trends exist but no side-by-side config comparison (feature gap)
- **pidfile untested** — only package with zero tests (P1)

---

## Phase 0: Foundation

**Goal:** Establish CI safety net and fix critical bugs so subsequent phases don't regress.

**Duration:** 1 sprint

| Task | Description | Owner | Size | Depends On |
|------|-------------|-------|------|------------|
| 0.1 | **Create CI pipeline** (#91) — `go build ./hyoka/...`, `go vet ./hyoka/...`, `go test -race ./hyoka/... -timeout 2m` on all PRs and pushes. Go 1.26.1, ubuntu-latest. Add `-race` from day one (concurrent code in ResourceMonitor, ProcessTracker, PanelReviewer). Skip golangci-lint until Phase 1. (DM8) | Tank | Medium | — |
| 0.2 | **Fix reviewer model bug** (#92) — `main.go:469-473` grabs reviewers from first config only. Fix: create reviewer panel per-task via factory function, not engine-scoped. Each config must use its own reviewer panel. (DM1) | Neo | Small | — |
| 0.3 | **Fix stale path in new-prompt** (#93) — `main.go:1276` says `go run ./tool/cmd/hyoka validate`, should be `go run ./hyoka validate`. | Tank | Small | — |
| 0.4 | **Add generator model validation** (#94) — Empty `Generator.Model` passes validation but fails at runtime (`config.go:256-287`). | Tank | Small | — |
| 0.5 | **Detect duplicate config names** (#95) — Two configs with the same name silently shadow. Second config becomes inaccessible. | Tank | Small | — |
| 0.6 | **Big-bang migrate config files** (#96) — Migrate all 8 config files to `Generator`/`Reviewer` sub-struct format. Delete all legacy fields, `Normalize()`, and 7 `Effective*()` getters (~130 lines / 35% of config.go). Direct field access replaces `tc.EffectiveModel()` → `tc.Generator.Model`. Also update ~17 call sites including `main.go:394` and `Validate()`. (DM15: independent of 0.2, can proceed in parallel) | Tank | Medium | — |
| 0.7 | **Log discarded errors** (#97) — `reviewer.go:352`, `copilot.go:83`, `main.go:219,263,270,286`, `fetcher.go:68,82,88,120`. | Neo | Small | — |
| 0.8 | **Fix Go version in docs** (#98) — 9+ files say Go 1.24.5+, go.mod requires 1.26.1. Includes `docs/getting-started.md`, `docs/contributing.md`, `README.md`, `AGENTS.md`, and internal docs. | Oracle | Small | — |
| 0.9 | **Fix flaky resourcemonitor tests** (#99) — Remove `time.Sleep` assertions; call `sample()` directly with event-driven checks. | Switch | Small | — |
| 0.10 | **Add pidfile package tests** — 136 lines, zero tests, safety-critical code. Write/Remove/ReadAlive, cross-platform alive check, stale cleanup. (DM9: moved from Phase 5 task 5.3a) | Switch | Small | — |

---

## Phase 1: Core Model Changes

**Goal:** Generalize hyoka's data model so it's no longer Azure-specific. This is the architectural foundation every subsequent phase depends on.

**Duration:** 2-3 sprints

### 1.1 — Generic Prompt Properties (Properties as Sole Representation)

**Current:** `Prompt` struct (`prompt/types.go:4-32`) has 12+ hardcoded fields: `Service`, `Plane`, `Language`, `Category`, `Difficulty`, `SDKPackage`, `DocURL`, etc.

**Target:** `Properties map[string]string` is the metadata representation. Drop typed struct fields for metadata (`Service`, `Language`, `Category`, `Difficulty`, `SDKPackage`, `DocURL`, etc.) and move them into the properties map. **Non-string fields stay typed:** `Tags []string`, `Timeout int`, `StarterProject` remain as typed struct fields — they cannot be losslessly represented as `map[string]string`. (DM2) Add convenience getter methods that read from the map: `p.Language()` returns `p.Properties["language"]`. The `Filter` struct also becomes property-based: `Filters map[string]string` instead of typed fields.

**Property key convention (DM13):** All property keys use `snake_case`. Validation rejects hyphens and camelCase. Migration script normalizes existing keys.

Prompt frontmatter becomes:
```yaml
id: key-vault-dp-python-crud
properties:
  service: key-vault
  plane: data_plane
  language: python
  category: crud
  difficulty: medium
  sdk_package: azure-keyvault-secrets
```

**Migration strategy:** Big-bang. Update all 87 prompts to nested `properties:` format. No backward compatibility. Write a migration script that reads existing frontmatter and converts to the new shape. Also update the `new-prompt` command scaffold template to emit the new `properties:` format. (DM20)

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.1a | Redesign `Prompt` struct: replace typed fields with `Properties map[string]string`. Add convenience getter methods (`Language()`, `Service()`, etc.). Update parser for nested `properties:` key. (#100) | Neo | Medium |
| 1.1b | Write migration script to convert all 87 prompts from flat typed fields to nested `properties:` format. (#101) | Tank | Medium |
| 1.1c | Update all filter flags (`--service`, `--language`, `--plane`, `--category`) to query properties map. Redesign `Filter` struct as `map[string]string`. Support generic `--filter key=value` alongside legacy aliases. (#102) | Tank | Medium |

### 1.2 — Criteria → Grader Configs

**Current:** `MatchCondition` struct (`criteria/criteria.go:28-34`) has 5 hardcoded fields: `Language`, `Service`, `Plane`, `Category`, `SDK`. Criteria are text blobs merged via `MergeCriteria()` and injected into a single LLM review prompt. Tier 1/2/3 system.

**Target:** Criteria become **grader configurations** — no tier system, no `MergeCriteria()`, no `FormatCriteria()`. Each grader is defined in YAML with typed fields:

```yaml
graders:
  - kind: file
    name: "main_file_exists"
    config:
      path: "main.py"
    weight: 1.0
    gate: true          # Hard pass/fail — score is irrelevant if this fails (DM3)

  - kind: program
    name: "builds_successfully"
    config:
      command: "python -m py_compile main.py"
    weight: 1.0
    gate: true          # "Doesn't compile" can't be hidden by high LLM scores (DM3)

  - kind: prompt
    name: "code_quality"
    config:
      model: "claude-opus-4.6"
      rubric: "Evaluate the code for correctness, style, and SDK usage..."
    weight: 0.5
    when:
      language: python

  - kind: behavior
    name: "used_mcp_tools"
    config:
      required_tools: ["azure-mcp"]
    when:
      service: key-vault
```

Grader applicability uses property-based `when:` conditions instead of `MatchCondition`. Graders without `when:` always apply. The system collects all applicable graders into a list — no merging logic needed. Criteria YAML files become grader config files. Graders with `gate: true` are hard pass/fail gates — if they fail, the overall result fails regardless of weighted scoring. (DM3)

**Decision (Q1, anchoring review):** Remove Tier 1 entirely. Replace tier system with composable grader configs. Absorb FR-14 and FR-17 into grader architecture.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.2a | Design grader config YAML schema (`kind:`, `name:`, `config:`, `when:`, `weight:`). Define initial grader types: `file`, `program`, `prompt`, `behavior`, `action_sequence`, `tool_constraint`. (#103) | Morpheus | Medium |
| 1.2b | Replace `MatchCondition` with `when: map[string]string` property matching on grader configs. Delete `MergeCriteria()` and `FormatCriteria()`. (#104) | Neo | Medium |
| 1.2c | Migrate existing criteria YAML files to grader config format. Remove Tier 1 system. (#105) | Tank | Medium |

### 1.3 — Tool Filters (Property-Based)

**Current:** Config YAML supports `available_tools` and `excluded_tools` as flat string lists.

**Target:** Tools can be filtered based on prompt properties. A config can say "for Python prompts, enable tool X; for Go prompts, disable tool Y." This enables property-driven tool availability — the foundation for pairwise testing.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.3a | Design tool filter schema that references prompt properties. (#106) | Morpheus | Small |
| 1.3b | Implement property-based tool filter resolution in config loading. (#107) | Tank | Medium |
| 1.3c | Wire tool filters into session config builder (`copilot.go:617-693`). (#108) | Neo | Medium |

### 1.4 — YAML Prompt Format

**Current:** Prompts are Markdown files with YAML frontmatter. The prompt text is the Markdown body.

**Target:** Support pure YAML prompt format alongside Markdown. YAML prompts are more structured and easier to generate/validate programmatically.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.4a | Add `ParsePromptYAML()` function alongside existing Markdown parser. (#109) | Neo | Medium |
| 1.4b | Update prompt loader to auto-detect format by file extension (`.prompt.md` vs `.prompt.yaml`). (#110) | Neo | Small |

### 1.5 — Split main.go

**Current:** 1329 lines in `hyoka/main.go`. All 14 cobra commands, flag definitions, path resolution, reviewer wiring, skill installation, and prompt scaffolding in one function.

**Target:** Create `hyoka/cmd/` package with per-command files: `run.go`, `list.go`, `serve.go`, `clean.go`, `validate.go`, `check_env.go`, `new_prompt.go`, `trends.go`, `rerender.go`.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.5a | Create `hyoka/cmd/` package structure with root command and shared state. (#111) | Tank | Large |
| 1.5b | Move each command to its own file with RunE function. (#112) | Tank | Large |
| 1.5c | Update main.go to be a thin entry point that registers commands from cmd/. (#113) | Tank | Small |

### 1.6 — Configurable System Prompts

**Current:** System prompt is hardcoded in `copilot.go:628-655` with 15 rules. All agent sessions get the same prompt.

**Target:** 
- **Default:** Zero system prompt. No rules injected. SDK session config handles working directory, tools, and isolation.
- **Config-level override:** Config YAML has optional `generator.system_prompt` and `reviewer.system_prompt` fields. Users who want specific behavior can set custom system prompts.
- **Both gen and review agents** support configurable system prompts.

**Decision (Q2, Q6):** Follow Waza's approach. Zero system prompt by default. All isolation/directory config handled through SDK `SessionConfig`. Config-specific custom system prompts available for users who want them. Response type (files vs text) is a system-prompt-level concern.

**Removal strategy (DM6):** System prompt removal is phased, not big-bang:
1. Remove bias rules (rules 9-10: research restrictions, Python rules)
2. Remove guidance rules (rules 1-2)
3. Remove path/operational rules (rules 3-7, 8) — after confirming SDK config handles isolation
4. Remove safety boundaries (rules 11-15) — move to code-level hooks first

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.6a | Add `system_prompt` field to both generator and reviewer config sections. (#114) | Tank | Small |
| 1.6b | Modify `buildSessionConfig()` to use config system prompt (or empty if not set) instead of hardcoded rules. (#115) | Neo | Medium |
| 1.6c | Remove the 15 hardcoded generator rules incrementally per DM6 phased strategy. Move any truly operational rules (file path enforcement) to SDK hooks or pre/post validation. (#116) | Neo | Medium |
| 1.6d | **Remove hardcoded reviewer system message** (`reviewer.go:180-183`). Make reviewer system prompt configurable via config YAML `reviewer.system_prompt` field. (DM7) | Neo | Small |

### 1.7 — Starter Files

**Current:** `StarterProject` field exists on `Prompt` struct but implementation is minimal. `copilot.go:83` silently discards errors when handling starter files.

**Target:** Full starter file support. Prompts can reference a directory of files that get copied into the agent's working directory before the session begins. This enables "fix this broken code" style prompts.

**Decision:** Core feature. Follow Waza's `ResourceFile` pattern — files placed in workspace before session starts.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.7a | Design starter file reference format in prompt frontmatter (directory path or file list). (#117) | Morpheus | Small |
| 1.7b | Implement file copying into workspace before session creation. Fix silent error handling at `copilot.go:83`. (#118) | Neo | Medium |
| 1.7c | Add starter file validation to prompt validation pipeline. (#119) | Switch | Small |

### 1.8 — Review Package Coverage (DM10: moved from Phase 5)

**Current:** Review package has 5 tests for 840 lines — critically undertested before Phase 2 replaces it.

**Target:** Increase review package test coverage. Must complete before Phase 2 replaces `review/` with `graders/`. Tests verify the `prompt` grader faithfully wraps old reviewer behavior.

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.8 | Increase review package test coverage (5 tests for 840 lines). Must complete before Phase 2 grader work. (DM10: moved from Phase 5 task 5.3c) | Switch | Medium |

---

## Phase 2: Evaluation Engine

**Goal:** Build advanced evaluation capabilities — pairwise testing, session limits, isolation, and resource efficiency.

**Duration:** 2-3 sprints  
**Depends on:** Phase 1 (generic properties, tool filters, system prompt changes)

### 2.1 — Pairwise Testing

**Decision (Q3):** `--pairwise` / `-pw` flag on the `run` command. When passed, expands one config into pairwise variants (each tool toggled on/off). Config YAML supports `always_on: true` per tool to exempt from toggling.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.1a | Create `pairwise/` package that generates N+1 config variants from a base config (one per tool toggle + baseline). (#120) | Neo | Large |
| 2.1b | Add `--pairwise` flag to run command. Wire into eval engine to expand configs before execution. (#121) | Tank | Medium |
| 2.1c | Add `always_on` field to tool config. Exempt marked tools from pairwise toggling. (#122) | Tank | Small |
| 2.1d | Generate pairwise comparison report showing per-tool impact scores. (#123) | Trinity | Medium |

### 2.2 — User-Configurable Session Limits

**Current:** Hardcoded guardrails: 25 turns, 50 files, 1 MB output, 50 max session actions.

**Target:** All limits configurable in config YAML with sensible defaults. Users can tune for their use case.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.2a | Add session limit fields to config schema (max_turns, max_files, max_output_size, max_session_actions). (#124) | Tank | Small |
| 2.2b | Wire config limits into eval engine guardrail checks. (#125) | Neo | Medium |

### 2.3 — Isolated Evaluation Environment

**Current:** Agent sessions run in the user's environment. Leakage possible.

**Target:** Each evaluation session runs in a clean, isolated workspace. No leakage from user's dev environment into agent sessions.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.3a | Create isolated workspace directory per session with clean environment. (#126) | Neo | Medium |
| 2.3b | Copy only declared starter files and config-specified resources into workspace. (#127) | Neo | Small |
| 2.3c | Clean up workspace after session completes (integrate with existing clean command). (#128) | Neo | Small |

### 2.4 — Resource Efficiency

**Current:** Some concerns about memory usage with large eval runs. No resource monitoring by default.

**Target:** Bounded memory usage, proper goroutine cleanup, no resource leaks. `--monitor-resources` flag already exists — ensure it catches real issues.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.4a | Audit goroutine lifecycle across eval/review pipeline. Ensure all goroutines terminate on context cancellation. (#129) | Neo | Medium |
| 2.4b | Add memory bounds for in-flight report data. Stream large reports to disk. (#130) | Trinity | Medium |

### 2.5 — Grader Architecture

**Current:** Review system uses a monolithic `Reviewer` interface — one Copilot session reads all files, scores all criteria, returns one JSON blob. `PanelReviewer` runs multiple models doing the same monolithic review and consolidates. Everything goes through `BuildReviewPrompt()` → LLM → parse JSON.

**Target:** Replace `Reviewer`/`PanelReviewer` with a pluggable `Grader` interface and typed implementations. Create new `hyoka/internal/graders/` package. The current LLM review becomes one grader type (`prompt`) among many:

```go
type Grader interface {
    Kind() string
    Name() string
    Grade(ctx context.Context, input GraderInput) (GraderResult, error)
}

// GraderInput is a concrete struct, not an interface (DM5).
// Contains everything a grader might need; graders use what they need and ignore the rest.
type GraderInput struct {
    WorkspacePath string
    ActionLog     []ActionEvent
    PromptMeta    PromptMetadata
    Config        GraderConfig
    Files         []FileInfo
}

// GraderResult uses typed optional fields instead of interface{} (DM4).
// Templates check `if .FileDetails` directly — no type assertions needed.
type GraderResult struct {
    Kind    string   // "file", "program", "prompt", "behavior", etc.
    Name    string   // Human-readable name
    Score   float64  // 0.0–1.0 normalized score
    Weight  float64  // Weight for aggregation
    Pass    bool     // Binary pass/fail
    Gate    bool     // If true, failure overrides weighted scoring (DM3)
    // Typed optional fields per grader kind (DM4)
    FileDetails     *FileGraderDetails
    ProgramDetails  *ProgramGraderDetails
    PromptDetails   *PromptGraderDetails
    BehaviorDetails *BehaviorGraderDetails
}
```

**GraderInput/GraderResult types must be frozen before grader implementation begins (DM14).** The `file` grader (2.5b) serves as type validation — simplest grader, exposes any design issues before complex graders are built.

**Grader types (initial set):**
- `file` — Check file existence, content patterns, structure
- `program` — Run a command (linter, compiler, test suite), pass/fail on exit code
- `prompt` — LLM-as-judge (current reviewer behavior), one rubric and **one model** per grader instance. Multi-model review is achieved by composing multiple `prompt` grader instances with different models and weights. (DM19)
- `behavior` — Check agent behavior from action log (tool usage, turn count, etc.)
- `action_sequence` — Verify agent followed expected action patterns
- `tool_constraint` — Verify tool usage constraints (required/forbidden tools)

**Architectural impact:**
- `hyoka/internal/review/` → phased replacement by `hyoka/internal/graders/`
- `criteria/` → grader config files with `when:` conditions (see §1.2)
- `report/types.go` → `ReviewResult` evolves to `[]GraderResult`
- FR-05 (Transparent Review Panel) absorbed — grader results ARE transparent by design
- FR-17 (Reviewer Tools) absorbed — `program` grader IS a tool
- FR-14 (Criteria Filters) absorbed — grader `when:` conditions

**Decision (anchoring review):** Adopt Waza's grader model. This is the single highest-impact architectural change.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.5a | Define `Grader` interface and `GraderResult` type in `hyoka/internal/graders/`. (#131) | Morpheus | Medium |
| 2.5b | Implement `file` grader (file existence, content pattern checks). (#132) | Neo | Medium |
| 2.5c | Implement `program` grader (run command, pass/fail on exit code, capture output). (#133) | Neo | Medium |
| 2.5d | Implement `prompt` grader (wrap current LLM reviewer as a grader — one rubric per grader instance, structured output). (#134) | Neo | Large |
| 2.5e | Implement `behavior`, `action_sequence`, and `tool_constraint` graders. (#135) | Neo | Medium |
| 2.5f | Wire grader execution into eval engine — collect applicable graders, run all, aggregate results. (#136) | Neo | Large |
| 2.5g | Update report types — `EvalReport` centers on `[]GraderResult` instead of monolithic `ReviewResult`. Add `SchemaVersion int` field for future-proofing. Migrate existing reports in-place to v2 schema (ESC-4: no dual-format support). (#137) | Trinity | Medium |

---

## Phase 3: Transparency

**Goal:** Make every agent action visible. Nothing hidden from the evaluator.

**Duration:** 1-2 sprints  
**Depends on:** Phase 1 (system prompt changes), Phase 0 (bug fixes)

### 3.0 — Template Extraction (DM12: prerequisite for Phase 3 display work)

**Current:** `html.go` is 1374 lines of string concatenation for HTML reports. Adding per-grader display components would push it past 1600 lines.

**Target:** Extract HTML templates from string concatenation into `.gohtml` files using `embed.FS`. This is a prerequisite for all Phase 3 template work.

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 3.0 | Extract HTML templates from html.go string concatenation into `.gohtml` files using `embed.FS`. (DM12) | Trinity | Medium |

### 3.1 — Full Agent Action History

**Current:** Reports show final output and scores. Individual agent actions (tool calls, file reads, bash commands) are logged but not structured for display.

**Target:** Every action the agent takes during a session is captured in a structured timeline. Reports show the full action sequence, not just the result.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 3.1a | Define action event schema (timestamp, type, tool, input, output, duration). (#138) | Morpheus | Small |
| 3.1b | Capture all agent events from SDK session hooks into structured action log. (#139) | Neo | Medium |
| 3.1c | Include full action timeline in JSON report output. (#140) | Trinity | Medium |
| 3.1d | Render action timeline in HTML reports (expandable, searchable). (#141) | Trinity | Large |

### 3.2 — Grader Result Transparency

**Current:** Review panel submits scores. Individual reviewer reasoning is captured but not prominently displayed.

**Target:** With the grader architecture (§2.5), transparency is inherent — each grader produces typed, structured output. A `program` grader result includes the command run, exit code, and stdout/stderr. A `prompt` grader result includes the model name, rubric, and full LLM reasoning. A `file` grader result includes which files were checked and their status. No special "transparency layer" needed — the grader results ARE the transparent view.

**What changes from the original plan:**
- Per-reviewer reasoning display → replaced by per-grader result display (more granular)
- Consolidation algorithm display → replaced by weighted aggregation of grader scores (simpler, explicit)
- Reviewer tool environments (FR-17) → absorbed into `program` grader type

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 3.2a | Render per-grader results in HTML reports — each grader type gets a typed display component (build output for `program`, LLM reasoning for `prompt`, check results for `file`). Depends on 3.0 template extraction. (#142) | Trinity | Large |
| 3.2b | Show weighted score aggregation formula and per-grader contributions in report. (#143) | Trinity | Small |

---

## Phase 4: Insights & Comparison

**Goal:** Make hyoka's primary output actionable data comparison, not just pass/fail grades.

**Duration:** 2 sprints  
**Depends on:** Phase 2 (pairwise testing, grader architecture), Phase 3 (action history)

### 4.1 — Comparison Engine

**Current:** Trends package (`trends/trends.go`, 857 lines) does cross-run analysis but it's static and batch-oriented.

**Target:** Dynamic comparison engine that can diff any two configs, runs, or time periods. Works on `[]GraderResult` — comparison is per-grader, not monolithic. Answer questions like "How did adding the Azure MCP server change the `builds_successfully` grader score for Python prompts?"

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 4.1a | Create `comparison/` package with config-vs-config, run-vs-run, and temporal diff logic. (#144) | Neo | Large |
| 4.1b | Add `hyoka compare` command for CLI-based comparison. (#145) | Tank | Medium |
| 4.1c | Expose comparison data via serve API for the dashboard. (#146) | Trinity | Medium |

### 4.2 — Serve Site Evolution

**Current:** `serve` command hosts static HTML reports and a basic dashboard.

**Target:** Interactive dashboard with:
- Config comparison views
- Pairwise impact visualizations
- Trend charts over time
- Filter by properties (language, service, etc.)
- Drill-down from summary to individual action timeline

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 4.2a | Design dashboard API endpoints for comparison, trends, and drill-down. (#147) | Trinity | Medium |
| 4.2b | Implement dynamic comparison UI with filter controls. (#148) | Trinity | Large |
| 4.2c | Add pairwise impact visualization (tool heatmap, contribution charts). (#149) | Trinity | Large |

### 4.3 — Enhanced Trends

**Current:** Trends show score changes over time with basic aggregation.

**Target:** Property-aware trend analysis. Slice trends by any prompt property combination. Detect regressions automatically.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 4.3a | Add property-based trend slicing (trend by language, by service, by tool config). (#150) | Neo | Medium |
| 4.3b | Add automatic regression detection with configurable thresholds. (#151) | Neo | Medium |

---

## Phase 5: Ecosystem

**Goal:** Make hyoka portable, extensible, and team-friendly.

**Duration:** 2-3 sprints  
**Depends on:** Phase 1 (generic properties, YAML format)

### 5.1 — `.hyoka` Project Directory

**Decision (Q5):** Project-scoped only. No global install. Structured subdirectories.

**Structure:**
```
.hyoka/
  configs/        # Team's evaluation configs
  prompts/        # Team's prompts
  criteria/       # Team's criteria
  skills/         # Team's skills
  reports/        # Output (gitignored)
```

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 5.1a | Add `hyoka init` command that scaffolds `.hyoka/` directory with subdirs and gitignore. (#152) | Tank | Medium |
| 5.1b | Auto-discover `.hyoka/` in current directory and ancestors. Walk-up stops at Git repository root (detected by `.git/` directory) — does not escape repo boundary. (DM16) Merge with CLI flags. (#153) | Tank | Medium |
| 5.1c | Update all path resolution to check `.hyoka/` first, then fall back to repo root paths. (#154) | Tank | Medium |

### 5.2 — Tool Marketplace / Repos

**Target:** Curated tool configurations (MCP servers, skills) that teams can reference by name instead of full config blocks.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 5.2a | Design tool registry format (YAML catalog of MCP server configs, skills). (#155) | Morpheus | Medium |
| 5.2b | Implement `hyoka tools list` and `hyoka tools add` commands. (#156) | Tank | Medium |
| 5.2c | Support remote tool registries (GitHub repos as tool catalogs). (#157) | Neo | Medium |

### 5.3 — Test Infrastructure

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| ~~5.3a~~ | ~~Add pidfile tests~~ — **Moved to Phase 0 task 0.10 (DM9)** | ~~Switch~~ | ~~Medium~~ |
| 5.3b | Add stub-based integration test (StubEvaluator + StubReviewer → engine.Run() → verify report). (#159) | Switch | Medium |
| ~~5.3c~~ | ~~Increase review package coverage~~ — **Moved to Phase 1 task 1.8 (DM10)** | ~~Switch~~ | ~~Medium~~ |
| 5.3d | Add serve handler tests (path traversal for `runID`, functional endpoint tests). (#161) | Switch | Medium |

### 5.4 — Skills

**Target:** Create 14 project-specific skills to encode hyoka's architecture, patterns, and domain knowledge. Skills are advisory, not prescriptive — guardrails not cages.

**Skill categories:**
- Core Architecture (5): eval-pipeline, error-handling, config-system, copilot-sdk-integration, criteria-system
- Working Patterns (4): testing-patterns, cli-patterns, report-generation, logging-conventions
- Human Developer (2): contributor-guide, prompt-conventions
- Evolution Support (3): property-migration, process-lifecycle, serve-patterns

**Task:** (#162) Implement all 14 skills

---

## Dependency Graph

```
Phase 0 (Foundation)
  ├── 0.1 CI Pipeline ─────────────────────────┐
  ├── 0.2 Reviewer model bug fix (DM1)         │
  ├── 0.3 Fix stale path                       │
  ├── 0.4 Generator model validation            │
  ├── 0.5 Duplicate config detection            │
  ├── 0.6 Config file migration (delete legacy) │  (DM15: independent of 0.2)
  ├── 0.7 Log discarded errors                  │
  ├── 0.8 Fix Go version in docs               │
  ├── 0.9 Fix flaky tests                      │
  └── 0.10 Pidfile tests (DM9: from Phase 5)   │
                                                │
Phase 1 (Core Model) ──── requires 0.1 ────────┘  (DM17: hard gate)
  ├── 1.1 Generic Properties (metadata-only) ──┐  (DM2: typed fields stay for non-string)
  │     └── 1.2 Grader Configs ────────────────┤  (requires 1.1) (DM3: gate field)
  │     └── 1.3 Tool Filters ─────────────────┤  (requires 1.1)
  ├── 1.4 YAML Prompt Format                    │
  ├── 1.5 Split main.go                        │
  ├── 1.6 Configurable System Prompts (DM6-7)  │  (includes 1.6d reviewer msg)
  ├── 1.7 Starter Files                        │
  └── 1.8 Review Package Coverage (DM10)       │  (must complete before Phase 2)
                                                │
Phase 2 (Eval Engine) ── requires 1.1-1.3 ─────┘
  ├── 2.1 Pairwise Testing ──── requires 1.3 Tool Filters
  ├── 2.2 Session Limits
  ├── 2.3 Isolated Environment
  ├── 2.4 Resource Efficiency
  └── 2.5 Grader Architecture ── requires 1.2   (DM4/5/14: freeze types first, file grader validates)
                                                
Phase 3 (Transparency) ── requires 1.6, 2.5
  ├── 3.0 Template Extraction (DM12) ── prerequisite for 3.2a
  ├── 3.1 Full Action History
  └── 3.2 Grader Result Transparency ── requires 2.5, 3.0
                                                
Phase 4 (Insights) ── requires 2.1, 2.5, 3.1
  ├── 4.1 Comparison Engine (on []GraderResult)
  ├── 4.2 Serve Site Evolution ── requires 4.1 (React SPA)
  └── 4.3 Enhanced Trends
                                                
Phase 5 (Ecosystem) ── requires 1.1, 1.4
  ├── 5.1 .hyoka Directory (DM16: walk-up stops at Git root)
  ├── 5.2 Tool Marketplace
  ├── 5.3 Test Infrastructure (pidfile/review moved earlier)
  └── 5.4 Skills (can start anytime)
```

**Critical path:** CI (0.1) → Generic Properties (1.1) → Grader Configs (1.2) → Grader Architecture (2.5) → Comparison Engine (4.1)

**Hard constraints (from design meeting):**
- **DM17:** No Phase 1 PR merges until CI (0.1) is green and enforced
- **DM14:** GraderInput/GraderResult types frozen before grader implementations begin
- **DM15:** Config migration (0.6) and property migration (1.1b) MUST NOT overlap
- **DM10:** Review package tests (1.8) complete before Phase 2 replaces reviewer

Generic Properties is the single most important model change because grader configs, tool filters, comparison engine, and `.hyoka` portability all depend on prompts having arbitrary properties. The Grader Architecture is the single most important structural change because it replaces the monolithic reviewer with composable, typed graders.

---

## Future Enhancement: Run Spec Files

> **Note (anchoring review, approved; ESC-2 — kept as future per Morpheus recommendation):** The current `hyoka run` command has 33+ flags and Phase 2 adds more. Consider a future `hyoka run eval.yaml` pattern where a run specification file absorbs most flags into a declarative spec (similar to Waza's eval spec files). This doesn't block the current CLI work (§1.5 split) but should be explored as the next evolution of the CLI surface after the split is complete and grader config format is battle-tested.

---

## Resolved Decisions (from Design Meeting 2026-04-04)

All 4 escalated decisions resolved by Ronnie Geraghty (2026-04-04T19:09Z):

### ESC-1: Serve `runID` path traversal security fix ✅

**Decision:** (b) Hotfix PR immediately, outside the plan. Security issues don't wait for sprint planning.

### ESC-2: Run spec file promotion to Phase 2 ✅

**Decision:** (b) Keep as "future". Let main.go split and grader config format stabilize first.

### ESC-3: Branch protection timing ✅

**Decision:** (a) Enable immediately once CI (#91) merges. Every subsequent Phase 0 PR benefits.

### ESC-4: Report migration ✅

**Decision:** Migrate reports in-place during Phase 2. No dual-format support, no new command. Old JSON gets rewritten to v2 schema as part of grader architecture work. `schema_version` field included for future-proofing but only latest version supported. Project is not in stable mode.

---

## Team Assignments

### Neo (Eval Engine, Review Pipeline, Graders)
- Phase 0: Reviewer model bug (0.2) — per-task factory function (DM1), log discarded errors (0.7)
- Phase 1: Generic properties (1.1a), grader configs (1.2b), tool filter wiring (1.3c), system prompt removal (1.6b-c), **reviewer system message removal (1.6d, DM7)**, starter files (1.7b)
- Phase 2: Pairwise package (2.1a), session limits wiring (2.2b), isolation (2.3a-c), resource audit (2.4a), grader implementations (2.5b-f)
- Phase 3: Action history capture (3.1b)
- Phase 4: Comparison engine (4.1a), trend slicing (4.3a-b)

### Tank (CLI, Config, Environment)
- Phase 0: CI pipeline (0.1) — with `-race`, 2m timeout (DM8), stale path (0.3), generator validation (0.4), duplicate configs (0.5), config file migration (0.6) — ~17 call sites including `main.go:394` and `Validate()`
- Phase 1: Property migration script (1.1b) — also update `new-prompt` scaffold (DM20), filter flag updates (1.1c), grader config migration (1.2c), tool filter schema (1.3b), main.go split (1.5a-c), config system prompt field (1.6a)
- Phase 2: Pairwise flag (2.1b), always_on field (2.1c), session limit config (2.2a)
- Phase 4: Compare command (4.1b)
- Phase 5: .hyoka init (5.1a-c) — walk-up stops at Git root (DM16), tool marketplace CLI (5.2b)

### Trinity (Reports, Templates, Serve)
- Phase 2: Pairwise report (2.1d), memory bounds (2.4b), grader report types + schema versioning (2.5g, DM11)
- Phase 3: **Template extraction (3.0, DM12)**, action timeline in reports (3.1c-d), grader result display (3.2a-b)
- Phase 4: Serve API (4.2a), comparison UI — React SPA (4.2b), pairwise visualization (4.2c)

### Switch (Testing, Quality)
- Phase 0: Flaky test fix (0.9), **pidfile tests (0.10, DM9: moved from Phase 5)**
- Phase 1: Starter file validation (1.7c), **review package coverage (1.8, DM10: moved from Phase 5)**
- Phase 5: integration test (5.3b), serve tests (5.3d)

### Oracle (Documentation)
- Phase 0: Go version docs (0.8) — 9+ files, not just 4
- All phases: Keep docs/ updated as features land. **Enumerate 12 distributed doc tasks during Phase 0 (DM18).** Feature owners draft docs; Oracle reviews and polishes. Breaking changes (Phase 1, Phase 2) require migration guides published BEFORE code lands.

### Morpheus (Architecture, Design)
- Phase 1: Grader config schema design (1.2a) — include `gate: bool` (DM3), tool filter schema design (1.3a), starter file format design (1.7a)
- Phase 2: Grader interface design (2.5a) — **types must be frozen before implementation (DM14)**
- Phase 3: Action event schema (3.1a)
- Phase 5: Tool registry format (5.2a)
- All phases: Architecture review, decision arbitration

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Generic properties breaks all 87 prompts | High | Medium | Big-bang migration with automated script + validation. DM2: metadata-only in map, typed fields retained for non-string data |
| **GraderResult type is most expensive to fix later** | **Critical** | Medium | DM4/DM5/DM14: typed fields, concrete input, freeze before implementation. File grader validates first |
| **~44 tests break in Phase 1** (~17% of suite) | High | High | Test helpers + golden files at Phase 1 start. Testing investment significantly underestimated per Switch |
| **Serve runID path traversal** (active vulnerability) | High | Low | ESC-1 ✅: hotfix PR outside the plan (approved) |
| Pairwise combinatorial explosion (too many configs) | Medium | Medium | Smart defaults (only toggle tools, not skills), always_on exemptions |
| Zero system prompt causes agent misbehavior | Medium | Low | DM6: phased removal in 4 stages. SDK session config handles isolation |
| main.go split introduces regressions | Medium | Medium | CI pipeline (Phase 0) catches regressions; split is mostly mechanical |
| Waza patterns don't translate to Copilot SDK | Medium | Low | Waza uses same Copilot SDK; patterns are directly applicable |
| Config migration (0.6) and property migration (1.1b) overlap | High | Medium | DM15: these must NOT overlap — 0.6 completes fully before 1.1b begins |
| Phase 2.5 Neo overload (5 tasks, 2 Large) | High | Medium | Morpheus delivers interface design (2.5a) early to unblock |
| Report schema migration breaks rerender | Medium | Medium | ESC-4 ✅: migrate reports in-place during Phase 2. No dual-format support. schema_version for future-proofing only |

---

## Timeline Estimate

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 0: Foundation | 1 sprint | 1 sprint |
| Phase 1: Core Model | 2-3 sprints | 3-4 sprints |
| Phase 2: Eval Engine | 2-3 sprints | 5-7 sprints |
| Phase 3: Transparency | 1-2 sprints | 6-9 sprints |
| Phase 4: Insights | 2 sprints | 8-11 sprints |
| Phase 5: Ecosystem | 2-3 sprints | 10-14 sprints |

**Note:** Phases 3 and 5 can overlap with Phase 2. Phase 5 testing tasks (5.3) can begin immediately. Skills work (5.4) can begin anytime.

---

## Reference Architecture

**microsoft/waza** — Azure SDK's production agent evaluation tool. Key patterns adopted:
- Zero system prompt (SDK config handles everything)
- `ResourceFile` pattern for starter files
- Session config–based isolation (working directory, tools, permissions)
- Config-level system prompt override for teams that want one
- **Pluggable grader architecture** — 12 grader types (code, prompt, file, program, behavior, action_sequence, etc.) with typed outputs and composable weights
