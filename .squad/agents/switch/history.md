# Project Context

- **Project:** hyoka — Go evaluation tool that runs prompts through the Copilot SDK, reviews code via a multi-model panel, produces criteria-based pass/fail reports
- **Stack:** Go 1.26.1+, GitHub Copilot CLI/SDK, MCP servers (Azure MCP via npx)
- **User:** Ronnie Geraghty
- **Created:** 2026-04-03
- **Repo:** /home/rgeraghty/projects/hyoka
- **Key paths:** hyoka/main_test.go, hyoka/testdata/, hyoka/internal/ (packages to test)

## Core Context

Agent Switch initialized as Tester for hyoka. Guardrail defaults to test: max turns 25, max files 50, max output 1MB, max session actions 50. Safety boundaries prevent real Azure provisioning (--allow-cloud to opt out). Fan-out confirmation at >10 evals. Tests run via `go test ./...` from workspace root.

## Recent Updates

📌 Team initialized on 2026-04-03

📋 **Morpheus Audit (2026-04-03):** Testing audit complete. Key findings: (1) **pidfile package has zero tests (P1)** — needs test coverage. (2) **no integration test for eval pipeline (P1)** — TestRunEndToEnd recommended with StubEvaluator + StubReviewer. (3) **error logging gaps** — several discarded errors not logged. See `.squad/decisions.md` for prioritized test work.

## Learnings

Initial setup complete. Coverage is good across packages except pidfile. Integration test gap is highest priority for catching regressions.

### Session 2026-04-04T00-05 (Morpheus Evolution Plan)

Evolution plan assigns you pidfile tests, review package tests, integration test, flaky test fixes. Read `.squad/decisions.md` for full plan.
