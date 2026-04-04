# Squad Decisions

## Active Decisions

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
