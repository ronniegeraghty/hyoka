# Hyoka Feature Requirements

**Date:** 2026-10-14  
**Author:** Morpheus (Lead/Architect)  
**Source:** Ronnie Geraghty's product direction + architecture analysis  
**Status:** Approved

---

## FR-01: Full Agent Action History

**Description:** Capture and display every action the agent takes during an evaluation session — tool calls, file reads/writes, bash commands, web fetches — in a structured timeline.

**User Story:** As an evaluator, I want to see the complete sequence of actions an agent took so that I can understand *how* it arrived at its output, not just what the output was.

**Acceptance Criteria:**
- [ ] Every agent action is captured with timestamp, type, tool name, input, output, and duration
- [ ] JSON reports include full action timeline array
- [ ] HTML reports render action timeline as an expandable, searchable section
- [ ] Action timeline is sortable by time, type, or tool
- [ ] No actions are omitted or summarized — full fidelity

**Dependencies:** FR-18 (configurable system prompts — action capture relies on SDK hooks, not system prompt rules)

**Phase:** 3

---

## FR-02: Config-Driven Tool Availability

**Description:** Allow config YAML to specify which tools are available to the generation agent, with support for property-based filters (e.g., "enable tool X only for Python prompts").

**User Story:** As a team lead, I want to control which tools agents can use per-config so that I can measure tool impact across different configurations.

**Acceptance Criteria:**
- [ ] Config YAML supports `available_tools` and `excluded_tools` with property-based conditions
- [ ] Tool filters resolve against prompt properties at runtime
- [ ] Invalid tool references produce clear error messages
- [ ] Tool availability is reflected in session config passed to SDK

**Dependencies:** FR-12 (customizable prompt properties — tool filters reference properties)

**Phase:** 1

---

## FR-03: Zero System Prompt

**Description:** Default to an empty system prompt for agent evaluation sessions. All operational configuration (working directory, tool availability, isolation) handled through SDK `SessionConfig`, not prompt injection.

**User Story:** As an evaluator, I want agents to be tested without hidden behavioral guidance so that I get unbiased measurements of tool effectiveness.

**Acceptance Criteria:**
- [ ] Default system prompt is empty string (no rules injected)
- [ ] All 15 current hardcoded rules removed from `copilot.go:628-655`
- [ ] Working directory, tool config, and permissions set via SDK `SessionConfig` fields
- [ ] File path validation moved to pre/post hooks (not system prompt rules)
- [ ] Agent sessions function correctly with zero system prompt (file creation, bash usage work)
- [ ] Config YAML supports optional `generator.system_prompt` override for users who want custom prompts

**Dependencies:** FR-18 (configurable system prompts — zero is the default, custom is the override)

**Phase:** 1

---

## FR-04: Per-Project `.hyoka` Directory

**Description:** Support a `.hyoka/` directory in any project that contains team-specific configs, prompts, criteria, and skills. Auto-discovered by hyoka when run from within the project.

**User Story:** As a team member, I want to check in our evaluation setup alongside our code so that the whole team uses consistent eval configs without global setup.

**Acceptance Criteria:**
- [ ] `hyoka init` command creates `.hyoka/` with subdirectories: `configs/`, `prompts/`, `criteria/`, `skills/`, `reports/`
- [ ] `.hyoka/reports/` is gitignored by default
- [ ] hyoka auto-discovers `.hyoka/` in CWD and ancestor directories
- [ ] `.hyoka/` resources merge with (and override) CLI flag paths
- [ ] Works without global hyoka installation — project-scoped only
- [ ] Structured subdirectories (not flat files)

**Dependencies:** FR-12 (customizable prompt properties — `.hyoka` prompts use generic properties)

**Phase:** 5

---

## FR-05: Transparent Review Panel → Grader Result Transparency

**Description:** ~~Expose the full reasoning, scores, and criteria evaluation from each individual reviewer in the review panel.~~ **Absorbed into grader architecture (D-AR1).** With pluggable graders, transparency is inherent — each grader produces typed, structured output. A `program` grader shows command + exit code + output. A `prompt` grader shows model + rubric + reasoning. A `file` grader shows which files were checked.

**User Story:** As an evaluator, I want to see exactly what each grader checked and how it scored so that I can trust the evaluation and identify which checks passed or failed.

**Acceptance Criteria:**
- [ ] Each grader result includes kind, name, score, weight, pass/fail, and typed details
- [ ] HTML reports render each grader type with an appropriate display component
- [ ] Weighted score aggregation formula visible in report
- [ ] No "black box" — every evaluation check is individually inspectable

**Dependencies:** FR-19 (Grader Architecture)

**Phase:** 3

---

## FR-06: Tool Marketplace / Repos

**Description:** Curated registry of tool configurations (MCP servers, skills) that teams can reference by name instead of writing full config blocks.

**User Story:** As a new user, I want to add well-known tools by name (e.g., `azure-mcp`, `github-mcp`) so that I don't have to figure out the correct config syntax from scratch.

**Acceptance Criteria:**
- [ ] Tool registry format defined (YAML catalog with name, type, config block)
- [ ] `hyoka tools list` shows available tools
- [ ] `hyoka tools add <name>` adds tool config to current config file
- [ ] Support remote registries (GitHub repos as tool catalogs)
- [ ] Registry entries include version, description, and compatibility metadata

**Dependencies:** FR-02 (config-driven tool availability — marketplace tools go into configs)

**Phase:** 5

---

## FR-07: First-Class Insights and Comparison

**Description:** Dynamic comparison engine that can diff configs, runs, or time periods. Interactive dashboard with filter controls, drill-down, and visualization.

**User Story:** As a team lead, I want to compare two configs side-by-side to see which tools improved scores and which caused regressions, without manually reading JSON reports.

**Acceptance Criteria:**
- [ ] `hyoka compare` CLI command for config-vs-config and run-vs-run comparison
- [ ] Comparison output shows per-criteria score deltas
- [ ] Serve dashboard includes interactive comparison view with filters
- [ ] Property-based filtering (compare "only Python prompts" or "only data-plane")
- [ ] Pairwise impact visualization (tool heatmap, contribution charts)
- [ ] Automatic regression detection with configurable thresholds

**Dependencies:** FR-15 (pairwise testing — comparison is most valuable with pairwise data), FR-01 (action history — drill-down shows what changed)

**Phase:** 4

---

## FR-08: User-Configurable Session Limits

**Description:** Make all evaluation guardrails configurable in config YAML while keeping sensible defaults.

**User Story:** As an evaluator, I want to increase the turn limit for complex prompts so that agents have enough room to complete the task without artificial constraints.

**Acceptance Criteria:**
- [ ] Config YAML supports: `max_turns`, `max_files`, `max_output_size`, `max_session_actions`
- [ ] Defaults match current values: 25 turns, 50 files, 1 MB output, 50 actions
- [ ] Limits enforced by eval engine guardrail checks
- [ ] When a limit is hit, report clearly states which limit and what the configured value was

**Dependencies:** None

**Phase:** 2

---

## FR-09: Resource Efficiency

**Description:** Ensure hyoka is a responsible resource citizen — bounded memory, proper goroutine cleanup, no leaks during long evaluation runs.

**User Story:** As a user running large batch evaluations, I want hyoka to not consume unbounded memory or leave orphan processes so that my machine remains usable.

**Acceptance Criteria:**
- [ ] All goroutines terminate on context cancellation (audited and verified)
- [ ] Memory usage bounded for in-flight report data (stream large reports to disk)
- [ ] `--monitor-resources` flag reports peak memory and goroutine count
- [ ] No orphan processes after normal or abnormal termination
- [ ] PID birth-time validation prevents killing wrong process on PID reuse

**Dependencies:** None

**Phase:** 2

---

## FR-10: Isolated Evaluation Environment

**Description:** Each evaluation session runs in a clean, isolated workspace with no leakage from or to the user's environment.

**User Story:** As an evaluator, I want each eval session to start with a clean slate so that results aren't contaminated by leftover state from previous runs or my dev environment.

**Acceptance Criteria:**
- [ ] Each session gets a fresh workspace directory
- [ ] Only declared starter files and config resources copied into workspace
- [ ] No ambient environment variables, dotfiles, or user configs leak in
- [ ] Workspace cleaned up after session completes (or by `hyoka clean`)
- [ ] Isolation verified by test that checks workspace contents pre/post eval

**Dependencies:** FR-16 (starter files — isolated workspace needs to know what to pre-populate)

**Phase:** 2

---

## FR-11: Tool Filters (Property-Based)

**Description:** Filter tool availability based on prompt properties. A config can conditionally enable or disable tools based on which prompt is being evaluated.

**User Story:** As a team lead, I want to enable the Azure MCP server only for Azure SDK prompts so that non-Azure evaluations don't have irrelevant tools available.

**Acceptance Criteria:**
- [ ] Config YAML supports property conditions on tool entries (e.g., `when: { service: key-vault }`)
- [ ] Conditions evaluated against prompt properties at runtime
- [ ] Tools with no conditions are always available
- [ ] Tools can be marked `always_on: true` (exempt from pairwise toggling)
- [ ] Clear error when conditions reference non-existent properties

**Dependencies:** FR-12 (customizable prompt properties — conditions reference properties)

**Phase:** 1

---

## FR-12: Customizable Prompt Properties (Sole Representation)

**Description:** Replace hardcoded prompt struct fields with `Properties map[string]string` as THE representation. Drop typed struct fields entirely. Add convenience getter methods (`Language()`, `Service()`, etc.) that read from the properties map.

**User Story:** As a non-Azure team, I want to add custom metadata to my prompts (e.g., `framework: django`, `complexity: high`) so that I can filter and match graders using my own taxonomy.

**Acceptance Criteria:**
- [ ] `Prompt` struct uses `Properties map[string]string` as sole data store for metadata
- [ ] Typed fields (`Service`, `Language`, `Plane`, etc.) removed from struct — replaced by getter methods
- [ ] All filter flags (`--service`, `--language`, etc.) query properties map via getter methods
- [ ] `Filter` struct redesigned as `map[string]string` with support for generic `--filter key=value`
- [ ] All 87 existing prompts migrated to nested `properties:` format (big-bang, no backward compat)
- [ ] Migration script provided and validated
- [ ] Prompt frontmatter uses `properties:` key with nested key-value pairs

**Dependencies:** None (foundational — many other features depend on this)

**Phase:** 1

---

## FR-13: YAML Prompt Format

**Description:** Support pure YAML prompt format (`.prompt.yaml`) alongside existing Markdown format (`.prompt.md`).

**User Story:** As a CI pipeline author, I want to generate prompts in YAML format so that I can programmatically create and validate prompts without markdown parsing.

**Acceptance Criteria:**
- [ ] `ParsePromptYAML()` function parses `.prompt.yaml` files
- [ ] Prompt loader auto-detects format by file extension
- [ ] Both formats produce identical `Prompt` structs
- [ ] `hyoka validate` works with both formats
- [ ] `hyoka new-prompt` can scaffold either format

**Dependencies:** FR-12 (customizable properties — YAML format uses properties)

**Phase:** 1

---

## FR-14: Criteria → Grader Configs

**Description:** ~~Replace hardcoded `MatchCondition` fields with property-based include/exclude matching. Remove Tier 1 criteria.~~ **Absorbed into grader architecture (D-AR1).** Criteria become grader configurations — no tier system, no `MergeCriteria()`. Each grader is defined in YAML with `kind:`, `name:`, `config:`, `when:`, `weight:`. Property-based `when:` conditions replace `MatchCondition`.

**User Story:** As a team lead, I want to define evaluation checks as composable graders with typed configs so that I can mix deterministic checks (file existence, build success) with LLM-based review, each weighted appropriately.

**Acceptance Criteria:**
- [ ] Grader config YAML format defined with `kind:`, `name:`, `config:`, `when:`, `weight:`
- [ ] `MatchCondition` replaced with `when: map[string]string` property matching
- [ ] Tier 1 criteria removed — no built-in defaults
- [ ] `MergeCriteria()` and `FormatCriteria()` deleted
- [ ] Existing criteria files migrated to grader config format
- [ ] Graders without `when:` always apply; graders with `when:` are property-matched

**Dependencies:** FR-12 (customizable properties — `when:` conditions reference properties), FR-19 (Grader Architecture)

**Phase:** 1

---

## FR-15: Pairwise Testing

**Description:** `--pairwise` flag that expands a config into all pairwise variants — each tool toggled on/off — to measure individual tool impact.

**User Story:** As an evaluator, I want to run pairwise tests so that I can isolate which individual tool improved or degraded the agent's output quality.

**Acceptance Criteria:**
- [ ] `--pairwise` / `-pw` flag on `run` command
- [ ] Generates N+1 config variants: baseline (all tools) + one per tool removed
- [ ] Config YAML supports `always_on: true` per tool (exempt from toggling)
- [ ] Pairwise results include per-tool impact score (delta from baseline)
- [ ] Reports clearly label each variant and which tool was toggled
- [ ] `--pairwise` works with `--config` (expands specified config)

**Dependencies:** FR-11 (tool filters — pairwise toggling uses tool filter mechanism), FR-02 (config-driven tools)

**Phase:** 2

---

## FR-16: Starter Files

**Description:** Pre-populate the agent's working directory with files before the session begins. Enables "fix this code" and "extend this project" style prompts.

**User Story:** As a prompt author, I want to give the agent a project with broken code so that I can evaluate its ability to debug and fix existing code, not just generate from scratch.

**Acceptance Criteria:**
- [ ] Prompt frontmatter supports `starter_files` field (directory path or explicit file list)
- [ ] Files copied into agent workspace before SDK session starts
- [ ] Starter files are read-only source — originals never modified
- [ ] Missing starter files produce clear validation errors
- [ ] `hyoka validate` checks that referenced starter files exist
- [ ] Starter file errors logged (fix silent discard at `copilot.go:83`)

**Dependencies:** None

**Phase:** 1

---

## FR-17: Reviewer Tools → `program` Grader Type

**Description:** ~~Configurable tool environments for review panel agents.~~ **Absorbed into grader architecture (D-AR1).** The concept of "reviewer tools" becomes the `program` grader type — a grader that runs an external command (linter, compiler, test suite, style checker) and scores based on the result. No separate "reviewer tools" concept needed; a `program` grader IS a tool.

**User Story:** As an evaluator, I want to run linters and build checks as part of evaluation so that I get deterministic quality signals alongside LLM-based review.

**Acceptance Criteria:**
- [ ] `program` grader type runs external commands with configurable args
- [ ] Exit code determines pass/fail; stdout/stderr captured in grader result
- [ ] Multiple `program` graders can run per evaluation (linter + compiler + tests)
- [ ] `program` grader config specifies command, args, working directory, timeout
- [ ] Results include full command output for debugging

**Dependencies:** FR-19 (Grader Architecture)

**Phase:** 2

---

## FR-18: Configurable System Prompts

**Description:** Both the generation agent and review agents support optional system prompts specified in config YAML. Default is empty (zero system prompt).

**User Story:** As a team lead, I want to set a custom system prompt for my team's evaluations so that I can instruct agents to follow our coding conventions — but only when I choose to.

**Acceptance Criteria:**
- [ ] Config YAML supports `generator.system_prompt` (string or file path)
- [ ] Config YAML supports `reviewer.system_prompt` (string or file path)
- [ ] Default system prompt is empty (zero system prompt)
- [ ] System prompts passed to SDK `SessionConfig.SystemMessage`
- [ ] System prompt content visible in reports (transparency)
- [ ] Existing 15 hardcoded rules removed from code
- [ ] Response type (files vs text) handled via system prompt, not hardcoded

**Dependencies:** None

**Phase:** 1

---

## FR-19: Grader Architecture

**Description:** Replace the monolithic `Reviewer`/`PanelReviewer` pattern with a pluggable `Grader` interface. Each grader is focused (one concern), typed (specific input/output), composable (weighted list), and deterministic where possible. The current LLM reviewer becomes the `prompt` grader type. New grader types enable deterministic checks that don't require LLMs.

**User Story:** As an evaluator, I want evaluation to use deterministic checks (file existence, build success, schema validation) alongside LLM review so that objective criteria are checked reliably without LLM variance.

**Acceptance Criteria:**
- [ ] `Grader` interface defined in `hyoka/internal/graders/` with `Kind()`, `Name()`, `Grade()` methods
- [ ] `GraderResult` type with kind, name, score, weight, pass/fail, and typed details
- [ ] Initial grader types: `file`, `program`, `prompt`, `behavior`, `action_sequence`, `tool_constraint`
- [ ] Current LLM reviewer wrapped as `prompt` grader (one rubric per grader instance)
- [ ] Graders defined in config YAML with `kind:`, `name:`, `config:`, `when:`, `weight:`
- [ ] Eval engine collects applicable graders, runs all, aggregates weighted scores
- [ ] Reports center on `[]GraderResult` — each grader's output individually visible
- [ ] `program` grader runs external commands — enables linters, compilers, test suites as evaluation checks

**Dependencies:** FR-12 (customizable properties — grader `when:` conditions reference properties)

**Phase:** 2

---

## Feature Dependency Map

```
FR-12 (Properties) ──┬── FR-14 (Grader Configs)
                      ├── FR-11 (Tool Filters) ──── FR-15 (Pairwise)
                      ├── FR-02 (Config Tools) ──── FR-15 (Pairwise)
                      ├── FR-04 (.hyoka Dir)
                      ├── FR-13 (YAML Format)
                      └── FR-19 (Grader Architecture)

FR-19 (Graders) ──┬── FR-05 (Grader Transparency)
                   └── FR-17 (program grader)

FR-18 (System Prompts) ── FR-03 (Zero System Prompt)
                       └── FR-01 (Action History)

FR-16 (Starter Files) ── FR-10 (Isolation)

FR-15 (Pairwise) ─┬── FR-07 (Comparison)
FR-01 (History)  ──┘
```

---

## Priority Order

| Priority | Feature | Rationale |
|----------|---------|-----------|
| 1 | FR-12 Properties | Foundation — everything depends on this |
| 2 | FR-18 System Prompts | Unblocks zero system prompt and action history |
| 3 | FR-03 Zero System Prompt | Core principle — unbiased measurement |
| 4 | FR-14 Grader Configs | Needed for composable evaluation |
| 5 | FR-11 Tool Filters | Needed for pairwise |
| 6 | FR-16 Starter Files | Enables new prompt categories |
| 7 | FR-13 YAML Format | Enables programmatic prompt creation |
| 8 | FR-02 Config Tools | Foundation for pairwise |
| 9 | FR-19 Grader Architecture | Core evaluator redesign |
| 10 | FR-08 Session Limits | Low effort, high value |
| 11 | FR-15 Pairwise | Core differentiator |
| 12 | FR-01 Action History | Core transparency |
| 13 | FR-05 Grader Transparency | Inherent from grader design |
| 14 | FR-17 program grader | Extends eval with deterministic checks |
| 14 | FR-10 Isolation | Production readiness |
| 15 | FR-09 Resource Efficiency | Production readiness |
| 16 | FR-07 Comparison | Primary insights output |
| 17 | FR-04 .hyoka Dir | Ecosystem / adoption |
| 18 | FR-06 Tool Marketplace | Ecosystem / adoption |
