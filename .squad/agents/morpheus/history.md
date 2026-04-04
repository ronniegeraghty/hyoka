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
