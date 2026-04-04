# Project Context

- **Project:** hyoka — Go evaluation tool that runs prompts through the Copilot SDK, reviews code via a multi-model panel, produces criteria-based pass/fail reports
- **Stack:** Go 1.26.1+, GitHub Copilot CLI/SDK, MCP servers (Azure MCP via npx)
- **User:** Ronnie Geraghty
- **Created:** 2026-04-03
- **Repo:** /home/rgeraghty/projects/hyoka
- **Key paths:** hyoka/internal/ (core engine), hyoka/main.go (entry point), prompts/ (evaluation prompts), criteria/ (pass/fail criteria)

## Core Context

Agent Neo initialized as Core Dev for hyoka. Owns the evaluation engine, review panel, criteria logic, and Copilot SDK integration. The tool has guardrails (max turns: 25, max files: 50, max output: 1MB, max session actions: 50). Safety boundaries prevent real Azure resource provisioning by default (`--allow-cloud` to opt out).

## Recent Updates

📌 Team initialized on 2026-04-03

📋 **Morpheus Audit (2026-04-03):** Comprehensive codebase health assessment complete. Key finding: **reviewer model bug in main.go:469-473** (P0) — multi-config evaluations share one reviewer panel. See `.squad/decisions.md` for full P0/P1/P2 action items.

## Learnings

### Issue #92: Per-Task Reviewer Panel Creation (2025-01-19)
**Branch:** `ronniegeraghty/issue-92-reviewer-model-bug`  
**PR:** [#170](https://github.com/ronniegeraghty/hyoka/pull/170)  
**Status:** ✅ Complete

Fixed critical bug where multi-config evaluations reused the reviewer panel from the FIRST config for ALL configs, causing every evaluation to use incorrect reviewer models.

**Implementation:**
- Introduced `ReviewerFactory` function type that creates reviewers per-config
- Replaced `Engine.reviewer/panelReviewer` fields with `reviewerFactory` field
- Moved reviewer creation from main.go into `runSingleEval()` using `task.Config`
- Each config now gets its own reviewer panel with correct models
- Added `NewEngineWithReviewerFactory()` constructor
- Maintained backward compatibility with deprecated `NewEngineWithReviewer()`

**Testing:**
- Created `reviewer_factory_test.go` with 3 tests verifying correct behavior
- All existing tests pass
- Build and vet clean

**Learnings:**
1. **Reviewer Factory Pattern**: When multiple configs need different reviewer settings, create reviewers lazily per-task rather than once at engine creation. Use a factory function that closes over shared resources (clientOpts) but creates instances based on task.Config.

2. **Backward Compatibility**: When refactoring constructors, wrap deprecated APIs to call new implementation. Preserves existing call sites while enabling new patterns.

3. **Testing Concurrent Tasks**: Don't assert on execution order. Use maps to track outcomes by task ID when testing engines with concurrent workers.

Initial setup complete. Architecture is sound. Main engineering focus should be: (1) fix reviewer model bug, (2) refactor main.go into cmd/ package, (3) add integration tests.

### Session 2026-04-04T00-05 (Morpheus Evolution Plan)

Evolution plan assigns you Phase 1 core model changes (generic properties, criteria filters, tool filters) and Phase 2 pairwise testing. Read `.squad/decisions.md` for full plan. Also assigned: reviewer model bug (P0), discarded error logging, early auth check.

### Session 2026-04-04T19:45 (Phase 0 Execution — Reviewer Factory Fix)

**Status:** COMPLETE  
**Issue:** #92  
**PR:** #170

Implemented ReviewerFactory pattern to fix multi-config reviewer panel bug. Each config now receives correct reviewer models instead of all configs using first config's reviewers.

**Key outcome:** Lazy per-task reviewer creation replaces engine-scoped setup. Factory pattern enables clean separation of concerns and tested backward compatibility.

**Cross-agent dependency:** Tank's config migration (#96, PR #171) enabled clean Generator/Reviewer schema that makes this fix viable. Switch's flaky test fix (#99, PR #167) and Tank's CI pipeline (#91, PR #168) ensure test reliability in review panel code.

**Files:** engine.go, main.go, reviewer_factory_test.go

