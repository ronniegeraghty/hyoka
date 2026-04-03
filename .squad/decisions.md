# Squad Decisions

## Active Decisions

### Decision: Hardening Audit — Priorities

**Date:** 2026-04-03  
**Author:** Morpheus 🕶️  
**Status:** Proposed  

#### P0 — Fix Now

1. **Fix reviewer model bug in main.go:469-473.** Reviewer models are grabbed from the first config only. When running multi-config evaluations with different reviewer panels, all configs incorrectly share one panel. The reviewer setup should be per-config or validated to be consistent.

2. **Fix stale path in new-prompt output (main.go:1277).** Says `go run ./tool/cmd/hyoka validate` — should be `go run ./hyoka validate`.

#### P1 — Do Soon

3. **Split main.go (1329 lines) into cmd/ package.** Each cobra command should be its own file. The run command alone is ~300 lines of wiring logic. Proposed structure: `hyoka/cmd/{root,run,list,configs,validate,trends,report,serve,clean,plugins,newprompt,checkenv,version}.go`.

4. **Add integration test for the eval pipeline.** Currently no test exercises generate→review→report end-to-end, even with a stub evaluator. One `TestRunEndToEnd` with StubEvaluator + StubReviewer would catch wiring regressions.

5. **Add tests for pidfile package.** Only package with zero tests.

6. **Log discarded errors instead of silently ignoring.** Key locations: review/reviewer.go:352 (referenceFiles), eval/copilot.go:83 (starterFiles), multiple filepath.Walk returns.

#### P2 — Do Eventually

7. **Extract HTML templates from html.go (1374 lines).** The report HTML is built with string concatenation — consider embedding template files.

8. **Reduce action-counting duplication.** Action limits enforced in 3 places (engine, CopilotReviewer, PanelReviewer) with identical logic. Extract to shared helper.

9. **Address the 10 open issues** — #72 (auth check) and #71 (session disconnect) would directly improve robustness.

10. **Consider removing legacy config fields** after a deprecation period. The dual-path (legacy + new) adds maintenance burden.

#### Rationale

The codebase is in surprisingly good shape for its maturity stage. The architecture is clean, the dependency graph is acyclic, error handling is mostly proper, and test coverage exists for every package except pidfile. The main risks are: (1) the reviewer model bug silently producing wrong results, (2) the main.go monolith slowing iteration, and (3) the lack of end-to-end integration tests.

## Governance

- All meaningful changes require team consensus
- Document architectural decisions here
- Keep history focused on work, decisions focused on direction
