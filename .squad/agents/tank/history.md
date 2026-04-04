# Project Context

- **Project:** hyoka — Go evaluation tool that runs prompts through the Copilot SDK, reviews code via a multi-model panel, produces criteria-based pass/fail reports
- **Stack:** Go 1.26.1+, GitHub Copilot CLI/SDK, MCP servers (Azure MCP via npx)
- **User:** Ronnie Geraghty
- **Created:** 2026-04-03
- **Repo:** /home/rgeraghty/projects/hyoka
- **Key paths:** hyoka/main.go (CLI entry), configs/ (run configs), reports/ (output), site/ (docs/serving), go.work (workspace)

## Core Context

Agent Tank initialized as Platform Dev for hyoka. Owns CLI, config, build, reports, site, and plugins. The CLI supports: `list` (show prompts), `run` (execute evaluations with filters like --service, --language, --all-configs). Smart path detection checks ./prompts then ../prompts. Fan-out confirmation prompts at >10 evals. Uses go.work workspace for multi-module builds.

## Recent Updates

📌 Team initialized on 2026-04-03

📋 **Morpheus Audit (2026-04-03):** Audit of CLI and platform layer complete. Key findings: (1) **stale path in main.go:1277** (P0) — new-prompt output references `go run ./tool/cmd/hyoka validate`, should be `go run ./hyoka validate`. (2) **main.go refactor candidate (P1)** — 1329 lines, split into cmd/ package recommended. See `.squad/decisions.md` for full list.

## Learnings

- **Config migration pattern**: When removing backward compatibility code, the best approach is to update all configs first (all 8 YAML files), then delete legacy struct fields, then remove helper methods (Normalize, Effective*), then update all call sites. This ensures compiler errors guide you to every place that needs updating.
- **Test-driven refactoring**: Large structural changes benefit from running tests after each phase (struct changes, method deletions, call site updates). The test failures become a checklist of what still needs updating.
- **Unused function cleanup**: After removing legacy fields, helper functions like resolveSkillsDirs() that worked with those fields become unused. The compiler catches these with "declared and not used" errors.


Initial setup complete. Platform is well-structured. Quick wins: fix stale path, plan main.go refactor.

### Session 2026-04-04T00-05 (Morpheus Evolution Plan)

Evolution plan assigns you Phase 0 CI pipeline (P0), main.go split, YAML prompts, session limits, .hyoka directory. Read `.squad/decisions.md` for full plan. Also assigned: config validation, duplicate detection, stale path fixes.

### Session 2026-04-04T03:28 (Issue Creation)

Created 72 GitHub issues (#91–#162) for evolution plan across 5 phases:
- Phase 0 (CI pipeline): #91–#99 (9 issues)
- Phase 1 (CLI & config): #100–#119 (20 issues)
- Phase 2 (Report & skill system): #120–#137 (18 issues)
- Phase 3 (SDK & validation): #138–#143 (6 issues)
- Phase 4 (Extensibility): #144–#151 (8 issues)
- Phase 5 (Polish & tools): #152–#162 (11 issues)

All issues labeled, assigned, and staged for backlog prioritization. Backlog is fully populated.

### Session 2026-04-04T19:30 (CI Pipeline Implementation)

**Issue #91:** Created GitHub Actions CI workflow (`.github/workflows/ci.yml`)
- Triggers: All PRs + pushes to main/ronniegeraghty/dev
- Go 1.26.1 (matches go.mod), ubuntu-latest
- Steps: build → vet → test with `-race` flag
- 2-minute timeout (tests run in ~5s with race, 24× headroom)
- Race detection required from day one per D-AUTO-DM8 (concurrent code in ResourceMonitor, ProcessTracker, PanelReviewer)
- Verified all commands pass locally before pushing
- Branch: `ronniegeraghty/issue-91-ci-pipeline`
- PR: #168 → ronniegeraghty/dev

**Key learnings:**
- CI (task 0.1) is explicit blocker for Phase 1 per D-AUTO-DM17 — nothing merges until CI is green
- Race detector adds ~4-5s overhead to test runtime but catches concurrency bugs early
- Phase 0 keeps it simple: no golangci-lint yet (deferred to Phase 1)
- File path: `.github/workflows/ci.yml` (34 lines, YAML)
- **CI verified working:** First run on PR #168 passed in 1m3s (build + vet + test with -race)
