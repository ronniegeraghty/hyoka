# Work Routing

How to decide who handles what.

## Routing Table

| Work Type | Route To | Examples |
|-----------|----------|----------|
| Eval engine, review panel, criteria, prompts, SDK integration, skills, plugins | Neo 💊 | Build reviewer pipeline, add criteria, fix prompt parsing, Copilot SDK client, skills framework, plugin architecture |
| CLI commands, flags, config, environment checks, validation, progress | Tank 📡 | Add CLI flag, fix config loading, checkenv logic, validate command, progress output, go.work issues |
| Reports, templates, site, serve, trends, frontend rendering | Trinity 🖤 | HTML/JSON/Markdown reports, Go template authoring, serve command, site assets, trends visualization |
| Architecture, system design, scope, technical direction | Morpheus 🕶️ | Design module boundaries, review architecture proposals, scope decisions |
| Code review | Morpheus 🕶️ | Review PRs, check quality, approve/reject changes |
| Testing, QA, edge cases, CI/CD, GitHub Actions | Switch 🤍 | Write tests, find edge cases, verify guardrails, coverage analysis, CI workflow maintenance |
| Documentation, README, CHANGELOG, inline docs, AGENTS.md | Oracle 🔮 | Update docs for new features, keep README current, CHANGELOG entries, Go doc comments |
| Scope & priorities | Morpheus 🕶️ | What to build next, trade-offs, triage issues |
| Session logging | Scribe | Automatic — never needs routing |

## Issue Routing

| Label | Action | Who |
|-------|--------|-----|
| `squad` | Triage: analyze issue, assign `squad:{member}` label | Lead |
| `squad:{name}` | Pick up issue and complete the work | Named member |

### How Issue Assignment Works

1. When a GitHub issue gets the `squad` label, the **Lead** triages it — analyzing content, assigning the right `squad:{member}` label, and commenting with triage notes.
2. When a `squad:{member}` label is applied, that member picks up the issue in their next session.
3. Members can reassign by removing their label and adding another member's label.
4. The `squad` label is the "inbox" — untriaged issues waiting for Lead review.

## Rules

1. **Eager by default** — spawn all agents who could usefully start work, including anticipatory downstream work.
2. **Scribe always runs** after substantial work, always as `mode: "background"`. Never blocks.
3. **Quick facts → coordinator answers directly.** Don't spawn an agent for "what port does the server run on?"
4. **When two agents could handle it**, pick the one whose domain is the primary concern.
5. **"Team, ..." → fan-out.** Spawn all relevant agents in parallel as `mode: "background"`.
6. **Anticipate downstream work.** If a feature is being built, spawn the tester to write test cases from requirements simultaneously.
7. **Issue-labeled work** — when a `squad:{member}` label is applied to an issue, route to that member. The Lead handles all `squad` (base label) triage.
