# Hyoka Session Decisions Log

**Session Date:** 2026-10-14  
**Participants:** Ronnie Geraghty, Morpheus (Lead/Architect)

This document records all decisions made during the October 2026 hardening and evolution planning session.

---

## 1. Hardening Audit Priorities

**Decision:** Prioritized 20 hardening tasks into P0/P1/P2 tiers based on impact and risk.

**P0 — Must Fix Before Real Use:**
1. Fix reviewer model bug — multi-config evals use wrong reviewer panel (`main.go:469-473`)
2. Create CI pipeline — no `go build`/`go test`/`go vet` in CI today
3. Fix stale path in `new-prompt` command (`main.go:1276`)

**P1 — Should Fix Soon:**
4. Add generator model validation (empty model passes validation)
5. Split `main.go` into per-command files (1329 lines)
6. Add pidfile tests (zero tests — only untested package)
7. Add stub integration test (no end-to-end test exists)
8. Log discarded errors (reviewer.go:352, copilot.go:83, others)
9. Fix Go version in docs (says 1.24.5+, should be 1.26.1+)
10. Detect duplicate config names (silent shadowing)
11. Validate `runID` in serve handler (inconsistent path traversal defense)
12. Fix flaky resourcemonitor tests (`time.Sleep` assertions)
13. Add early auth check (issue #72)

**P2 — Nice to Have:**
14. Extract HTML templates from `html.go` (1374 lines of string concat)
15. Add "did you mean?" for config names
16. Call `session.Disconnect()` before `DeleteSession` (issue #71)
17. Embed Copilot CLI binary (issue #73)
18. Add PID birth-time validation
19. Review package needs more tests (5 tests for 732 lines)
20. Remove legacy config fields

---

## 2. Evolution Plan Phase Structure

**Decision:** 5-phase structure with hardening integrated into Phase 0.

| Phase | Focus | Duration |
|-------|-------|----------|
| 0 | Foundation (CI + P0 bugs) | 1 sprint |
| 1 | Core Model Changes (properties, criteria, tools, prompts, system prompt) | 2-3 sprints |
| 2 | Evaluation Engine (pairwise, limits, isolation, resources) | 2-3 sprints |
| 3 | Transparency (action history, review panel, reviewer tools) | 1-2 sprints |
| 4 | Insights & Comparison (comparison engine, dashboard, trends) | 2 sprints |
| 5 | Ecosystem (.hyoka dir, marketplace, tests, skills) | 2-3 sprints |

**Critical path:** CI → Generic Properties → Tool Filters → Pairwise → Comparison

---

## 3. Open Question Answers (Q1-Q6)

### Q1: Tier 1 Criteria — Remove Entirely

**Question:** Keep Tier 1 (built-in default) criteria as opt-in, or remove entirely?  
**Decision:** Remove entirely. Prompts and configs must supply their own criteria. No built-in defaults.  
**Rationale:** Defaults bias evaluation. Teams should own their criteria.

### Q2: System Prompt Scope — Zero by Default

**Question:** How minimal is "minimal" for system prompts?  
**Decision:** Zero system prompt by default. Only isolation-related config (working directory, tools) handled through SDK `SessionConfig`, not prompt injection. Hardcoded guardrails (in code) are better than system prompt guardrails. If isolation can be achieved through SDK config alone, don't put it in the system prompt at all.  
**Rationale:** System prompt rules bias agent behavior. The purpose of hyoka is measuring natural agent behavior with different tools.

### Q3: Pairwise Testing — Flag-Based with Opt-Out

**Question:** Full combinatorial pairwise or MCP-server-only?  
**Decision:** `--pairwise` / `-pw` flag on the `run` command. When passed, expands one config into pairwise variants. Config YAML supports `always_on: true` per tool to exempt from toggling.  
**Rationale:** Flag makes it explicit. `always_on` prevents toggling infrastructure tools that shouldn't be tested.

### Q4: Property Migration — Big-Bang

**Question:** Gradual backward-compatible migration or big-bang?  
**Decision:** Big-bang. Update all 87 prompts to new format. No backward compatibility for old fields.  
**Rationale:** Backward compat adds permanent complexity. 87 prompts is a manageable migration scope.

### Q5: `.hyoka` Directory — Project-Scoped Only

**Question:** Project-scoped only or both project and global?  
**Decision:** `.hyoka` only, project-scoped. Structured subdirectories: `configs/`, `prompts/`, `criteria/`, `skills/`, `reports/`. No global install mode.  
**Rationale:** Global config creates hidden state. Project-scoped is explicit and versionable.

### Q6: Response Type — Config-Level System Prompt

**Question:** Files only, or text responses valid?  
**Decision:** Response type is a system-prompt-level concern. Config-specific system prompt decides whether agent should produce files, text, or both. Look at microsoft/waza for reference.  
**Rationale:** Different evaluation scenarios need different response types. This is configurable, not hardcoded.

---

## 4. Late Additions

### 4a. Reviewer Tools

**Decision:** Review panel agents should be able to have tools added to their environments — not just the generation agent. Reviewers can use linters, style checkers, documentation references, etc.  
**Config location:** `reviewer.tools` section in config YAML.  
**Phase:** 3 (Transparency)

### 4b. Configurable System Prompts for Both Agents

**Decision:** System prompt for BOTH the generation agent AND review agents should be configurable in config YAML. Supports the "minimal to no system prompt bias" goal.  
**Config location:** `generator.system_prompt` and `reviewer.system_prompt` in config YAML.  
**Phase:** 1 (Core Model)

### 4c. Starter Files

**Decision:** Core feature. Prompts can reference starter files that get placed in the agent's working directory before the session begins. Follows Waza's `ResourceFile` pattern.  
**Use case:** "Fix this broken code" prompts, "Extend this project" prompts.  
**Phase:** 1 (Core Model)

### 4d. Skills Investigation

**Decision:** Investigate and recommend project-specific skills for the repo. Morpheus identified 14 skills across 4 categories.  
**Phase:** 5 (Ecosystem)

---

## 5. Waza Reference Architecture

**Decision:** Adopt microsoft/waza patterns as reference architecture for hyoka's evolution:
- Zero system prompt (SDK config handles everything)
- `ResourceFile` for starter files
- Session config–based isolation
- Config-level system prompt override

**Rationale:** Waza is the Azure SDK team's production agent eval tool, uses the same Copilot SDK, and has solved many of the same problems hyoka is approaching.

---

## 6. Zero System Prompt

**Decision:** Follow Waza's approach. Zero system prompt for agent evaluation sessions by default. All configuration handled through SDK `SessionConfig`. Config-specific custom system prompts remain available for users who want them.

**What gets removed:** All 15 hardcoded rules currently in `copilot.go:628-655`:
- File creation rules (1-7) — handled by SDK working directory config
- Bash rules (8) — handled by SDK working directory config
- Research restrictions (9) — removed (bias)
- Python rules (10) — removed (bias)
- Safety boundaries (11-15) — moved to code-level hooks or SDK config

**What stays:** Pre/post tool hooks for file path validation (`OnPreToolUse`, `OnPostToolUse`). These are code-level guardrails, not system prompt injection.

---

## 7. Skill Philosophy

**Decision:** Project-specific skills should be advisory, not prescriptive. They should NOT say "the core eval process should always work like this" because the project is evolving. Instead, skills capture core principles and warn when work goes against them — guardrails, not cages.

**Rationale:** The project is in active evolution (hardening + product vision). Rigid skills would block progress. Skills should encode *why* patterns exist, not mandate specific implementations.

---

## 8. Plan Directory Structure

**Decision:** Create a `plan/` directory at repo root for evolution planning documents. Separate from `docs/` which documents the current tool.

**Contents:**
- `plan/evolution-plan.md` — 5-phase plan with tasks, assignments, dependencies
- `plan/core-principles.md` — Guiding principles of hyoka as a product
- `plan/feature-requirements.md` — Structured PRD with 18 features
- `plan/engineering-standards.md` — Software engineering principles for implementation
- `plan/decisions-log.md` — This file

**Rationale:** `docs/` = current state, `plan/` = future state. Someone reading `plan/` should understand the full vision without reading `docs/`.

---

## Decision Index

| # | Decision | Section | Impact |
|---|----------|---------|--------|
| D1 | Hardening priorities (P0/P1/P2) | §1 | Immediate work queue |
| D2 | 5-phase structure | §2 | Project timeline |
| D3 | Remove Tier 1 criteria | §3 Q1 | Criteria system redesign |
| D4 | Zero system prompt default | §3 Q2, §6 | Core architecture change |
| D5 | Pairwise flag with always_on | §3 Q3 | New feature design |
| D6 | Big-bang property migration | §3 Q4 | Migration strategy |
| D7 | .hyoka project-scoped only | §3 Q5 | Ecosystem design |
| D8 | Response type via system prompt | §3 Q6 | Config flexibility |
| D9 | Reviewer tools | §4a | Review panel extension |
| D10 | Configurable system prompts | §4b | Config schema change |
| D11 | Starter files | §4c | New feature |
| D12 | Skills investigation (14 recommended) | §4d | Ecosystem work |
| D13 | Waza reference architecture | §5 | Architectural direction |
| D14 | Skills are guardrails not cages | §7 | Philosophy |
| D15 | plan/ directory for vision docs | §8 | Project organization |
