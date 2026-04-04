# Hyoka Core Principles

**Date:** 2026-10-14  
**Author:** Morpheus (Lead/Architect), based on direction from Ronnie Geraghty  
**Status:** Active — advisory, not prescriptive

These principles guide hyoka's design decisions. They describe *why* we build things a certain way. They are guardrails, not cages — the project is evolving, and these principles evolve with it.

---

## 1. Transparency

**Nothing is hidden.**

Every agent action — every tool call, every file read, every bash command — is visible to the evaluator. The review panel's reasoning is exposed, not just its scores. The consolidation algorithm is explicit. System prompts are visible in config files, not buried in code.

If a user can't see what happened during an evaluation, the evaluation has failed regardless of the score.

*Implication:* Full action timeline in reports. Per-reviewer reasoning displayed. No hidden system prompt rules.

---

## 2. Unbiased Measurement

**Zero system prompt by default. No hidden behavior shaping.**

The whole point of hyoka is measuring what agents do *naturally* with different tool configurations. Injecting behavioral rules into the system prompt defeats that purpose. If an agent writes bad code without guidance, that's a valid data point — not something to paper over with prompt engineering.

System prompts are available as a configurable option for teams that want them, but the default is empty. Operational concerns (working directory, tool availability) are handled through SDK session configuration, not prompt injection.

*Implication:* Default system prompt is empty. Config YAML supports optional `system_prompt` for both generator and reviewer agents. Isolation achieved through SDK config, not system prompt rules.

---

## 3. Tool Impact Is the Core Question

**Which tools help agents write better code? Which ones hurt?**

This is the fundamental question hyoka exists to answer. Every feature, every report, every comparison should make it easier to answer this question. Pairwise testing isolates individual tool impact. Comparison views show tool-by-tool score deltas. Trends track how tool configurations perform over time.

Hyoka is not a pass/fail grader. It's a measurement instrument for tool effectiveness.

*Implication:* Pairwise testing as a first-class feature. Comparison engine as primary output. Tool impact scores in every report.

---

## 4. Generality

**Not Azure-specific. Not any-team-specific.**

Hyoka started as an Azure SDK evaluation tool, but it's becoming a general-purpose AI agent benchmarking platform. Any team should be able to bring their own prompts, criteria, tools, and system prompts. The data model must support arbitrary properties, not hardcoded fields like `Service`, `Plane`, or `SDKPackage`.

The repository's `prompts/` and `configs/` directories contain Azure SDK examples, but hyoka itself makes no assumptions about what's being evaluated.

*Implication:* Generic `Properties map[string]string` on prompts. Property-based criteria matching. Property-based tool filters. `.hyoka` project directory for team-specific configuration.

---

## 5. Isolation

**The evaluation environment doesn't leak into the user's setup.**

Running hyoka should not install tools globally, modify system configs, leave orphan processes, or pollute the user's environment. Each evaluation session runs in its own workspace. Cleanup is thorough and automatic.

Conversely, the user's environment should not leak into evaluations. The agent being tested should see only what the config explicitly provides — starter files, tools, and SDK resources. No ambient environment contamination.

*Implication:* Isolated workspace per session. Clean shutdown with orphan process detection. PID file tracking. `hyoka clean` command for manual recovery. No global state mutation.

---

## 6. Resource Responsibility

**No memory hogging. Clean shutdown. Proper cleanup.**

Hyoka runs long evaluations that spawn subprocesses, maintain SDK sessions, and generate large reports. It must be a responsible citizen:
- Bounded memory usage, even for large eval runs
- Graceful two-phase shutdown (SIGTERM → wait → SIGKILL)
- All goroutines terminate on context cancellation
- PID files track child processes for orphan recovery
- Session state cleaned up even on crashes

*Implication:* Resource monitoring (`--monitor-resources`). Proper signal handling. Deferred cleanup. Streaming for large outputs. `hyoka clean` as safety net.

---

## 7. Insights First

**Data comparison is the primary output, not just pass/fail.**

A score of 7/10 is meaningless without context. *Compared to what?* Hyoka's primary value is comparative: this config vs. that config, this tool set vs. that tool set, this week vs. last week. Reports should surface actionable insights — where did adding a tool improve scores? Where did it cause regressions?

*Implication:* Comparison engine as a core package. Side-by-side config views. Trend analysis with property-based slicing. Regression detection. Interactive dashboard with drill-down.

---

## 8. Configurable, Not Opinionated

**Users control system prompts, criteria, tool filters, session limits.**

Hyoka provides sensible defaults (zero system prompt, reasonable session limits, no built-in criteria) but users can override everything. The tool doesn't impose opinions about how agents should behave — it measures how they *do* behave under user-specified conditions.

This applies to the entire evaluation pipeline:
- **Generation agent:** Configurable system prompt, tools, skills, model, session limits
- **Review agents:** Configurable system prompt, tools, skills, models, criteria
- **Criteria:** User-defined, property-matched, with include/exclude filters
- **Reports:** Multiple formats (JSON, HTML, Markdown), customizable templates

*Implication:* All behavioral knobs exposed in config YAML. No hardcoded criteria tiers. No mandatory system prompt rules. Sensible defaults that users can override.

---

## 9. Simplicity of Dependencies

**Minimal external dependencies. Standard library preferred.**

Hyoka has 3 direct dependencies: Copilot SDK (core requirement), Cobra (CLI framework), and YAML parser. Everything else uses Go's standard library — `log/slog` for logging, `net/http` for serving, `text/template` for templating, `os/exec` for processes.

This isn't minimalism for its own sake. Fewer dependencies mean fewer supply chain risks, fewer version conflicts, faster builds, and easier auditing.

*Implication:* No third-party logging libraries. No HTTP frameworks. No ORM. New dependencies require explicit justification.

---

## 10. Skills as Guardrails

**Advisory, not prescriptive. Guardrails, not cages.**

Project-specific skills encode architectural knowledge and conventions, but they should never say "the core eval process should always work like this." The project is evolving — rigid skills block progress.

Good skills warn when work goes against established principles. They don't prevent the work from happening.

*Implication:* Skills capture *why* patterns exist, not just *what* they are. Skills reference principles, not implementation details. Skills evolve as the project evolves.

---

## How to Use These Principles

These principles resolve ambiguity in design decisions. When choosing between approaches:

1. Does one approach expose more information to the user? → Choose transparency.
2. Does one approach inject hidden behavior? → Choose the unbiased option.
3. Does one approach hardcode assumptions? → Choose the configurable option.
4. Does one approach add dependencies? → Choose the simpler option.

When principles conflict (e.g., transparency vs. simplicity), the higher-numbered principle yields to the lower-numbered one. Transparency and unbiased measurement are the most important.
