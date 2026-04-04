# Hyoka Evolution Plan

**Version:** 1.0  
**Date:** 2026-10-14  
**Author:** Morpheus (Lead/Architect)  
**Status:** Approved by Ronnie Geraghty

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
| 0.1 | **Create CI pipeline** — `go build`, `go test`, `go vet` on all PRs and pushes. Add to `.github/workflows/ci.yml`. | Tank | Medium | — |
| 0.2 | **Fix reviewer model bug** — `main.go:469-473` grabs reviewers from first config only. Each config must use its own reviewer panel. | Neo | Small | — |
| 0.3 | **Fix stale path in new-prompt** — `main.go:1276` says `go run ./tool/cmd/hyoka validate`, should be `go run ./hyoka validate`. | Tank | Small | — |
| 0.4 | **Add generator model validation** — Empty `Generator.Model` passes validation but fails at runtime (`config.go:256-287`). | Tank | Small | — |
| 0.5 | **Detect duplicate config names** — Two configs with the same name silently shadow. Second config becomes inaccessible. | Tank | Small | — |
| 0.6 | **Log discarded errors** — `reviewer.go:352`, `copilot.go:83`, `main.go:219,263,270,286`, `fetcher.go:68,82,88,120`. | Neo | Small | — |
| 0.7 | **Fix Go version in docs** — 4 files say Go 1.24.5+, go.mod requires 1.26.1. | Oracle | Small | — |
| 0.8 | **Fix flaky resourcemonitor tests** — Replace `time.Sleep` assertions with event-driven checks. | Switch | Small | — |

---

## Phase 1: Core Model Changes

**Goal:** Generalize hyoka's data model so it's no longer Azure-specific. This is the architectural foundation every subsequent phase depends on.

**Duration:** 2-3 sprints

### 1.1 — Generic Prompt Properties

**Current:** `Prompt` struct (`prompt/types.go:4-32`) has 12+ hardcoded fields: `Service`, `Plane`, `Language`, `Category`, `Difficulty`, `SDKPackage`, `DocURL`, etc.

**Target:** Add `Properties map[string]string` field. All filtering, matching, and criteria selection works against properties. Existing fields become convenience accessors that read from the properties map.

**Migration strategy:** Big-bang. Update all 87 prompts to the new format. No backward compatibility for old fields. Write a migration script that reads existing frontmatter and converts to properties format.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.1a | Add `Properties map[string]string` to `Prompt` struct. Update parser to populate properties from frontmatter. | Neo | Medium |
| 1.1b | Write migration script to convert all 87 prompts from fixed fields to properties. | Tank | Medium |
| 1.1c | Update all filter flags (`--service`, `--language`, `--plane`, `--category`) to query properties map instead of struct fields. | Tank | Medium |

### 1.2 — Criteria Filters (Property-Based with Exclude)

**Current:** `MatchCondition` struct (`criteria/criteria.go:28-34`) has 5 hardcoded fields: `Language`, `Service`, `Plane`, `Category`, `SDK`.

**Target:** Replace with property-based matching. Support both include and exclude filters. Remove Tier 1 criteria entirely — prompts/configs must supply their own criteria.

**Decision (Q1):** Remove Tier 1 entirely. No built-in default criteria.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.2a | Replace `MatchCondition` with `map[string]string` include/exclude property matching. | Neo | Medium |
| 1.2b | Remove Tier 1 criteria system. Update all code paths that reference default criteria. | Neo | Small |
| 1.2c | Update criteria YAML format to use property-based conditions. Migrate existing criteria files. | Tank | Medium |

### 1.3 — Tool Filters (Property-Based)

**Current:** Config YAML supports `available_tools` and `excluded_tools` as flat string lists.

**Target:** Tools can be filtered based on prompt properties. A config can say "for Python prompts, enable tool X; for Go prompts, disable tool Y." This enables property-driven tool availability — the foundation for pairwise testing.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.3a | Design tool filter schema that references prompt properties. | Morpheus | Small |
| 1.3b | Implement property-based tool filter resolution in config loading. | Tank | Medium |
| 1.3c | Wire tool filters into session config builder (`copilot.go:617-693`). | Neo | Medium |

### 1.4 — YAML Prompt Format

**Current:** Prompts are Markdown files with YAML frontmatter. The prompt text is the Markdown body.

**Target:** Support pure YAML prompt format alongside Markdown. YAML prompts are more structured and easier to generate/validate programmatically.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.4a | Add `ParsePromptYAML()` function alongside existing Markdown parser. | Neo | Medium |
| 1.4b | Update prompt loader to auto-detect format by file extension (`.prompt.md` vs `.prompt.yaml`). | Neo | Small |

### 1.5 — Split main.go

**Current:** 1329 lines in `hyoka/main.go`. All 14 cobra commands, flag definitions, path resolution, reviewer wiring, skill installation, and prompt scaffolding in one function.

**Target:** Create `hyoka/cmd/` package with per-command files: `run.go`, `list.go`, `serve.go`, `clean.go`, `validate.go`, `check_env.go`, `new_prompt.go`, `trends.go`, `rerender.go`.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.5a | Create `hyoka/cmd/` package structure with root command and shared state. | Tank | Large |
| 1.5b | Move each command to its own file with RunE function. | Tank | Large |
| 1.5c | Update main.go to be a thin entry point that registers commands from cmd/. | Tank | Small |

### 1.6 — Configurable System Prompts

**Current:** System prompt is hardcoded in `copilot.go:628-655` with 15 rules. All agent sessions get the same prompt.

**Target:** 
- **Default:** Zero system prompt. No rules injected. SDK session config handles working directory, tools, and isolation.
- **Config-level override:** Config YAML has optional `generator.system_prompt` and `reviewer.system_prompt` fields. Users who want specific behavior can set custom system prompts.
- **Both gen and review agents** support configurable system prompts.

**Decision (Q2, Q6):** Follow Waza's approach. Zero system prompt by default. All isolation/directory config handled through SDK `SessionConfig`. Config-specific custom system prompts available for users who want them. Response type (files vs text) is a system-prompt-level concern.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.6a | Add `system_prompt` field to both generator and reviewer config sections. | Tank | Small |
| 1.6b | Modify `buildSessionConfig()` to use config system prompt (or empty if not set) instead of hardcoded rules. | Neo | Medium |
| 1.6c | Remove the 15 hardcoded rules. Move any truly operational rules (file path enforcement) to SDK hooks or pre/post validation. | Neo | Medium |

### 1.7 — Starter Files

**Current:** `StarterProject` field exists on `Prompt` struct but implementation is minimal. `copilot.go:83` silently discards errors when handling starter files.

**Target:** Full starter file support. Prompts can reference a directory of files that get copied into the agent's working directory before the session begins. This enables "fix this broken code" style prompts.

**Decision:** Core feature. Follow Waza's `ResourceFile` pattern — files placed in workspace before session starts.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 1.7a | Design starter file reference format in prompt frontmatter (directory path or file list). | Morpheus | Small |
| 1.7b | Implement file copying into workspace before session creation. Fix silent error handling at `copilot.go:83`. | Neo | Medium |
| 1.7c | Add starter file validation to prompt validation pipeline. | Switch | Small |

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
| 2.1a | Create `pairwise/` package that generates N+1 config variants from a base config (one per tool toggle + baseline). | Neo | Large |
| 2.1b | Add `--pairwise` flag to run command. Wire into eval engine to expand configs before execution. | Tank | Medium |
| 2.1c | Add `always_on` field to tool config. Exempt marked tools from pairwise toggling. | Tank | Small |
| 2.1d | Generate pairwise comparison report showing per-tool impact scores. | Trinity | Medium |

### 2.2 — User-Configurable Session Limits

**Current:** Hardcoded guardrails: 25 turns, 50 files, 1 MB output, 50 max session actions.

**Target:** All limits configurable in config YAML with sensible defaults. Users can tune for their use case.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.2a | Add session limit fields to config schema (max_turns, max_files, max_output_size, max_session_actions). | Tank | Small |
| 2.2b | Wire config limits into eval engine guardrail checks. | Neo | Medium |

### 2.3 — Isolated Evaluation Environment

**Current:** Agent sessions run in the user's environment. Leakage possible.

**Target:** Each evaluation session runs in a clean, isolated workspace. No leakage from user's dev environment into agent sessions.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.3a | Create isolated workspace directory per session with clean environment. | Neo | Medium |
| 2.3b | Copy only declared starter files and config-specified resources into workspace. | Neo | Small |
| 2.3c | Clean up workspace after session completes (integrate with existing clean command). | Neo | Small |

### 2.4 — Resource Efficiency

**Current:** Some concerns about memory usage with large eval runs. No resource monitoring by default.

**Target:** Bounded memory usage, proper goroutine cleanup, no resource leaks. `--monitor-resources` flag already exists — ensure it catches real issues.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 2.4a | Audit goroutine lifecycle across eval/review pipeline. Ensure all goroutines terminate on context cancellation. | Neo | Medium |
| 2.4b | Add memory bounds for in-flight report data. Stream large reports to disk. | Trinity | Medium |

---

## Phase 3: Transparency

**Goal:** Make every agent action visible. Nothing hidden from the evaluator.

**Duration:** 1-2 sprints  
**Depends on:** Phase 1 (system prompt changes), Phase 0 (bug fixes)

### 3.1 — Full Agent Action History

**Current:** Reports show final output and scores. Individual agent actions (tool calls, file reads, bash commands) are logged but not structured for display.

**Target:** Every action the agent takes during a session is captured in a structured timeline. Reports show the full action sequence, not just the result.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 3.1a | Define action event schema (timestamp, type, tool, input, output, duration). | Morpheus | Small |
| 3.1b | Capture all agent events from SDK session hooks into structured action log. | Neo | Medium |
| 3.1c | Include full action timeline in JSON report output. | Trinity | Medium |
| 3.1d | Render action timeline in HTML reports (expandable, searchable). | Trinity | Large |

### 3.2 — Transparent Review Panel

**Current:** Review panel submits scores. Individual reviewer reasoning is captured but not prominently displayed.

**Target:** Each reviewer's full reasoning, scores, and criteria evaluation is visible. The consolidation logic (how individual scores become final scores) is explicit.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 3.2a | Include per-reviewer full response in report output (not just scores). | Neo | Medium |
| 3.2b | Show consolidation algorithm and weights in report. | Trinity | Small |
| 3.2c | Add reviewer tool environments — configurable tools for review agents (linters, style checkers, doc references). | Neo | Medium |

---

## Phase 4: Insights & Comparison

**Goal:** Make hyoka's primary output actionable data comparison, not just pass/fail grades.

**Duration:** 2 sprints  
**Depends on:** Phase 2 (pairwise testing), Phase 3 (action history)

### 4.1 — Comparison Engine

**Current:** Trends package (`trends/trends.go`, 857 lines) does cross-run analysis but it's static and batch-oriented.

**Target:** Dynamic comparison engine that can diff any two configs, runs, or time periods. Answer questions like "How did adding the Azure MCP server change scores for Python prompts?"

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 4.1a | Create `comparison/` package with config-vs-config, run-vs-run, and temporal diff logic. | Neo | Large |
| 4.1b | Add `hyoka compare` command for CLI-based comparison. | Tank | Medium |
| 4.1c | Expose comparison data via serve API for the dashboard. | Trinity | Medium |

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
| 4.2a | Design dashboard API endpoints for comparison, trends, and drill-down. | Trinity | Medium |
| 4.2b | Implement dynamic comparison UI with filter controls. | Trinity | Large |
| 4.2c | Add pairwise impact visualization (tool heatmap, contribution charts). | Trinity | Large |

### 4.3 — Enhanced Trends

**Current:** Trends show score changes over time with basic aggregation.

**Target:** Property-aware trend analysis. Slice trends by any prompt property combination. Detect regressions automatically.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 4.3a | Add property-based trend slicing (trend by language, by service, by tool config). | Neo | Medium |
| 4.3b | Add automatic regression detection with configurable thresholds. | Neo | Medium |

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
| 5.1a | Add `hyoka init` command that scaffolds `.hyoka/` directory with subdirs and gitignore. | Tank | Medium |
| 5.1b | Auto-discover `.hyoka/` in current directory and ancestors. Merge with CLI flags. | Tank | Medium |
| 5.1c | Update all path resolution to check `.hyoka/` first, then fall back to repo root paths. | Tank | Medium |

### 5.2 — Tool Marketplace / Repos

**Target:** Curated tool configurations (MCP servers, skills) that teams can reference by name instead of full config blocks.

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 5.2a | Design tool registry format (YAML catalog of MCP server configs, skills). | Morpheus | Medium |
| 5.2b | Implement `hyoka tools list` and `hyoka tools add` commands. | Tank | Medium |
| 5.2c | Support remote tool registries (GitHub repos as tool catalogs). | Neo | Medium |

### 5.3 — Test Infrastructure

**Tasks:**

| Task | Description | Owner | Size |
|------|-------------|-------|------|
| 5.3a | Add pidfile tests (zero tests today, only untested package). | Switch | Medium |
| 5.3b | Add stub-based integration test (StubEvaluator + StubReviewer → engine.Run() → verify report). | Switch | Medium |
| 5.3c | Increase review package coverage (5 tests for 732 lines currently). | Switch | Medium |
| 5.3d | Add serve handler tests (path traversal for `runID`, functional endpoint tests). | Switch | Medium |

### 5.4 — Skills

**Target:** Create 14 project-specific skills to encode hyoka's architecture, patterns, and domain knowledge. Skills are advisory, not prescriptive — guardrails not cages.

**Skill categories:**
- Core Architecture (5): eval-pipeline, error-handling, config-system, copilot-sdk-integration, criteria-system
- Working Patterns (4): testing-patterns, cli-patterns, report-generation, logging-conventions
- Human Developer (2): contributor-guide, prompt-conventions
- Evolution Support (3): property-migration, process-lifecycle, serve-patterns

---

## Dependency Graph

```
Phase 0 (Foundation)
  ├── 0.1 CI Pipeline ─────────────────────────┐
  ├── 0.2 Reviewer model bug fix               │
  ├── 0.3 Fix stale path                       │
  ├── 0.4 Generator model validation            │
  ├── 0.5 Duplicate config detection            │
  ├── 0.6 Log discarded errors                  │
  ├── 0.7 Fix Go version in docs               │
  └── 0.8 Fix flaky tests                      │
                                                │
Phase 1 (Core Model) ──── requires 0.1 ────────┘
  ├── 1.1 Generic Properties ──────────────────┐
  │     └── 1.2 Criteria Filters ──────────────┤ (requires 1.1)
  │     └── 1.3 Tool Filters ─────────────────┤ (requires 1.1)
  ├── 1.4 YAML Prompt Format                    │
  ├── 1.5 Split main.go                        │
  ├── 1.6 Configurable System Prompts          │
  └── 1.7 Starter Files                        │
                                                │
Phase 2 (Eval Engine) ── requires 1.1-1.3 ─────┘
  ├── 2.1 Pairwise Testing ──── requires 1.3 Tool Filters
  ├── 2.2 Session Limits
  ├── 2.3 Isolated Environment
  └── 2.4 Resource Efficiency
                                                
Phase 3 (Transparency) ── requires 1.6
  ├── 3.1 Full Action History
  └── 3.2 Transparent Review Panel
                                                
Phase 4 (Insights) ── requires 2.1, 3.1
  ├── 4.1 Comparison Engine
  ├── 4.2 Serve Site Evolution ── requires 4.1
  └── 4.3 Enhanced Trends
                                                
Phase 5 (Ecosystem) ── requires 1.1, 1.4
  ├── 5.1 .hyoka Directory
  ├── 5.2 Tool Marketplace
  ├── 5.3 Test Infrastructure (can start anytime)
  └── 5.4 Skills (can start anytime)
```

**Critical path:** CI (0.1) → Generic Properties (1.1) → Tool Filters (1.3) → Pairwise Testing (2.1) → Comparison Engine (4.1)

Generic Properties is the single most important model change because criteria filters, tool filters, comparison engine, and `.hyoka` portability all depend on prompts having arbitrary properties.

---

## Team Assignments

### Neo (Eval Engine, Review Pipeline)
- Phase 0: Reviewer model bug (0.2), log discarded errors (0.6)
- Phase 1: Generic properties (1.1a), criteria filters (1.2a-b), tool filter wiring (1.3c), system prompt removal (1.6b-c), starter files (1.7b)
- Phase 2: Pairwise package (2.1a), session limits wiring (2.2b), isolation (2.3a-c), resource audit (2.4a)
- Phase 3: Action history capture (3.1b), review panel transparency (3.2a, 3.2c)
- Phase 4: Comparison engine (4.1a), trend slicing (4.3a-b)

### Tank (CLI, Config, Environment)
- Phase 0: CI pipeline (0.1), stale path (0.3), generator validation (0.4), duplicate configs (0.5)
- Phase 1: Property migration script (1.1b), filter flag updates (1.1c), criteria migration (1.2c), tool filter schema (1.3b), main.go split (1.5a-c), config system prompt field (1.6a)
- Phase 2: Pairwise flag (2.1b), always_on field (2.1c), session limit config (2.2a)
- Phase 4: Compare command (4.1b)
- Phase 5: .hyoka init (5.1a-c), tool marketplace CLI (5.2b)

### Trinity (Reports, Templates, Serve)
- Phase 2: Pairwise report (2.1d), memory bounds (2.4b)
- Phase 3: Action timeline in reports (3.1c-d), consolidation display (3.2b)
- Phase 4: Serve API (4.2a), comparison UI (4.2b), pairwise visualization (4.2c)

### Switch (Testing, Quality)
- Phase 0: Flaky test fix (0.8)
- Phase 1: Starter file validation (1.7c)
- Phase 5: pidfile tests (5.3a), integration test (5.3b), review coverage (5.3c), serve tests (5.3d)

### Oracle (Documentation)
- Phase 0: Go version docs (0.7)
- All phases: Keep docs/ updated as features land

### Morpheus (Architecture, Design)
- Phase 1: Tool filter schema design (1.3a), starter file format design (1.7a)
- Phase 3: Action event schema (3.1a)
- Phase 5: Tool registry format (5.2a)
- All phases: Architecture review, decision arbitration

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Generic properties breaks all 87 prompts | High | Medium | Big-bang migration with automated script + validation |
| Pairwise combinatorial explosion (too many configs) | Medium | Medium | Smart defaults (only toggle tools, not skills), always_on exemptions |
| Zero system prompt causes agent misbehavior | Medium | Low | SDK session config handles isolation; hooks validate file paths; monitor first runs closely |
| main.go split introduces regressions | Medium | Medium | CI pipeline (Phase 0) catches regressions; split is mostly mechanical |
| Waza patterns don't translate to Copilot SDK | Medium | Low | Waza uses same Copilot SDK; patterns are directly applicable |

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
