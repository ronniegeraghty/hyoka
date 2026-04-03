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
