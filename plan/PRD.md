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

## FR-05: Transparent Review Panel

**Description:** Expose the full reasoning, scores, and criteria evaluation from each individual reviewer in the review panel. Show how individual scores are consolidated into final scores.

**User Story:** As an evaluator, I want to see what each reviewer thought and how the final score was calculated so that I can trust the review process and identify disagreements.

**Acceptance Criteria:**
- [ ] Per-reviewer full response included in report output (not just aggregated scores)
- [ ] Consolidation algorithm visible — weights, method, tie-breaking rules
- [ ] Reviewer disagreements highlighted (e.g., when one reviewer gives 3/10 and another gives 8/10)
- [ ] Each reviewer identified by model name

**Dependencies:** None

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

## FR-12: Customizable Prompt Properties

**Description:** Replace hardcoded prompt struct fields with a generic `Properties map[string]string` that supports arbitrary key-value metadata.

**User Story:** As a non-Azure team, I want to add custom metadata to my prompts (e.g., `framework: django`, `complexity: high`) so that I can filter and match criteria using my own taxonomy.

**Acceptance Criteria:**
- [ ] `Prompt` struct has `Properties map[string]string` field
- [ ] All filter flags (`--service`, `--language`, etc.) query properties map
- [ ] Existing Azure-specific fields (`Service`, `Plane`, etc.) become convenience accessors reading from properties
- [ ] All 87 existing prompts migrated to properties format (big-bang, no backward compat)
- [ ] Migration script provided and validated
- [ ] `KnownFields(true)` relaxed for properties section of frontmatter

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

## FR-14: Criteria Filters (Property-Based with Exclude)

**Description:** Replace hardcoded `MatchCondition` fields with property-based include/exclude matching. Remove Tier 1 (built-in default) criteria entirely.

**User Story:** As a team lead, I want to write criteria that apply to specific prompt categories and explicitly exclude others so that I have fine-grained control over evaluation standards.

**Acceptance Criteria:**
- [ ] `MatchCondition` replaced with `map[string]string` include/exclude matching
- [ ] Tier 1 criteria removed — no built-in defaults
- [ ] Criteria YAML supports `match` (include) and `exclude` property conditions
- [ ] When no criteria match a prompt, evaluation still runs but reports "no criteria matched"
- [ ] Existing criteria files migrated to property-based format

**Dependencies:** FR-12 (customizable properties — criteria reference properties)

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

## FR-17: Reviewer Tool Environments

**Description:** Configurable tool environments for review panel agents. Reviewers can have access to linters, style checkers, documentation references, and other evaluation tooling.

**User Story:** As an evaluator, I want review agents to use a Python linter when reviewing Python code so that the review catches style and correctness issues that pure LLM review might miss.

**Acceptance Criteria:**
- [ ] Config YAML supports `reviewer.tools` section (MCP servers, skills for reviewers)
- [ ] Reviewer tool environment is separate from generator tool environment
- [ ] Each reviewer model can have different tool configurations
- [ ] Reviewer tools are reflected in the review session config
- [ ] Review reports indicate which tools each reviewer had available

**Dependencies:** FR-05 (transparent review panel — tool availability is part of transparency)

**Phase:** 3

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

## Feature Dependency Map

```
FR-12 (Properties) ──┬── FR-14 (Criteria Filters)
                      ├── FR-11 (Tool Filters) ──── FR-15 (Pairwise)
                      ├── FR-02 (Config Tools) ──── FR-15 (Pairwise)
                      ├── FR-04 (.hyoka Dir)
                      └── FR-13 (YAML Format)

FR-18 (System Prompts) ── FR-03 (Zero System Prompt)
                       └── FR-01 (Action History)

FR-16 (Starter Files) ── FR-10 (Isolation)

FR-15 (Pairwise) ─┬── FR-07 (Comparison)
FR-01 (History)  ──┘

FR-05 (Review Panel) ── FR-17 (Reviewer Tools)
```

---

## Priority Order

| Priority | Feature | Rationale |
|----------|---------|-----------|
| 1 | FR-12 Properties | Foundation — everything depends on this |
| 2 | FR-18 System Prompts | Unblocks zero system prompt and action history |
| 3 | FR-03 Zero System Prompt | Core principle — unbiased measurement |
| 4 | FR-14 Criteria Filters | Needed for general use |
| 5 | FR-11 Tool Filters | Needed for pairwise |
| 6 | FR-16 Starter Files | Enables new prompt categories |
| 7 | FR-13 YAML Format | Enables programmatic prompt creation |
| 8 | FR-02 Config Tools | Foundation for pairwise |
| 9 | FR-08 Session Limits | Low effort, high value |
| 10 | FR-15 Pairwise | Core differentiator |
| 11 | FR-01 Action History | Core transparency |
| 12 | FR-05 Review Panel | Core transparency |
| 13 | FR-17 Reviewer Tools | Extends review quality |
| 14 | FR-10 Isolation | Production readiness |
| 15 | FR-09 Resource Efficiency | Production readiness |
| 16 | FR-07 Comparison | Primary insights output |
| 17 | FR-04 .hyoka Dir | Ecosystem / adoption |
| 18 | FR-06 Tool Marketplace | Ecosystem / adoption |
