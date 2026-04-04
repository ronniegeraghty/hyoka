# Squad Decisions

## Active Decisions

### Decision: Plan Directory Created (2026-04-04)

**Author:** Morpheus 🕶️  
**Status:** Implemented

**What:** Created `plan/` directory at repo root with 5 comprehensive documents capturing the full evolution vision from the hardening session:

1. `plan/evolution-plan.md` — 5-phase plan, 40+ tasks, dependency graph, team assignments
2. `plan/core-principles.md` — 10 guiding principles
3. `plan/PRD.md` — 18 features as structured PRD (FR-01 through FR-18)
4. `plan/engineering-standards.md` — 10 engineering standard areas
5. `plan/decisions-log.md` — 15 indexed session decisions

**Why:** Separation of concerns — `docs/` documents the current tool while `plan/` captures the forward-looking vision. The evolution plan is now persistent and serves as the master task list for Phase 0–Phase 4 execution.

**Incorporated directives:**
- Ronnie's Q1-Q6 answers (Tier 1 removal, zero system prompt, pairwise flag, big-bang migration, project-scoped .hyoka, config-level response type)
- Reviewer tools addition
- Configurable system prompts (gen + review)
- Starter files (Waza ResourceFile pattern)
- Zero system prompt (Waza pattern)
- Skill philosophy (guardrails not cages)
- 14 skills recommendations

**Team impact:** All squad members should read `plan/evolution-plan.md` for their assigned tasks and `plan/engineering-standards.md` for coding conventions.

**Orchestration Log:** See `.squad/orchestration-log/2026-04-04T00-52-morpheus-plan-docs.md`

---

### Decision: Hyoka Evolution Plan — Hardening + Product Vision

**Date:** 2026-04-04  
**Author:** Morpheus 🕶️  
**Status:** Proposed  
**Summary:** Integrated 5-phase plan combining October 2026 audit P0–P2 fixes with product vision to evolve hyoka into a general-purpose AI agent benchmarking platform. Covers 25+ tasks across 5 squad members, includes dependency graph, and identifies 6 open questions for team consensus.

**Full Plan:** See `.squad/decisions/inbox/morpheus-evolution-plan.md` (39 KB, 5 phases, dependency graph, open questions)

---

### Decision: User Directives (2026-04-04)

**Date:** 2026-04-04  
**By:** Ronnie Geraghty (via Copilot)  
**Status:** Captured

#### 2026-04-04T00:08:37Z: Reviewer tools & configurable system prompts

**What:**
1. Review panel agents should be able to have tools added to their environments as well — not just the generation agent. This allows reviewers to reference specific evaluation tooling (e.g., linters, style checkers, documentation references).
2. The system prompt for BOTH the agent attempting the prompt AND the review agents should be configurable in the config YAML files. Users should control what system prompt is used, supporting the "minimal to no system prompt bias" goal.

**Why:** User request — additions to the hyoka product vision for the hardening/evolution effort. Integrated into Morpheus's Phase 3 work.

#### 2026-04-04T00:12:40Z: Skills investigation

**What:** Morpheus should investigate what agent skills we may want in the repo to help each squad member and human devs working on the project.

**Why:** User request — skills improve agent effectiveness and developer onboarding. Captured as Phase 5 research task in the evolution plan.

#### 2026-04-04T00:28:37Z: User directive — skill philosophy

**What:** Project-specific skills should be advisory, not prescriptive. They should NOT say "the core eval process should always work like this" because the project is evolving and we may want to change things. Instead, consider a skill that captures core principles and warns when work goes against them — a guardrail, not a cage.

**Why:** User request — the project is in active evolution (hardening + product vision). Rigid skills would block progress.

#### 2026-04-04T00:46:08Z: Ronnie's answers to evolution plan open questions

**By:** Ronnie Geraghty (via Copilot)

**Q1 — Tier 1 Criteria:** Option A. Remove entirely. Prompts/configs must supply their own criteria.

**Q2 — System Prompt:** Super minimal. Only isolation-related rules. Hardcoded guardrails (in code) are better than system prompt guardrails. If isolation can be achieved through SDK session config alone, don't put it in the system prompt at all. Make agent configs very transparent — keep them in a config file that gets loaded in.

**Q3 — Pairwise Testing:** `--pairwise` / `-pw` flag on the `run` command. When passed, it expands one config into the full set of pairwise eval variants. In the config YAML, tools should have an option to mark "not part of pairwise testing, always on" — so some tools are exempt from toggling.

**Q4 — Property Migration:** Big-bang. Update all prompts to the new format. No backward compatibility for old fields.

**Q5 — .hyoka Directory:** `.hyoka` only, project-scoped. Structured like a `.agents` dir with specific subdirs: `configs/`, `prompts/`, `criteria/`, etc. No global install mode.

**Q6 — Response Type:** This is something to specify in a config-specific system prompt. Look at microsoft/waza for how they handle agent eval system prompts. For text responses, need to think about how they get passed to review agents.

**NEW REQUIREMENT — Starter Files:** Core feature: ability to start the agent attempting the prompt in an environment with files already existing. Example: "I get an error when I try to build my code, can you fix it" — and we give the agent the failing code. This means prompts need a way to reference starter files that get placed in the agent's working directory before the session begins.

**Why:** User decisions on evolution plan open questions — these are binding direction for Phase 1+ implementation.

#### 2026-04-04T00:49:46Z: User directive — zero system prompt for agent sessions

**What:** Follow Waza's approach: zero system prompt for agent evaluation sessions. All configuration (working directory, tools, isolation) handled through SDK SessionConfig, not prompt injection. Config-specific custom system prompts remain an option in config YAML for users who want them, but the default is empty.

**Why:** User decision — system prompt biases agent behavior. The whole point of hyoka is measuring what agents do naturally with different tools. Injecting 15 rules defeats that purpose. Waza proves it works with zero system prompt.

#### 2026-04-04T00:52:00Z: User directive — plan directory for evolution docs

**What:** Create a `plan/` directory for documents related to the decisions and choices made during this hardening/evolution session. Existing `docs/` is documentation on how the tool currently works and should stay as-is. The plan dir captures the forward-looking vision, decisions, principles, and requirements.

**Why:** Separation of concerns — `docs/` = current state, `plan/` = future state. The evolution plan, core principles, PRD, and engineering standards belong in plan/ since they describe what hyoka is becoming, not what it is today.

---

### Decision: Recommended Skills for hyoka (2026-04-04)

**Date:** 2026-04-04  
**Author:** Morpheus 🕶️  
**Status:** Proposed  
**Summary:** 14 skills recommended across 4 categories to encode hyoka's architecture, Go patterns, domain conventions, and operational knowledge.

#### Skill Categories & Names

**Category 1: Core Architecture (All Agents + Human Devs)**
- `hyoka-eval-pipeline` — generate→review→report orchestration
- `hyoka-error-handling` — error wrapping, propagation, non-fatal logging
- `hyoka-config-system` — YAML loading, normalization, validation
- `copilot-sdk-integration` — session lifecycle, event handling, resource cleanup

**Category 2: Go Patterns & Conventions**
- `hyoka-criteria-system` — tiered evaluation, multi-level scoring
- `hyoka-testing-patterns` — test structure, table-driven tests, mocks

**Category 3: Subsystem Expertise (6 skills)**
- `hyoka-cli-patterns` — Cobra commands, flags, safety boundaries
- `hyoka-report-generation` — JSON/HTML/MD output, templates
- `hyoka-logging-conventions` — slog structured logging
- `hyoka-contributor-guide` — workflow, testing, iteration
- `hyoka-prompt-conventions` — frontmatter, validation, categorization
- `hyoka-property-migration` — legacy fields, idempotent normalization

**Category 4: Operational Knowledge**
- `hyoka-process-lifecycle` — session management, PID files, cleanup
- `hyoka-serve-patterns` — web handlers, path safety, report serving

#### Rationale

Each skill encodes patterns discovered during comprehensive codebase audits. Skills prevent agents from rediscovering architectural knowledge, Go conventions, and domain patterns during each task. All 14 should be created and published to `.squad/skills/` to enable effective asynchronous collaboration on hardening and evolution work.

**Full recommendations:** See `.squad/orchestration-log/2026-04-04T00-12-morpheus-skills.md` (orchestration log with detailed rationale, audiences, file references for each skill).

---

### Decision: Hardening Audit — October 2026 (Integrated)

**Date:** 2026-10-14  
**Author:** Morpheus 🕶️  
**Status:** Proposed  
**Scope:** Full codebase audit — hardening for production reliability

#### Executive Summary

The codebase has not changed structurally since the July 2026 audit. All 10 previously identified issues remain open. The reviewer model bug (P0) is still present. main.go is still 1329 lines. pidfile still has zero tests. **No CI pipeline exists for build/test** — only squad orchestration workflows. The Go module has been bumped to 1.26.1 but docs still reference 1.24.5+.

The good news: build passes clean, go vet clean, all 264 tests pass across 21 packages. Error handling remains excellent with no `fmt.Errorf` missing `%w`, no log-and-return antipatterns, and no panics in production code. The dependency footprint is minimal (3 direct deps). Safety boundaries for cloud access are properly implemented.

---

#### Phase 2: Area-by-Area Assessment

##### 1. Error Handling — 🟢 Production-Ready

**Strengths:**
- All `fmt.Errorf` calls use `%w` for proper error wrapping
- No log-and-return antipattern anywhere
- No panics in non-test code
- `os.Exit` calls are all appropriate (main entry, validation commands, emergency signal handler)

**One warning:** `review/reviewer.go:352` silently discards `ReadDirFilesFiltered` error for reference files. Review proceeds with empty refs, potentially degrading review quality without any indication.

**Multiple `filepath.Abs` errors discarded:**
- `main.go:219, 263, 270, 286` — `abs, _ := filepath.Abs(...)`
- `skills/fetcher.go:68, 82, 88, 120` — same pattern
- These almost never fail, but defense-in-depth says log on error.

##### 2. Configuration — 🟡 Needs Work

**Strengths:**
- `KnownFields(true)` catches YAML typos ✅
- Skill type/field validation is thorough ✅
- Legacy config migration is robust and idempotent ✅
- Missing config names produce clear errors ✅

**Gaps:**
- **No generator model validation** — a config with empty `Generator.Model` passes all validation but fails at runtime. This is a silent footgun.
- **Duplicate config names not detected** — two configs with the same `name:` field silently shadow each other; only the first is accessible.
- **No "did you mean?" suggestions** — typo in `--config` just says "not found" without listing available configs.
- **Skill paths not validated to exist** — bad `path` passes parse but fails at runtime.

##### 3. Process Lifecycle — 🟡 Needs Work

**Strengths:**
- Two-phase SIGTERM→SIGKILL shutdown with 5s timeout ✅
- Proper mutex usage in ProcessTracker ✅
- Deferred cleanup ensures processes terminated even on panic ✅
- PID file management is idempotent and cross-platform ✅

**Concerns:**
- **PID reuse risk** — PID files store PID + metadata but `kill` doesn't validate the metadata matches. If PID is recycled, wrong process gets killed. Low probability on 64-bit Linux but real on 32-bit or busy systems.
- **No session lock files** — `clean` command can remove sessions that are currently in-use by another hyoka instance.
- **Heuristic session detection** — `isHyokaSession()` uses string matching for "hyoka", "reports/", etc. Could miss sessions or falsely match non-hyoka sessions.

##### 4. CLI UX — 🟢 Production-Ready

**Strengths:**
- Help text is clear and comprehensive for all commands ✅
- `--dry-run` works correctly ✅
- Safety boundary is opt-in (`--allow-cloud`) with sensible default ✅
- Good flag documentation with defaults shown ✅

**One bug:** `new-prompt` command (line 1276) prints `go run ./tool/cmd/hyoka validate` — should be `go run ./hyoka validate`. Stale path from early prototype.

##### 5. Testing — 🟡 Needs Work

**Strengths:**
- 21/22 packages have tests (95% coverage) ✅
- 264 test functions across 29 test files ✅
- Engine tests are excellent — guardrails, timeouts, multi-config ✅
- Config tests are thorough — edge cases, backward compat ✅
- Serve tests include path traversal checks ✅

**Gaps:**
- **pidfile package: ZERO tests** — cross-platform process alive detection, PID file CRUD, and stale cleanup all untested.
- **No integration tests** — no test exercises generate→review→report end-to-end. Intentional (needs LLM), but a stub-based integration test would catch wiring regressions.
- **Flaky tests in resourcemonitor_test.go** — `time.Sleep(100ms)` and `time.Sleep(120ms)` for timing assertions. Will fail under load.
- **Review package thin** — only 5 tests in review_test.go for 839 lines of production code.

##### 6. Code Quality — 🟡 Needs Work

**main.go monolith (1329 lines):** Still the single biggest maintenance burden. All 14 cobra commands, flag definitions, path resolution, reviewer wiring, skill installation, and prompt scaffolding in one file. The `runCmd()` function alone is ~300 lines.

**Large files by package:**
| File | Lines | Concern |
|------|-------|---------|
| `report/html.go` | 1374 | HTML built via string concatenation. Templates would be better. |
| `main.go` | 1329 | Monolith. Should be split into per-command files. |
| `eval/engine.go` | 1035 | Acceptable — it's the core orchestrator. |
| `trends/trends.go` | 857 | Complex but cohesive. |
| `eval/copilot.go` | 813 | SDK integration — necessarily complex. |
| `review/reviewer.go` | 732 | Could extract panel logic. |

**No dead code found.** No `//nolint` suppressions. No `FIXME`/`TODO`/`HACK` comments (only template TODOs in prompt scaffolding, which is correct).

##### 7. Build & CI — 🔴 Blocking

**Build:** `go build ./hyoka/...` passes clean. `go vet ./hyoka/...` passes clean. ✅

**CI: NO BUILD OR TEST CI EXISTS.** The `.github/workflows/` directory contains only:
- `squad-heartbeat.yml` — squad orchestration
- `squad-issue-assign.yml` — issue assignment
- `squad-triage.yml` — issue triage
- `sync-squad-labels.yml` — label sync

**There is no workflow that runs `go build`, `go test`, or `go vet` on PRs or pushes.** This means regressions can be merged without any automated safety net.

##### 8. Documentation — 🟡 Needs Work

**Stale references:**
- `docs/getting-started.md:9` — says Go 1.24.5+, should be 1.26.1+
- `docs/contributing.md:5` — says Go 1.24.5+, should be 1.26.1+
- `README.md:9` — says Go 1.24.5+, should be 1.26.1+
- `AGENTS.md:61` — says Go 1.24.5+, should be 1.26.1+
- `main.go:1276` — says `go run ./tool/cmd/hyoka validate`, should be `go run ./hyoka validate`

**Doc completeness:** 9 docs covering architecture, CLI, config, guardrails, contributing, prompt authoring, getting started, eval plan, and cleanup plan. Good breadth.

##### 9. Security/Safety — 🟢 Production-Ready (with caveat)

**Strengths:**
- Cloud access boundary properly implemented (`--allow-cloud` opt-in) ✅
- Path traversal protection in serve handlers (`filepath.Clean` + `..` check) ✅
- Serve path traversal test exists (`TestAPIEvalTraversalBlocked`) ✅
- No credential handling in code (delegates to Copilot SDK) ✅
- Guardrails (turn limits, file limits, output size) properly enforced ✅

**Caveat:** In `serve.go:171`, `runID` is extracted from URL path but NOT validated for traversal. `filepath.Clean("..") = ".."`. While Go's HTTP server normalizes URL paths (removing `..` before dispatch), defense-in-depth says `runID` should be validated too. Low exploitability due to Go's URL normalization, but the `relPath` parameter on line 197-200 sets a precedent that `runID` doesn't follow.

##### 10. Dependencies — 🟢 Production-Ready

**Direct deps (3):**
- `github.com/github/copilot-sdk/go v0.2.0` — core requirement
- `github.com/spf13/cobra v1.10.2` — CLI framework, stable
- `gopkg.in/yaml.v3 v3.0.1` — YAML parsing, stable

**Indirect deps (10):** All from copilot-sdk (logr, otel, uuid, jsonschema). Reasonable transitive footprint.

**Go version:** 1.26.1 — current.

**No known vulnerabilities** in direct dependencies (all widely-used, maintained packages).

---

#### Phase 3: Prioritized Hardening Tasks

##### P0 — Must Fix Before Real Use

| # | Issue | File:Line | Why It Matters | Owner | Size |
|---|-------|-----------|---------------|-------|------|
| 1 | **Reviewer model bug: first-config-wins** | `main.go:469-473` | Multi-config evals silently use wrong reviewer panel. Results are incorrect without any error. | **Neo** | Small |
| 2 | **No CI pipeline for build/test** | `.github/workflows/` | Any PR can merge broken code. Zero safety net. | **Tank** | Medium |
| 3 | **Fix stale path in new-prompt** | `main.go:1276` | Tells users to run a command that doesn't exist. | **Tank** | Small |

##### P1 — Should Fix Soon

| # | Issue | File:Line | Why It Matters | Owner | Size |
|---|-------|-----------|---------------|-------|------|
| 4 | **Add generator model validation** | `config/config.go:256-287` | Empty `Generator.Model` passes validation but fails at runtime. Users get a cryptic SDK error instead of a clear config error. | **Tank** | Small |
| 5 | **Split main.go into per-command files** | `main.go` (1329 lines) | Every change touches the same file. Merge conflicts, cognitive overload, hard to review. | **Tank** | Large |
| 6 | **Add pidfile tests** | `internal/pidfile/` | Only untested package. Cross-platform process detection is error-prone and needs coverage. | **Switch** | Medium |
| 7 | **Add stub integration test** | (new file) | No test exercises generate→review→report wiring. A stub-based e2e test catches regressions in the assembly layer. | **Switch** | Medium |
| 8 | **Log discarded errors** | `reviewer.go:352`, `copilot.go:83`, `main.go:219,263,270,286`, `fetcher.go:68,82,88,120` | Silent failures degrade output quality without any diagnostic trail. | **Neo** | Small |
| 9 | **Fix Go version in docs** | `getting-started.md`, `contributing.md`, `README.md`, `AGENTS.md` | Says 1.24.5+ but go.mod requires 1.26.1. Users get a confusing build error. | **Oracle** | Small |
| 10 | **Detect duplicate config names** | `config/config.go:256-287` | Two configs with the same name silently shadow. Second config is invisible. | **Tank** | Small |
| 11 | **Validate runID in serve handler** | `serve/serve.go:171` | `runID` from URL not checked for traversal. Low exploitability due to Go URL normalization, but inconsistent with `relPath` defense on line 197. | **Trinity** | Small |
| 12 | **Fix flaky resourcemonitor tests** | `eval/resourcemonitor_test.go` | `time.Sleep(100ms)` assertions fail under load. Replace with event-driven checks. | **Switch** | Small |
| 13 | **Add early auth check** | `main.go` (near line 454) | Issue #72. Auth failures discovered late after config/prompt processing. Call `GetAuthStatus()` upfront. | **Neo** | Small |

##### P2 — Nice to Have

| # | Issue | File:Line | Why It Matters | Owner | Size |
|---|-------|-----------|---------------|-------|------|
| 14 | **Extract HTML templates from html.go** | `report/html.go` (1374 lines) | String concatenation for HTML is fragile and hard to maintain. Embed template files. | **Trinity** | Large |
| 15 | **Add "did you mean?" for config names** | `config/config.go:307-314` | Better UX when users typo config names. Levenshtein distance suggestion. | **Tank** | Small |
| 16 | **Call session.Disconnect() before DeleteSession** | `eval/copilot.go`, `review/reviewer.go` | Issue #71. Match SDK's intended two-phase teardown pattern. | **Neo** | Small |
| 17 | **Embed Copilot CLI binary** | (new package) | Issue #73. Eliminates CLI version skew, setup friction, shared state. | **Neo** | Large |
| 18 | **Add PID birth-time validation** | `pidfile/pidfile.go`, `clean/clean.go` | Prevents killing wrong process on PID reuse. Store start time, validate before kill. | **Neo** | Medium |
| 19 | **Review package needs more tests** | `review/reviewer.go` (732 lines, 5 tests) | Low test-to-code ratio for a critical package. | **Switch** | Medium |
| 20 | **Remove legacy config fields** | `config/config.go` | After deprecation period. Dual-path adds maintenance burden. | **Tank** | Medium |

---

#### Phase 4: Team Knowledge Updates

##### Neo (eval engine, review pipeline)
- Reviewer model bug at main.go:469-473 is the highest-priority fix. The `break` on line 472 means all configs share one reviewer panel.
- `reviewer.go:352` silently discards reference file read errors — add `slog.Warn`.
- `copilot.go:83` same pattern with starter files.
- Session.Disconnect() should be called before DeleteSession (issue #71).

##### Tank (CLI, config, environment)
- **P0:** Create a CI workflow with `go build`, `go test`, `go vet` on PR/push.
- **P0:** Fix stale path at main.go:1276 (`./tool/cmd/hyoka` → `./hyoka`).
- Config validation should reject empty `Generator.Model` and detect duplicate names.
- main.go split is the biggest maintainability win — propose `hyoka/cmd/` package with per-command files.

##### Switch (testing, CI)
- pidfile is the only zero-test package. Needs: Write/Remove/ReadAlive, cross-platform alive check, stale cleanup.
- Stub integration test: StubEvaluator + StubReviewer → engine.Run() → verify report output exists and is valid.
- resourcemonitor_test.go has flaky `time.Sleep` assertions — replace with channel-based or polling checks.
- review package has only 5 tests for 732 lines — needs more coverage.

##### Trinity (reports, templates, serve)
- Validate `runID` in serve.go:171 for directory traversal consistency.
- html.go (1374 lines) is the largest file — extracting to embedded templates would improve maintainability.
- Skills events not in HTML reports (issue #82) — coordinate with Neo on event data flow.

##### Oracle (documentation)
- Go version references in 4 files need updating from 1.24.5+ to 1.26.1+.
- `main.go:1276` stale path needs fixing (overlaps with Tank's fix).
- docs/ are otherwise accurate and comprehensive.

---

#### Changes Since Last Audit (July 2026)

**What changed:**
- Go module bumped from 1.24.5 to 1.26.1
- Some commits for dependency filtering (#75), strict YAML parsing, action limits refactor, Windows support, process tracking improvements
- Multiple bug fixes (process scoping, excluded dirs matching, orphan scanning)

**What did NOT change:**
- main.go still 1329 lines (not split)
- Reviewer model bug still present (main.go:469-473)
- Stale path still present (main.go:1276)
- pidfile still has zero tests
- All 10 open issues still open
- No CI pipeline added
- No integration tests added

**Net assessment:** The codebase is well-built but hasn't been hardened. The architecture is sound, error handling is strong, and the dependency footprint is minimal. But the same issues flagged in July remain. The biggest gap is the complete absence of CI — that's a P0 for any production use.

#### Rationale

The codebase is in surprisingly good shape for its maturity stage. The architecture is clean, the dependency graph is acyclic, error handling is mostly proper, and test coverage exists for every package except pidfile. The main risks are: (1) the reviewer model bug silently producing wrong results, (2) the main.go monolith slowing iteration, (3) the complete absence of CI for build/test, and (4) the lack of end-to-end integration tests.

## Governance

- All meaningful changes require team consensus
- Document architectural decisions here
- Keep history focused on work, decisions focused on direction

---

### Decision: Anchoring Review Decisions + Autonomy Directive (2026-04-04T02:48:44Z)

**By:** Ronnie Geraghty (via Copilot)  
**Status:** Binding

**What:**

1. **AUTONOMY:** Squad coordinator should have a lower bar for what decisions require Ronnie's input. Make good decisions autonomously — don't be too eager to ask.
2. **Q1 (Grader architecture):** YES — adopt Waza's pluggable grader model. Replace Reviewer/PanelReviewer with Grader interface and typed grader implementations.
3. **Q2 (Config cleanup timing):** YES — big-bang migrate configs in Phase 0 alongside CI. Delete Normalize() and Effective*() getters.
4. **Q3 (Run spec files):** YES — explore `hyoka run eval.yaml` pattern as future enhancement. Don't block current work on it.

**Why:** User decisions on anchoring review findings. These are binding architectural pivots.
# Decision: GitHub Issue Linking for Evolution Plan

**Date:** 2026-10-15  
**Author:** Morpheus (Lead/Architect)  
**Status:** Documented  

## Problem

Evolution plan tasks (72 across 5 phases) exist as plan entries with no direct traceback to GitHub issues, making it difficult to:
- Link plan work to project tracking
- Discover task scope from GitHub issue search
- Cross-reference plan→issue→PR→code during development

## Decision

**Link all plan tasks to GitHub issues at planning time, not retroactively.**

Every task entry in `plan/evolution-plan.md` now references its GitHub issue number using the format `(#NNN)` directly in the task description or title. A comprehensive Issue Tracking section at the top of the plan document provides phase-by-phase breakdowns.

### Rationale

1. **Actionability:** Issue numbers make plan entries clickable and searchable on GitHub. Engineers can jump directly from plan to issue.
2. **Single source of truth:** The plan document becomes the authoritative task list; issues are the execution tracking mechanism.
3. **Audit trail:** Commit messages and plan updates can reference issues; code reviews can close issues when tasks land.
4. **Phase visibility:** Breaking out issue counts by phase (9, 20, 18, 6, 8, 11) makes sprint planning easier.

### Format

**In-line format:**
```
| 0.1 | **Create CI pipeline** (#91) — `go build`, ... | Tank | Medium | — |
```

**Grouped summary format (top of plan):**
```
Phase 0 (Foundation): #91–#99 (9 issues)  
Phase 1 (Core Model): #100–#119 (20 issues)  
...
```

## Outcome

- All 72 tasks now reference issues #91–#162
- Plan document updated and committed
- Team can navigate plan→GitHub issue→PR→code seamlessly
- Sprint planning can count issues per phase

## Team Note

When creating plan tasks in the future, wait for Tank/Ronnie to create the GitHub issues, then link them immediately in the plan document. This prevents orphaned tasks and keeps documentation in sync with project tracking.
### 2026-04-04T03:55:26Z: Design Meeting — Evolution Plan Review

**Participants:** Morpheus 🕶️ (facilitator), Neo 💊, Tank 📡, Trinity 🖤, Switch 🤍, Oracle 🔮
**Requested by:** Ronnie Geraghty

---

## Meeting Summary

Five domain reviews converged on a clear picture: the evolution plan is structurally sound but under-specified in three critical areas — the `GraderResult` type design, the boundary between typed fields and the `Properties` map, and the testing investment required for safe migration. Every reviewer independently identified dependencies and ordering constraints that the flat task list obscures, and several found tasks that are missing entirely (reviewer system message, report schema versioning, template extraction, security sanitization). The consensus is that Phase 0 needs to be strengthened before Phase 1 begins — CI with race detection, pidfile tests, and a serve handler security fix should all land before the model-change work starts.

Neo's eval engine review was the deepest technically, exposing that the reviewer model bug is worse than described (engine-scoped when it should be task-scoped), that `Properties map[string]string` as written would lose type safety on non-string fields like `Tags []string` and `Timeout int`, and that `GraderResult.Details interface{}` will be a pain point in Go templates and test assertions. Trinity's report review revealed that `SessionEventRecord` and `TimelineStep` already exist — the timeline work (3.1c) is largely a no-op for JSON, and the React SPA (not Go templates) is the right surface for the Phase 4 dashboard. Switch's testing review is the most actionable: 44 tests break in Phase 1 (not "mostly mechanical"), a serve handler path traversal vulnerability exists today, and pidfile — a safety feature — has zero tests. Tank confirmed the config migration is safe as a single PR and proposed promoting D-AR3 (run spec file) to Phase 2, which needs Ronnie's approval. Oracle identified that a 72-task plan has only 1 documentation task, recommending 12 distributed across phases.

---

## Consensus Points

These are areas where reviewers independently reached the same conclusion. These are **confirmed decisions** going forward:

1. **CI from day one is non-negotiable.** Tank and Switch agree: `go build`, `go vet`, `go test` with `-race` on every PR. Single Go version (1.26.1), single OS. CI (0.1) is an explicit blocker for all Phase 1 work.

2. **Convenience getters are essential.** Neo, Tank, and Oracle all assume `p.Language()` instead of `p.Properties["language"]`. The properties migration must include getter methods.

3. **System prompt removal must be phased, not big-bang.** Neo recommends: bias rules (9,10) first → guidance rules (1,2) → path rules (3-7). Everyone agrees that 1.6 must come after 0.2 (reviewer bug fix) is merged.

4. **`file` grader should be built first** to validate the `GraderResult` type before implementing more complex graders. Neo and Trinity both identified that the type design is the highest-risk decision in Phase 2.

5. **Schema versioning is mandatory for reports.** Trinity identified that `ReviewResult → GraderResult` changes the scoring model (int → float64). Without `schema_version`, old reports break `rerender`. No dissent.

6. **Template extraction must precede grader display work.** Trinity says html.go will exceed 1600 lines. Extract to `.gohtml` files with `embed.FS` before Phase 3.2a begins.

7. **Testing investment in Phase 1 is significantly underestimated.** Switch quantified: ~8 prompt assertion tests, ~6 filter tests, ~12 config tests, 6 criteria tests, and 12 main tests need updating. Plan should acknowledge this.

8. **Documentation is underrepresented.** Oracle found 1 explicit doc task in a 72-task plan. Feature owners draft docs; Oracle polishes. Breaking changes need migration guides before code lands.

9. **Config big-bang migration (0.6) is safe.** Tank confirmed: 6 of 8 configs use legacy format, ~17 call sites to update, test coverage is excellent. Single PR.

10. **React SPA is the right surface for Phase 4 dashboard.** Trinity confirmed: static HTML reports stay Go-templated, interactive comparison/drill-down goes in the existing React SPA (Vite + React + Radix + Recharts already present).

---

## Decisions Made (Coordinator Authority — D-AUTO)

### D-AUTO-DM1: Reviewer bug fix — per-task reviewer creation

The reviewer must be created per-task in `runSingleEval()`, not shared across the engine. This aligns with the grader direction where each evaluation task assembles its own grader set. Neo's option (b) — reviewer factory function — is the correct approach.

**Rationale:** Engine-scoped reviewer means multi-config runs silently use the wrong reviewer panel. Per-task creation is the minimal correct fix and sets the pattern for the grader architecture.

### D-AUTO-DM2: Properties map is metadata-only — typed fields retained for non-string data

`Properties map[string]string` replaces the hardcoded Azure-specific metadata fields (`Service`, `Language`, `Plane`, `Category`, `Difficulty`, `SDKPackage`, `DocURL`). The following fields remain typed on the `Prompt` struct:
- `Tags []string`
- `ExpectedPkgs []string`
- `ExpectedTools []string`
- `Timeout int`
- `StarterProject string`
- `ProjectContext map[string]string`
- `ReferenceAnswer string`

**Rationale:** `map[string]string` cannot represent `[]string` or `int` without lossy serialization. These fields have semantic meaning beyond key-value metadata. Convenience getters (`p.Language()`, `p.Service()`) read from the Properties map for the fields that moved.

### D-AUTO-DM3: Add `gate: bool` to grader config schema

Grader configs support an optional `gate: bool` field. When `gate: true`, a failing grader causes the entire evaluation to fail regardless of weighted average. Use cases: file exists, builds successfully, no forbidden tools used.

**Rationale:** Weighted averaging alone cannot express hard constraints. A program that doesn't compile should not pass evaluation just because LLM review scores were high. This is a small schema addition with large correctness impact.

### D-AUTO-DM4: GraderResult uses typed optional fields, not `interface{}`

Replace `Details interface{}` on `GraderResult` with typed optional fields per grader kind:
```go
type GraderResult struct {
    Kind       string
    Name       string
    Score      float64
    Weight     float64
    Pass       bool
    Gate       bool      // hard pass/fail gate
    // Typed details (only one populated per result)
    FileDetails    *FileGraderDetails
    ProgramDetails *ProgramGraderDetails
    PromptDetails  *PromptGraderDetails
    BehaviorDetails *BehaviorGraderDetails
    // ...
}
```

**Rationale:** `interface{}` requires type switches everywhere — templates, tests, serialization. Typed optional fields are explicit, discoverable, and template-friendly. Neo and Trinity both flagged this independently.

### D-AUTO-DM5: GraderInput is a concrete struct, not an interface

`GraderInput` must be a concrete struct containing everything a grader might need: session workspace path, action log, prompt metadata, config, file listing. Graders use what they need and ignore the rest.

**Rationale:** An `interface{}` or generic interface adds abstraction without value. All graders operate on the same evaluation output. A concrete struct is simpler, testable, and doesn't require type assertions.

### D-AUTO-DM6: Phase system prompt removal incrementally

System prompt removal (1.6b-c) follows this order:
1. Remove bias rules (9: research restrictions, 10: Python rules) — these are explicitly identified as behavioral bias
2. Remove guidance rules (1-2) — less critical than bias but still prescriptive
3. Remove path/operational rules (3-7, 8) — only after confirming SDK session config handles isolation
4. Remove safety boundaries (11-15) — move to code-level hooks first

**Rationale:** Big-bang removal risks breaking agent sessions if SDK config doesn't fully handle isolation. Incremental removal lets us verify each category independently.

### D-AUTO-DM7: Add explicit task for reviewer hardcoded system message

Add task **1.6d**: "Remove hardcoded system message from reviewer (`reviewer.go:180-183`). Make reviewer system prompt configurable via config YAML `reviewer.system_prompt` field." Assigned to Neo, size Small. This is part of the 1.6 system prompt work.

**Rationale:** Neo identified that the plan addresses the generator's system prompt but not the reviewer's. The reviewer has its own hardcoded message that must also be addressed for zero-system-prompt to be complete.

### D-AUTO-DM8: CI pipeline specification

Phase 0 CI (task 0.1) uses:
- Go 1.26.1 (matches go.mod), single OS (ubuntu-latest)
- `go build ./hyoka/... && go vet ./hyoka/... && go test -race ./hyoka/... -timeout 2m`
- Skip `golangci-lint` in Phase 0, add in Phase 1
- `-race` flag from day one (concurrent code in ResourceMonitor, ProcessTracker, PanelReviewer)
- 2-minute timeout (tests run in ~20s with -race, 2 min gives headroom)

**Rationale:** Switch correctly identified that concurrent code needs race detection from the start. Tank's 5-minute timeout is unnecessarily generous — 2 minutes is 6× the actual runtime.

### D-AUTO-DM9: Move pidfile tests from Phase 5 to Phase 0

Task 5.3a (pidfile tests) becomes **task 0.10**: "Add pidfile package tests — 136 lines, zero tests, safety-critical code." Assigned to Switch, size Small (~30 minutes per Switch's estimate).

**Rationale:** A safety feature with zero tests is a Phase 0 concern, not a Phase 5 nice-to-have. This is a 30-minute task that eliminates a gap in safety-critical code.

### D-AUTO-DM10: Move review package coverage from Phase 5 to Phase 1

Task 5.3c (review package coverage) moves to Phase 1 as **task 1.8**: "Increase review package test coverage (5 tests for 840 lines)." Must complete before Phase 2 replaces the reviewer.

**Rationale:** Switch's logic is correct — test the code before you replace it. Phase 2 replaces `review/` with `graders/`. If we don't test the reviewer in Phase 1, we lose the ability to verify the `prompt` grader faithfully wraps the old behavior.

### D-AUTO-DM11: Add schema_version to EvalReport

Task 2.5g gains an explicit requirement: add `SchemaVersion int` field to `EvalReport`. Version 1 = current format (ReviewResult-based). Version 2 = grader-based ([]GraderResult). Rerender checks schema version and handles both.

**Rationale:** Without versioning, the hundreds of existing reports in `reports/` will break when the report format changes. Trinity's recommendation is essential for backward compatibility.

### D-AUTO-DM12: Extract templates before Phase 3.2a

Add task **3.0** (pre-Phase 3): "Extract HTML templates from html.go string concatenation into `.gohtml` files using `embed.FS`." Assigned to Trinity, size Medium. This is a prerequisite for 3.2a.

**Rationale:** html.go is already 1374 lines of string concatenation. Adding per-grader display components without extraction would push it past 1600 lines. Template extraction is a necessary refactor that makes all Phase 3 template work cleaner.

### D-AUTO-DM13: Property key naming convention — enforce snake_case

All property keys in prompt frontmatter use `snake_case`: `service`, `language`, `data_plane`, `sdk_package`. Validation rejects keys with hyphens or camelCase. Migration script (1.1b) normalizes existing keys.

**Rationale:** Consistency matters for property-based matching. `snake_case` is the most common YAML convention and matches Go struct tag conventions.

### D-AUTO-DM14: Freeze GraderInput/GraderResult types before implementing graders

Task 2.5a (Grader interface design) must be fully reviewed and approved before 2.5b-f begin. The `file` grader (2.5b) serves as the type validation — it's the simplest grader and will expose any design issues in GraderInput/GraderResult before more complex graders are built.

**Rationale:** Neo correctly identifies GraderResult as the most expensive type to fix later. Every grader implementation, every report consumer, and every template depends on it. Get it right before building on it.

### D-AUTO-DM15: Tasks 0.2 and 0.6 are independent — no ordering constraint

The reviewer bug fix (0.2) and config migration (0.6) can proceed in parallel. 0.2 fixes a logic bug in reviewer scoping. 0.6 migrates config file format. They touch different code paths.

**Rationale:** Tank asked if 0.2 should wait for 0.6. The answer is no — 0.2 is a P0 correctness bug and should not be blocked by a cleanup task. However, config migration (0.6) and property migration (1.1b) MUST NOT overlap — Neo's hidden constraint is valid.

### D-AUTO-DM16: `.hyoka` walk-up discovery stops at Git root

The `.hyoka` directory search walks up from CWD and stops at the Git repository root (detected by `.git/` directory). It does not escape the repository boundary.

**Rationale:** Tank's recommendation. Walking past Git root would find unrelated `.hyoka` directories from parent projects, creating confusing behavior.

### D-AUTO-DM17: 0.1 (CI) is explicit blocker for all Phase 1 work

No Phase 1 PR merges until CI is green and enforced. The dependency graph already shows this but it was not explicitly stated as a hard gate.

**Rationale:** Tank requested this be explicit. The entire point of Phase 0 is establishing the safety net. Merging model changes without CI defeats the purpose.

### D-AUTO-DM18: Add 12 documentation tasks distributed across phases

Oracle's recommendation is adopted. Feature owners draft documentation; Oracle reviews and polishes. Breaking changes (Phase 1 properties migration, Phase 2 grader architecture) require migration guides published BEFORE code lands. Specific doc tasks to be enumerated by Oracle during Phase 0.

**Rationale:** A 72-task plan with 1 doc task is a documentation debt timebomb. Pairing docs with features ensures they stay current.

### D-AUTO-DM19: `prompt` grader instances each run ONE model

Each `prompt` grader config specifies a single model. To get multi-model review, configure multiple `prompt` grader instances with different models. Aggregation is handled by the grader framework's weighted scoring, not by internal panel logic.

**Rationale:** Trinity asked this question. Single-model-per-grader is simpler, more composable, and more transparent than hiding panel logic inside the grader. The user explicitly configures each reviewer model as a separate grader instance with its own weight.

### D-AUTO-DM20: `newPromptCmd` must update after Phase 1.1

Tank noted that `newPromptCmd` scaffolds old prompt format. Add to task 1.1b (migration script): also update the `new-prompt` command scaffold template to emit the new `properties:` format.

**Rationale:** Without this, every new prompt created after migration would use the old format and immediately fail validation.

---

## Decisions Requiring Ronnie's Input

### ESC-1: Serve runID path traversal fix — add to Phase 0?

**Issue:** Switch identified that `handleAPIRunDetail` in `serve.go` takes `runID` from the URL without sanitization. The path `/api/runs/../../etc/eval?path=passwd` could bypass directory checks. This is a **security vulnerability**.

**Options:**
- (a) Add as task 0.11 in Phase 0 — fix immediately before any other work
- (b) Fix as a hotfix PR outside the evolution plan, don't count as a plan task
- (c) Defer to Phase 1 (Switch recommends against this)

**Recommendation:** Option (b) — hotfix PR immediately, outside the plan's phase structure. Security issues don't wait for sprint planning. Switch can implement the fix and serve handler tests in a single PR.

### ESC-2: Promote D-AR3 (run spec file) from "future" to Phase 2?

**Issue:** Tank argues that `runCmd` already has 33 flags and Phase 2 adds more (pairwise, session limits). A declarative `hyoka run eval.yaml` pattern would absorb most flags.

**Options:**
- (a) Promote to Phase 2 as Tank recommends — adds ~1 Medium task
- (b) Keep as "future" — do the main.go split (1.5) first, revisit after Phase 2
- (c) Add to Phase 3 as a bridge between CLI simplification and dashboard

**Recommendation:** Option (b) — keep as future. The main.go split (1.5) is already Phase 1 work. Run spec files are a significant design investment. Let the split settle, let the grader config format stabilize, then design run specs that compose with both. Promoting to Phase 2 risks designing the spec format before grader configs are battle-tested.

### ESC-3: Branch protection timing

**Issue:** Tank recommends requiring CI green + 1 review on `main` branch. This changes the development workflow for everyone.

**Options:**
- (a) Enable immediately once CI (0.1) merges — strict from day one
- (b) Grace period — CI required, review recommended but not enforced, for Phase 0
- (c) Phase 1 — enable full protection after Phase 0 tasks are merged

**Recommendation:** Option (a) — enable immediately. Phase 0 is 9 small/medium tasks. If CI is the first thing we build, every subsequent Phase 0 PR benefits from it. One review requirement keeps quality high without being burdensome for a small team.

### ESC-4: `hyoka migrate-reports` command

**Issue:** Trinity recommends a command to migrate existing reports from schema version 1 (ReviewResult) to version 2 (GraderResult). This is a new feature not in the plan.

**Options:**
- (a) Add as task in Phase 2.5 — required for clean transition
- (b) Handle in rerender — rerender detects schema version and adapts
- (c) Skip — old reports stay old, new reports use new format

**Recommendation:** Option (b) — handle in rerender. A separate `migrate-reports` command is user-hostile (they have to know to run it). Rerender should transparently handle both schema versions. The schema_version field (D-AUTO-DM11) enables this. No new command needed.

---

## Phase 0 Action Items (Updated)

Changes from the original plan are marked with ⚡.

| Task | Description | Owner | Size | Depends On | Change |
|------|-------------|-------|------|------------|--------|
| 0.1 | **Create CI pipeline** (#91) — `go build`, `go vet`, `go test -race`, 2-min timeout, single Go version/OS | Tank | Medium | — | ⚡ Added `-race`, reduced timeout to 2 min |
| 0.2 | **Fix reviewer model bug** (#92) — Create reviewer per-task via factory function, not engine-scoped | Neo | Small | — | ⚡ Clarified fix approach: per-task reviewer factory |
| 0.3 | **Fix stale path in new-prompt** (#93) | Tank | Small | — | — |
| 0.4 | **Add generator model validation** (#94) | Tank | Small | — | — |
| 0.5 | **Detect duplicate config names** (#95) | Tank | Small | — | — |
| 0.6 | **Big-bang config migration** (#96) — Also update `main.go:394` write to legacy field, update `Validate()` internals | Tank | Medium | — | ⚡ Tank identified 2 additional call sites |
| 0.7 | **Log discarded errors** (#97) | Neo | Small | — | — |
| 0.8 | **Fix Go version in docs** (#98) — 4 public + 9 internal agent docs | Oracle | Small | — | ⚡ Oracle found 9 additional internal docs |
| 0.9 | **Fix flaky resourcemonitor tests** (#99) — Remove Sleep, call sample() directly, add focused ticker test | Switch | Small | — | ⚡ Switch specified fix approach |
| 0.10 | ⚡ **Add pidfile tests** (moved from 5.3a) — 136 lines, zero tests, safety-critical | Switch | Small | — | NEW — moved from Phase 5 |

**Pending Ronnie approval:**
| 0.11? | ⚡ **Fix serve runID path traversal** (ESC-1) — sanitize runID input, add handler tests | Switch | Small | — | NEW — security fix |

**Phase 0 exit criteria:** All 10 (or 11) tasks merged, CI green and enforced, branch protection enabled.

---

## Risks Identified

| # | Risk | Severity | Source | Mitigation |
|---|------|----------|--------|------------|
| R1 | Properties map design cascades to everything — wrong boundary between typed/map fields breaks graders, filters, reports | **Critical** | Neo | D-AUTO-DM2 resolves: metadata-only in map, typed fields retained for non-string data |
| R2 | GraderResult type is most expensive to fix later — every grader, report, and template depends on it | **Critical** | Neo, Trinity | D-AUTO-DM4/DM5/DM14: typed fields, concrete input, freeze before implementation |
| R3 | System prompt removal (1.6) before reviewer bug fix (0.2) could mask issues | **High** | Neo | D-AUTO-DM6: phased removal; 0.2 must merge first |
| R4 | Config migration (0.6) and property migration (1.1b) overlap creates merge conflicts | **High** | Neo | D-AUTO-DM15: these must NOT overlap; 0.6 completes fully before 1.1b begins |
| R5 | Serve runID path traversal — active security vulnerability | **High** | Switch | ESC-1: escalated to Ronnie for immediate fix |
| R6 | Phase 2.5 Neo overload — 5 tasks including 2 Large, all on critical path | **High** | Neo | Morpheus delivers 2.5a (interface design) early so Neo isn't blocked; consider splitting 2.5d |
| R7 | 44 tests need updating in Phase 1 — significantly more than "mostly mechanical" | **Medium** | Switch | Create test helpers and golden file infrastructure at Phase 1 start; track test count delta per PR |
| R8 | Report schema migration breaks rerender for existing reports | **Medium** | Trinity | D-AUTO-DM11: schema_version field; rerender handles both versions |
| R9 | html.go grows past 1600 lines in Phase 3 without template extraction | **Medium** | Trinity | D-AUTO-DM12: extract templates before 3.2a |
| R10 | 1 doc task in 72-task plan creates documentation debt | **Medium** | Oracle | D-AUTO-DM18: 12 distributed doc tasks; feature owner drafts, Oracle polishes |
| R11 | main.go split (1.5) concurrent with 5 other Phase 1 workstreams — merge conflict factory | **Medium** | Tank | Recommend 1.5 starts first in Phase 1 to minimize conflicts |
| R12 | `TrendEntry.Score` is `int`, needs `float64` for grader scores | **Low** | Trinity | Classification logic doesn't use Score; transition is isolated |

---

## Cross-Agent Questions Resolved

### Neo's Questions

**Q: Should reviewer system message be configurable in Phase 1 or deferred to Phase 2?**
→ **Phase 1.** Added as task 1.6d (D-AUTO-DM7). It's part of the same system prompt work.

**Q: After 0.6, do `cfg.EffectiveModel()` calls become `cfg.Generator.Model` directly?**
→ **Yes.** That's the entire point of 0.6. All ~17 `Effective*()` call sites become direct field access. Tank confirmed.

**Q: Property key naming convention — enforce snake_case?**
→ **Yes.** D-AUTO-DM13. Validation rejects non-snake_case keys. Migration script normalizes.

**Q: Which 6 of Waza's 12 grader types did we drop, and why?**
→ **Deferred.** Waza's full grader inventory needs review against hyoka's use cases. The initial set (file, program, prompt, behavior, action_sequence, tool_constraint) covers known requirements. Additional types can be added in later phases. This is not blocking.

**Q: After reviewer bug fix, how to verify existing reports weren't corrupted?**
→ **Spot-check.** Compare a sample of multi-config run reports: check that each config's review panel used the correct models. If corruption is found, affected reports should be re-generated. No automated migration needed — reports record which models were used.

### Tank's Questions

**Q: Should main.go split be a dependency for 1.2/1.3?**
→ **No hard dependency.** But recommend starting 1.5 first in Phase 1 to reduce merge conflicts. Other Phase 1 tasks can proceed in the split file structure.

**Q: Should 0.2 wait for 0.6?**
→ **No.** D-AUTO-DM15. These are independent. 0.2 is a P0 bug; fix immediately.

**Q: Should config_test.go backward-compat tests be kept or deleted after migration?**
→ **Deleted.** The whole point of big-bang is no backward compat. Tests that verify `Normalize()` and `Effective*()` behavior are deleted alongside that code.

**Q: `newPromptCmd` scaffolds old format — needs update after 1.1?**
→ **Yes.** D-AUTO-DM20. Added to task 1.1b scope.

**Q: Should `--model` override flag be deprecated in favor of run spec file?**
→ **Not yet.** Run spec files are still "future" (pending ESC-2). The `--model` flag stays for now.

### Trinity's Questions

**Q: Does each `prompt` grader instance run ONE model or multi-model panel?**
→ **One model.** D-AUTO-DM19. Multiple `prompt` grader instances for multi-model review. More composable and transparent.

**Q: How to type-assert `GraderResult.Details` in Go templates?**
→ **Resolved by D-AUTO-DM4.** Typed optional fields replace `interface{}`. Templates check `if .FileDetails` directly — no type assertion needed.

**Q: Is React site the intended Phase 4 dashboard home?**
→ **Yes.** D-AUTO-DM consensus. Static HTML reports stay Go-templated. Interactive dashboard is React SPA.

**Q: What's the pairwise report output format?**
→ **Deferred to task 2.1d design.** Likely a comparison matrix (baseline vs each toggle) per grader, with delta scores.

**Q: Are serve path traversal tests comprehensive enough for new API endpoints?**
→ **No.** Switch identified the current gap (ESC-1). New API endpoints (4.2a) must include traversal tests as part of implementation.

### Switch's Questions

**Q: GraderResult Details testing approach — generics or typed assertion helpers?**
→ **Typed assertion helpers.** D-AUTO-DM4 uses typed optional fields, so test helpers assert on concrete types: `assertFileDetails(t, result)`, `assertProgramDetails(t, result)`.

**Q: CI branch protection timing — need flaky fix (0.9) before CI?**
→ **No.** 0.1 and 0.9 can proceed in parallel. If the flaky test fails in CI before 0.9 merges, that's fine — it proves CI is working. 0.9 then fixes it.

**Q: Reuse or rewrite PanelReviewer.ReviewPanel() for prompt grader?**
→ **Wrap, don't rewrite.** The `prompt` grader (2.5d) wraps the current reviewer behavior. PanelReviewer becomes a multi-instance prompt grader pattern. Full replacement happens when all graders are working.

**Q: engine.go:301 production time.Sleep — intentional or TODO?**
→ **Needs investigation.** Neo should check this during 0.7 (log discarded errors) as it's the same code area. If it's a rate limiter, it should be documented. If it's a workaround, it should be tracked.

**Q: Should we track test count/coverage delta per PR?**
→ **Track count, don't enforce minimum.** Add test count to CI output. Don't block PRs on coverage thresholds in Phase 0, but monitor the trend. Target: 260 → 350+ by end of Phase 2.

### Oracle's Questions

No explicit questions raised, but the recommendation for 12 documentation tasks is adopted (D-AUTO-DM18). Oracle should enumerate specific doc tasks during Phase 0 so they're ready when Phase 1 begins.

---

## Unresolved Questions

1. **Waza grader inventory delta** — Neo asked which 6 of Waza's 12 grader types were dropped. Requires investigation of Waza's current grader set. Not blocking for Phase 0-1, needed before Phase 2.5.

2. **Pairwise report output format** — Trinity asked, and the answer depends on 2.1a design. Deferred to Phase 2 design phase.

3. **engine.go:301 `time.Sleep`** — Switch flagged. Needs investigation. May be intentional rate limiting or may be a TODO.

4. **`copilot.ClientOptions` shared by all reviewers** — Neo flagged as "currently safe but fragile." Monitor during 0.2 fix; may need per-task client options if any reviewer-specific options are added later.

5. **Phase 2.5 workload distribution** — Neo has 5 tasks (2 Large) on the critical path. May need to redistribute 2.5e (behavior/action_sequence/tool_constraint graders) if the `file`/`program`/`prompt` graders take longer than estimated.

6. **`generatorSkillsDirs` and `reviewerSkillsDirs` resolution** — Neo noted these are resolved once from `f.prompts`, not per-config. Same architectural flaw as the reviewer bug. Should be addressed in 0.2 or as a follow-up.

7. **Backward compatibility policy** — Oracle flagged the plan lacks a formal backward compat statement. Big-bang migrations for prompts (1.1b) and configs (0.6) are decided, but the report format transition (Phase 2.5g) needs a more nuanced policy. Schema versioning (D-AUTO-DM11) partially addresses this.

---

## Summary of Changes to Evolution Plan

| What | Original | Updated |
|------|----------|---------|
| CI spec | `go build/vet/test` | Added `-race`, 2-min timeout |
| Phase 0 tasks | 9 tasks (0.1–0.9) | 10 tasks (+0.10 pidfile tests), possibly 11 (+serve security fix) |
| Properties map scope | Replace ALL typed fields | Metadata-only; Tags, Timeout, etc. stay typed |
| GraderResult.Details | `interface{}` | Typed optional fields per grader kind |
| GraderInput | Unspecified | Concrete struct |
| Grader config schema | weight only | Added `gate: bool` for hard pass/fail |
| System prompt removal | Implicit big-bang | Explicit phased: bias → guidance → path → safety |
| Phase 1 tasks | 20 tasks | +1 (1.6d reviewer system message), +1 (1.8 review coverage) |
| Pre-Phase 3 | — | +1 (3.0 template extraction) |
| prompt grader | Multi-model unclear | One model per instance, compose for multi-model |
| Documentation | 1 task | 12+ tasks distributed across phases |
| Test target | Unspecified | 260 → 350+ by Phase 2, 400+ by Phase 5 |

---

### Decision: Escalated Decisions from Design Meeting (2026-04-04T19:09Z)

**By:** Ronnie Geraghty  
**Status:** Binding

**ESC-1 (Serve runID path traversal):** Option (b) — Hotfix PR immediately, outside the evolution plan. Security issues don't wait for sprint planning.

**ESC-2 (Run spec file timing):** Option (b) — Keep as "future". Let main.go split and grader config format stabilize first.

**ESC-3 (Branch protection):** Option (a) — Enable immediately once CI (#91) merges. Every subsequent Phase 0 PR benefits.

**ESC-4 (Report migration):** Migrate reports in-place during Phase 2. No dual-format support, no new command. Old JSON gets rewritten to v2 schema as part of grader architecture work. `schema_version` field (DM11) included for future-proofing but only latest version supported. Project is not in stable mode — no backward compatibility obligation.
