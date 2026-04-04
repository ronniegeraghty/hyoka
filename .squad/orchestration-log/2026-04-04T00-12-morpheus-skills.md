# Orchestration Log: Morpheus Skills Investigation

**Date:** 2026-04-04T00:12  
**Agent:** Morpheus 🕶️  
**Task:** Skills investigation for squad effectiveness  
**Mode:** background  
**Outcome:** Completed  

## Summary

Morpheus investigated what agent skills should live in the hyoka repository to support squad members and human developers. The investigation identified 14 discrete skills organized across 4 categories, grounded in patterns observed during comprehensive codebase audits.

## Outputs

- **Primary:** `.squad/decisions/inbox/morpheus-skills-recommendations.md` (24 KB)
  - 14 skills recommended across 4 categories
  - Each skill includes: audience, scope, rationale, key content to encode, file references
  - Skills cover architecture, Go patterns, domain conventions, testing, CLI, reporting, and team onboarding
  
- **Secondary:** `.squad/agents/morpheus/history.md` (updated)
  - Session history from skills investigation work

## Skill Categories

1. **Core Architecture (All Agents + Human Devs)** — 4 skills
   - `hyoka-eval-pipeline` — generate→review→report orchestration
   - `hyoka-error-handling` — error wrapping, propagation, non-fatal logging conventions
   - `hyoka-config-system` — YAML loading, normalization, validation, `Effective*()` getters
   - `copilot-sdk-integration` — session lifecycle, event handling, resource cleanup

2. **Go Patterns & Conventions** — 2 skills
   - `hyoka-criteria-system` — tiered evaluation (pass/warn/fail), multi-level scoring, rollups
   - `hyoka-testing-patterns` — test file structure, table-driven tests, mock patterns

3. **Subsystem Expertise** — 6 skills
   - `hyoka-cli-patterns` — Cobra command registration, flag defaults, safety boundaries
   - `hyoka-report-generation` — JSON/HTML/MD templates, data transformation, multi-format output
   - `hyoka-logging-conventions` — slog structured logging, role-prefixed output, diagnostic levels
   - `hyoka-contributor-guide` — workflow, local testing, quick iteration with Python prompts
   - `hyoka-prompt-conventions` — frontmatter schema, validation, categorization
   - `hyoka-property-migration` — legacy field handling, idempotent normalization, backward compatibility

4. **Operational Knowledge** — 2 skills
   - `hyoka-process-lifecycle` — session management, PID files, orphan detection, cleanup heuristics
   - `hyoka-serve-patterns` — web handler routing, path traversal safety, report serving

## Recommendation

All 14 skills should be created and published to `.squad/skills/` to enable effective asynchronous agent collaboration on hyoka hardening and evolution tasks. Skills encode architectural knowledge, Go patterns, domain conventions, and operational expertise that would otherwise be rediscovered repeatedly during implementation.

## Next Steps

Scribe will merge the skill summary into `decisions.md`, then delete the inbox file.
