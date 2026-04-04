# Project Context

- **Project:** hyoka — Go evaluation tool that runs prompts through the Copilot SDK, reviews code via a multi-model panel, produces criteria-based pass/fail reports
- **Stack:** Go 1.24.5+, GitHub Copilot CLI/SDK, MCP servers (Azure MCP via npx)
- **User:** Ronnie Geraghty
- **Created:** 2026-04-03
- **Repo:** /home/rgeraghty/projects/hyoka
- **Key paths:** hyoka/ (Go tool), prompts/ (evaluation prompts), criteria/ (pass/fail criteria), configs/ (run configs), reports/ (output), templates/, site/

## Core Context

Agent Morpheus initialized as Lead/Architect for hyoka. The project has guardrails for generation (turn count, file count, output size, session actions) and safety boundaries preventing real Azure resource provisioning by default. CLI supports list, run with filters (--service, --language), and smart path detection.

## Recent Updates

📌 Team initialized on 2026-04-03

## Learnings

Initial setup complete.

### Comprehensive Audit (2026-07-14)

**Codebase stats:** ~20K lines Go across 21 packages + 1329-line main.go. 264 test functions. 87 prompts. 8 configs. All tests pass, go vet clean.

**Architecture:** Clean layered design with eval→review→report pipeline. Good use of interfaces (CopilotEvaluator, Reviewer). Dependency graph is acyclic and well-scoped. Config normalization handles legacy→new format migration elegantly. Criteria system (Tier 2 attribute-matched + Tier 3 per-prompt) is well-designed for growth.

**Critical bug found:** Reviewer models grabbed from first config only (main.go:469-473), ignoring per-config reviewer model differences when running multi-config evaluations.

**Key concern — main.go monolith:** 1329 lines in single file. All CLI commands, flag definitions, path resolution, reviewer wiring, and skill installation in one function. Should be split into cmd/ package with per-command files.

**Error handling:** Generally excellent — errors wrapped with %w, propagated up the stack. A few discarded errors: referenceFiles (reviewer.go:352), starterFiles (copilot.go:83), filepath.Walk returns (multiple packages). Most are intentionally non-fatal but should be logged.

**Testing gaps:** 264 test functions but heavily concentrated in eval (57) and report (19). Key untested areas: pidfile package (no tests at all), clean/kill logic, progress display, serve HTTP handlers (functional but only basic routing tests), trends analysis. No integration tests that exercise the full pipeline.

**Open issues (10):** #84 skill loading dir, #82 skills events HTML, #78 Azure tool investigation, #77 local skill config, #75 filter deps from review, #73 embed CLI, #72 auth check, #71 session disconnect, #18 Waza comparison, #14 portability.

**Process lifecycle:** Well-designed signal handling with two-stage shutdown (SIGTERM→wait→SIGKILL). PID file tracking for orphan detection. HYOKA_SESSION env tag for process identification. Clean command handles both session state and process cleanup.

**Config system:** Legacy backward compat adds complexity but works. KnownFields(true) in YAML parsing is good for catching typos. Skill types (local/remote) validated at parse time.

**Documentation:** 8 docs in docs/ covering architecture, CLI reference, configuration, guardrails, contributing, prompt authoring, getting started, and an eval-tool-plan. README is comprehensive (430 lines). AGENTS.md matches actual structure. One stale reference in new-prompt: says "go run ./tool/cmd/hyoka validate" (should be "go run ./hyoka validate").

**Key file paths for future reference:**
- CLI entry: hyoka/main.go (monolith — needs splitting)
- Engine core: hyoka/internal/eval/engine.go (1035 lines)
- SDK integration: hyoka/internal/eval/copilot.go (813 lines)  
- Review panel: hyoka/internal/review/reviewer.go (732 lines)
- HTML report: hyoka/internal/report/html.go (1374 lines — largest file)
- Config: hyoka/internal/config/config.go (383 lines)
- Trends: hyoka/internal/trends/trends.go (857 lines)

### Deep Hardening Audit (2026-10-14)

**Delta from July audit:** Zero structural changes. All 10 issues still open. Reviewer model bug (main.go:469-473) still present. main.go still 1329 lines. pidfile still untested. Go module bumped to 1.26.1 but docs reference 1.24.5+.

**Critical new finding: No CI pipeline.** `.github/workflows/` has only squad orchestration — no `go build`, `go test`, or `go vet` in CI. Any PR can merge broken code.

**Config validation gap:** Empty `Generator.Model` passes validation but fails at runtime. Duplicate config names silently shadow — second config inaccessible.

**Serve security note:** `runID` in serve.go:171 not validated for directory traversal (unlike `relPath` on line 197). Low exploitability due to Go's URL normalization but inconsistent defense-in-depth.

**Error handling remains strong:** No `fmt.Errorf` missing `%w`, no log-and-return, no panics in production code. Only `reviewer.go:352` and `copilot.go:83` silently discard errors that should be logged.

**Test quality high but gaps remain:** 264 tests, 21/22 packages covered. Flaky `time.Sleep` in resourcemonitor_test.go. Review package thin (5 tests for 732 lines).

**Full hardening plan written to:** `.squad/decisions/inbox/morpheus-hardening-plan.md` with 20 prioritized tasks (3 P0, 10 P1, 7 P2).

**Decision:** CI is the single biggest hardening priority. Without it, every other fix can regress silently.

### Evolution Plan (2026-10-14)

**Scope:** Integrated hardening + product vision plan mapping Ronnie's 15 requirements against current architecture. Deep-dived into all 6 major subsystems: config, prompt, eval engine, review/criteria, serve/report, skills.

**Key architectural findings for the vision:**
- `Prompt` struct has 12 hardcoded fields + `KnownFields(true)` — blocks any non-Azure use. Must add `Properties map[string]string` and relax strict parsing.
- `MatchCondition` in criteria has 5 hardcoded fields (language, service, plane, category, sdk) — must generalize to property-based matching to support tool filters and criteria filters.
- System prompt (`copilot.go:628-655`) has 15 rules — many are bias, not operational. Must be reduced for "minimal system prompt" requirement.
- Pairwise testing is entirely new — needs new `pairwise/` package generating N+1 config variants per tool set.
- Comparison/insights capability partially exists (dashboard, trends) but is static — needs API-driven dynamic comparison engine.
- Per-project `.hyoka` directory is entirely new — needs auto-discovery and `hyoka init` command.
- YAML prompt format is entirely new — needs `ParsePromptYAML()` alongside existing markdown parser.

**Critical dependency chain:** CI (0.1) → Generic Properties (1.1) → Everything else. Properties are the single biggest model change because criteria filters, tool filters, comparison engine, and `.hyoka` portability all depend on prompts having arbitrary properties.

**Plan written to:** `.squad/decisions/inbox/morpheus-evolution-plan.md` with 5 phases, 25+ tasks, 6 open questions for Ronnie, risk assessment, and ~4-5 month timeline.

**6 open questions identified for Ronnie's input:**
1. Tier 1 criteria removal — keep as opt-in default or remove entirely?
2. System prompt scope — how minimal is "minimal"?
3. Pairwise testing granularity — full combinatorial or MCP-server-only?
4. Properties migration strategy — big-bang script or gradual backward-compat?
5. `.hyoka` vs global install — project-scoped only or both?
6. Response type — files only or text responses valid?

### Skills Investigation (2026-10-14)

**Finding:** `.squad/skills/` directory does not exist — zero project-specific skills. 29 generic template skills exist in `.squad/templates/skills/` but none encode hyoka's architecture, Go patterns, or domain knowledge. The `prompt-authoring` project skill at repo root covers prompt creation but not schema details.

**Recommended 14 skills across 4 categories:**
- Core Architecture (5): eval-pipeline, error-handling, config-system, copilot-sdk-integration, criteria-system
- Working Patterns (4): testing-patterns, cli-patterns, report-generation, logging-conventions
- Human Developer (2): contributor-guide, prompt-conventions
- Evolution Support (3): property-migration, process-lifecycle, serve-patterns

**Key insight:** The three highest-priority skills are `hyoka-eval-pipeline` (pipeline literacy prevents cross-package wiring bugs), `hyoka-error-handling` (strict %w wrapping convention is non-obvious to new agents), and `hyoka-contributor-guide` (reduces new-contributor ramp-up from hours to minutes).

**Template skills to copy/customize:** test-discipline, ci-validation-gates, reviewer-protocol, secret-handling (copy as-is); git-workflow (fork — hyoka uses `{username}/issue-{N}-{desc}` not `squad/{N}-{slug}`); project-conventions (replace with hyoka-contributor-guide).

**Recommendations written to:** `.squad/decisions/inbox/morpheus-skills-recommendations.md` with full details, priority matrix, and 3 open questions.

### Plan Documentation (2026-10-14)

**Task:** Created comprehensive `plan/` directory at repo root capturing the full evolution vision, separated from `docs/` (which documents the current tool).

**5 documents created:**
1. `plan/evolution-plan.md` (~25K chars) — Full 5-phase plan with 40+ tasks, dependency graph, team assignments, timeline, risk assessment. Incorporates all of Ronnie's Q1-Q6 answers and late additions (starter files, reviewer tools, configurable system prompts, zero system prompt).
2. `plan/core-principles.md` (~8K chars) — 10 guiding principles: transparency, unbiased measurement, tool impact focus, generality, isolation, resource responsibility, insights first, configurable not opinionated, dependency simplicity, skills as guardrails.
3. `plan/feature-requirements.md` (~18K chars) — 18 features as structured PRD with IDs (FR-01 through FR-18), user stories, acceptance criteria, dependencies, and phase assignments. Includes dependency map and priority ordering.
4. `plan/engineering-standards.md` (~11K chars) — 10 standard areas: error handling, logging, testing, CLI conventions, configuration, dependencies, code organization, CI, git workflow, reference patterns (Waza). All based on existing codebase patterns.
5. `plan/decisions-log.md` (~9K chars) — 15 indexed decisions from the session: hardening priorities, phase structure, Q1-Q6 answers, late additions, Waza reference, zero system prompt, skill philosophy, plan directory structure.

**Key learning:** When creating standalone reference documents, cross-reference specific file paths and line numbers from audit findings. This makes the documents actionable — someone can go straight from a requirement to the code that needs changing.

### Anchoring Bias Review (2026-10-15)

**Task:** Reviewed all 5 plan documents for anchoring bias — places where proposed solutions are shaped by current implementation rather than being the objectively best approach.

**3 significant anchoring biases found:**

1. **Review system** — Plan proposes iterating on the LLM-monolith Reviewer/PanelReviewer pattern. Waza has a pluggable grader architecture (12 types: code, prompt, file, program, behavior, etc.) that is fundamentally better. Current review stuffs everything into one LLM prompt and parses JSON. Should adopt grader architecture where LLM-as-judge is one grader type among many.

2. **Config legacy fields** — Plan does big-bang for 87 prompts but keeps backward compat for 8 config files (10 legacy fields, Normalize(), 7 Effective*() getters = ~130 lines / 35% of config.go). Should big-bang migrate configs too.

3. **Prompt Properties bolt-on** — Plan adds `Properties map[string]string` alongside existing typed fields (Service, Language, etc.), creating two representations. Should make Properties the sole source of truth with convenience getter methods.

**Areas confirmed as NOT anchored (good calls):** Zero system prompt, big-bang migration strategy, starter files/ResourceFile, .hyoka project dir, pairwise testing, isolated workspaces.

**Key learning:** Anchoring bias is most visible in "improve existing X" proposals. Ask "would someone designing this from scratch do it this way?" For hyoka, the review system is the clearest case — the plan proposes transparency and tools for the existing reviewer, but the right move is replacing it with a grader architecture that makes those features emergent.

**Output:** `plan/anchoring-review.md` — 6 findings, prioritized recommendations, plan section impact analysis, 3 open questions for Ronnie.

### Anchoring Review Pivot Integration (2026-10-15)

**Task:** Ronnie approved all 3 anchoring review pivots (grader architecture, config cleanup, run spec files) plus directed that properties-as-sole-representation (Finding #3) and criteria-tiers-to-grader-config (Finding #5) be absorbed into the grader decision. Updated all 5 plan documents.

**Changes made across 5 files:**
1. `plan/evolution-plan.md` — Added Phase 0 task 0.6 (config migration), renumbered 0.6-0.8→0.7-0.9. Rewrote §1.1 (Properties as sole repr.), §1.2 (Criteria→Grader Configs). Added §2.5 (Grader Architecture). Rewrote §3.2 (Grader Result Transparency). Updated Phase 4 for `[]GraderResult`. Added run spec future note. Updated dependency graph and all team assignments.
2. `plan/PRD.md` — Updated FR-05 (absorbed into graders), FR-12 (sole representation), FR-14 (grader configs), FR-17 (program grader). Added FR-19 (Grader Architecture). Updated dependency map and priority table.
3. `plan/core-principles.md` — Added Principle §11: "Deterministic Where Possible."
4. `plan/engineering-standards.md` — Rewrote §5 (Configuration): removed Normalize()/Effective*() rules, added direct field access and grader config to schema. Added §11 (Grader Interface Pattern) with rules, grader type table, and report types note. Updated §7 package list and interface references.
5. `plan/decisions-log.md` — Added D-AR1 (grader pivot), D-AR2 (config cleanup), D-AR3 (run spec future), D-AUTO (coordinator autonomy). Updated decision index.

**Key learning:** When a human approves architectural pivots, the ripple across documentation is significant — 5 files, ~30 edits. Having a structured plan directory with clear section numbering makes these updates tractable. The alternative (unstructured docs) would make this kind of cross-cutting update error-prone.

### Issue Tracking Integration (2026-10-15)

**Task:** Tank created 72 GitHub issues (#91–#162) mapping to all evolution plan tasks. Updated `plan/evolution-plan.md` to reference each task's issue number.

**Work done:**
1. Added comprehensive Issue Tracking section at top of evolution plan with phase-by-phase breakdown.
2. Added issue references to all 72 task entries (formats: `(#NNN)` in task descriptions, Phase 5.4 as integration task).
3. Committed with proper trailer: `Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>`

**Key learning:** Linking plan documents to issue tracking at the time of planning (not retroactively) makes the plan actionable and creates audit trail. Every task in the plan is now discoverable via GitHub issue search.

**Issue ranges by phase:**
- Phase 0 (Foundation): #91–#99 (9 issues — CI, bug fixes, validation)
- Phase 1 (Core Model): #100–#119 (20 issues — Properties, graders, tool filters, formats, refactor)
- Phase 2 (Evaluation Engine): #120–#137 (18 issues — Pairwise, limits, isolation, resources, graders)
- Phase 3 (Transparency): #138–#143 (6 issues — Action history, grader result display)
- Phase 4 (Insights & Comparison): #144–#151 (8 issues — Comparison engine, serve evolution, trends)
- Phase 5 (Ecosystem): #152–#162 (11 issues — .hyoka init, tool registry, tests, skills)

**Total: 72 issues across 5 phases**
