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

Initial setup complete. Architecture is sound. Main engineering focus should be: (1) fix reviewer model bug, (2) refactor main.go into cmd/ package, (3) add integration tests.

### Session 2026-04-04T00-05 (Morpheus Evolution Plan)

Evolution plan assigns you Phase 1 core model changes (generic properties, criteria filters, tool filters) and Phase 2 pairwise testing. Read `.squad/decisions.md` for full plan. Also assigned: reviewer model bug (P0), discarded error logging, early auth check.
