# Eval Tool Cleanup — Implementation Plan

> **Produced by:** Gimli (Azure SDK Tools Agent PM)
> **Requested by:** Ronnie Geraghty
> **Date:** 2025-07-28
> **Status:** Investigation complete — awaiting approval to implement

---

## Item 1: Make Tool Callable from Repo Root

**Current state:** The tool already has smart path resolution (`main.go:52-81`) that tries `./prompts` → `../prompts`, `./configs/all.yaml` → `../configs/all.yaml`, etc. Running `go run ./hyoka` from repo root works today. A pre-built binary exists at `hyoka/azsdk-prompt-eval`.

**What needs to change:**
- Add a `Makefile` at repo root with targets: `build`, `install`, `test`, `lint`, `run`
- Add a `go.work` file at repo root so `go run ./hyoka` works without `cd hyoka/`
- Optionally add a thin wrapper script (`./eval` or `./run-eval.sh`) for quick invocation
- Update README to recommend repo-root invocation as the primary method

**Decisions for Ronnie:**
- Preferred invocation style: `make run ARGS="--service storage"` vs `./eval run --service storage` vs `go run ./hyoka ...`?
- Should the pre-built binary stay committed to the repo, or build on demand?

**Complexity:** S (Small) — path resolution already works; just need convenience wrappers

---

## Item 2: AI Analysis Default for Trends + Auto-Run After Evals

**Current state:**
- `trends` command has `--analyze` flag defaulting to `false` (`main.go:489`)
- After `run` completes (`main.go:281-294`), it simply returns — no automatic trends analysis
- `trends.Generate()` and `trends.AnalyzeTrends()` are already callable programmatically (`internal/trends/trends.go:85`, `internal/trends/analysis.go:14`)

**What needs to change:**

**(a) Flip the default for `--analyze`:**
- Change `main.go:489` from `BoolVar(&analyze, "analyze", false, ...)` to `BoolVar(&analyze, "analyze", true, ...)`
- Add `--no-analyze` flag (or use `--analyze=false`) for users who want to skip
- Update the `Long` description on the `trends` command

**(b) Auto-run trends after evals:**
- At the end of `runCmd()` (after line ~294), call `trends.Generate()` with the current run's reports directory
- Then call `trends.AnalyzeTrends()` on the result
- Add a `--skip-trends` flag to `run` for opting out
- Print a clear separator before trends output so it's visually distinct

**Decisions for Ronnie:**
- Should auto-trends after `run` be the default, or opt-in with `--trends`?
- Should `--no-analyze` be the flag name, or `--analyze=false` (cobra BoolVar supports both)?

**Complexity:** S (Small) — the functions exist; just wiring and flag changes

---

## Item 3: HTML Report Redesign — Show Agent Steps Clearly

**Current state:**
- Individual HTML reports (`html.go:329-685`) already show:
  - Tool Calls section with cards (lines 505-539) — index, name, MCP server, status, duration, collapsible args/results
  - Session Transcript with raw events (lines 661-682)
  - Generation Session with prompt, reasoning, reply (lines 479-503)
- `SessionEventRecord` (`types.go:8-21`) captures: type, tool_name, tool_args, content, error, tool_result, tool_success, duration_ms, mcp_server_name, mcp_tool_name, file_path
- `buildReportData()` (`html.go:172-221`) extracts `ToolAction` structs from session events

**What needs to change:**

**(a) Restructure report around agent workflow phases:**
- **Phase 1: Generation** — Show the generating agent's step-by-step journey: prompt → reasoning → tool calls → file writes → final reply. Currently these are separate sections; merge into a chronological timeline view.
- **Phase 2: Verification** — Show the reviewing agent's process: what it checked, tools it used, what it found. Currently verification is a simple pass/fail box.
- **Phase 3: Build** — Show build verification results with full command output.

**(b) Timeline/step view:**
- Replace the flat tool-call list with a chronological timeline showing interleaved reasoning + tool calls
- Each step shows: agent thought → tool invoked → result received → next thought
- Color-code by phase (generation = blue, review = green, build = orange)

**(c) Verifier improvements (bigger scope):**
- The verifier should attempt to build the code using language-appropriate tools
- Check that referenced SDK packages are real and current versions
- This touches `internal/verify/` and `internal/build/` — significant new logic

**Decisions for Ronnie:**
- Is the timeline view the right UX, or would a two-panel (generation | verification) layout be better?
- How much verifier enhancement is in scope vs a separate workstream?
- Should the verifier use skills (MCP tools) to check library versions, or direct API calls?

**Complexity:** L (Large) — HTML template redesign + potential verifier logic changes

---

## Item 4: Improve Prompt File Organization

**Current state:**
- **79 prompts** across 7 services, 2 planes, 7 languages, 10 categories
- **Actual structure:** `prompts/{service}/{plane}/{language}/{category-name}.prompt.md` (NOT `.../{category}/prompt.md` as originally assumed)
- Distribution is uneven: storage (22), identity (21), key-vault (14), cosmos-db (10), event-hubs/app-config/service-bus (4 each)
- Rich frontmatter with 14 fields (id, service, plane, language, category, difficulty, description, sdk_package, api_version, doc_url, tags, created, author, plus optional fields like expected_tools, expected_packages, starter_project, reference_answer, project_context, timeout)

**Assessment:** The current layout is reasonable and well-structured. The path hierarchy mirrors the frontmatter fields, making it predictable. Alternative layouts considered:

| Layout | Pros | Cons |
|--------|------|------|
| **Current:** `{service}/{plane}/{lang}/file.prompt.md` | Mirrors Azure SDK org; browsable by service | Deep nesting; language-specific comparison requires cross-tree navigation |
| **By language first:** `{lang}/{service}/{category}/file.prompt.md` | Easy per-language eval runs | Doesn't mirror Azure org structure |
| **Flat with metadata:** `prompts/{id}.prompt.md` | Simple; rely on frontmatter for filtering | Loses directory-browsable organization |
| **By category first:** `{category}/{service}/{lang}/file.prompt.md` | Groups similar patterns | Unusual; doesn't match filtering patterns |

**Recommendation:** Keep current structure. The tool's filtering (`--service`, `--language`, etc.) already makes any grouping queryable regardless of directory layout. Reorganizing 79 files would break existing report paths and run history with minimal gain.

**What could improve instead:**
- Add a `prompts/README.md` index showing the distribution matrix
- Add a `manifest` command output that generates a browsable index (already exists: `manifest` command)
- Consider adding prompt templates for each category to standardize new prompt creation

**Decisions for Ronnie:**
- Is the current structure actually causing pain, or is this fine as-is?
- Would a generated index/matrix file be sufficient?

**Complexity:** S (Small) if keeping current structure + adding index; L (Large) if reorganizing

---

## Item 5: CLI Command Usability Improvements

**Current state:** 8 commands registered in `rootCmd()` (main.go:39-47):

| Command | Key Flags |
|---------|-----------|
| `run` | `--prompts`, `--service`, `--language`, `--plane`, `--category`, `--tags`, `--prompt-id`, `--config`, `--config-file`, `--workers`, `--timeout`, `--model`, `--output`, `--skip-tests`, `--skip-review`, `--verify-build`, `--debug`, `--dry-run`, `--stub`, `--progress` |
| `list` | Same filter flags as `run` |
| `configs` | `--config-file` |
| `validate` | `--prompts` |
| `check-env` | (none) |
| `trends` | `--prompt-id`, `--service`, `--language`, `--reports-dir`, `--output`, `--analyze` |
| `report` | `[run-id]`, `--reports-dir`, `--all` |
| `version` | (none) |

**What needs to change:**

**(a) Missing convenience features:**
- `run` has no `--all` flag — running all prompts requires no filter flags (which is the default), but an explicit `--all` flag would make intent clear and prevent accidental full runs
- No command aliases (e.g., `ls` for `list`, `env` for `check-env`)
- No `--json` output flag for `list` command (useful for scripting)
- `run --dry-run` is good but could also show estimated duration based on historical data

**(b) Flag naming review:**
- `--skip-tests` and `--skip-review` are clear ✓
- `--verify-build` is opt-in (good — builds can be slow) ✓
- `--stub` is developer-facing; could be renamed to `--mock` for clarity
- `--progress auto|live|log|off` is good ✓

**(c) Suggested additions:**
- `run --all` — explicit "run everything" flag (currently default behavior, but ambiguous)
- `run --continue <run-id>` — resume a failed/partial run
- `run --rerun-failed <run-id>` — re-run only failed evals from a previous run
- `list --json` — machine-readable output
- `trends --open` — auto-open HTML report in browser after generation

**Decisions for Ronnie:**
- Which of these additions are worth doing now vs later?
- Should `run` with no filters require `--all` to prevent accidental full runs?
- Is `--continue`/`--rerun-failed` a priority?

**Complexity:** S-M (Small to Medium) — individual flag additions are small; `--continue`/`--rerun-failed` is medium

---

## Item 6: Reorganize Run Results

**Current state:**
- Results stored at: `reports/{runID}/results/{service}/{plane}/{language}/{category}/{config}/`
- Each eval produces: `report.json`, `report.html`, `report.md`, `generated-code/` directory
- Summaries at: `reports/{runID}/summary.{json,html,md}`
- Path constructed in `generator.go:13-24` using `ReportDir()` function
- Trends stored separately in `reports/trends/`

**Ronnie's idea:** Store per-prompt results near the prompt files, summary in results dir.

**Analysis:**

| Approach | Pros | Cons |
|----------|------|------|
| **Current: Centralized** (`reports/{run}/...`) | Clean separation of source vs output; easy to `.gitignore` all results; simple cleanup; trends analysis scans one directory | Deep nesting; hard to see a prompt's history across runs |
| **Co-located** (`prompts/{service}/.../results/{run}/{config}/`) | Easy to see one prompt's full history; natural grouping | Pollutes prompt directory with large generated files; harder to `.gitignore`; breaks trends scanning; complicates prompt validation |
| **Hybrid** (summaries centralized + symlinks or index to per-prompt results) | Best of both worlds conceptually | Complex implementation; fragile symlinks; confusing |

**Recommendation:** Keep centralized results but add a **per-prompt history view**:
- Add a `history` command: `hyoka history --prompt-id <id>` that scans all runs for a given prompt and shows its pass/fail timeline
- This gives the "see a prompt's history" benefit without moving files
- The `trends` command already partially does this — enhance it for single-prompt deep-dive

**Alternatively**, add a `results/` directory inside each prompt folder containing only a `latest.json` symlink or summary file (not full results), keeping the bulk in `reports/`.

**Decisions for Ronnie:**
- Is the pain point "I can't see one prompt's history" or "results are too far from prompts"?
- Would a `history` command solve the need without file reorganization?
- If co-location is strongly desired, are you okay with `.gitignore` complexity?

**Complexity:** S (Small) for history command; M (Medium) for hybrid approach; L (Large) for full reorganization

---

## Item 7: More Info in Summary Reports

**Current state:**
- `RunSummary` struct (`types.go:69-82`): run_id, timestamp, total_prompts, total_configs, total_evaluations, passed, failed, errors, duration_seconds, report_paths, results
- Summary HTML (`html.go:687-805`): stats cards, prompt×config matrix (status/score/duration/files/tool_calls), detailed results table
- Individual `EvalReport` objects are embedded in results array with full session events

**What's missing:**

**(a) Duration breakdown:**
- Average/median/p95 duration per config
- Average/median/p95 duration per prompt
- Slowest and fastest evaluations highlighted
- Time spent in generation vs verification vs build phases

**(b) Config comparison:**
- Side-by-side pass rates per config
- Score distribution per config (histogram or box plot)
- Tool usage differences between configs
- Which prompts pass on one config but fail on another (delta analysis)

**(c) Tool usage statistics:**
- Aggregate tool call counts across all evals
- Most/least used tools
- Tool success/failure rates
- MCP server utilization breakdown

**(d) Quality metrics aggregation:**
- Average review scores by dimension (if LLM-as-judge scores are captured)
- Score distribution across all evals
- Per-category/per-language quality breakdown

**(e) Error analysis:**
- Error type classification and frequency
- Common failure patterns
- Prompts that consistently fail across configs

**Suggested implementation approach:**
1. Add a `SummaryStats` struct with computed aggregates
2. Compute stats from the `Results` array before rendering
3. Add new sections to summary HTML template: duration breakdown, config comparison chart, tool usage table, error patterns

**Decisions for Ronnie:**
- Which of (a)-(e) are highest priority?
- Should charts be interactive (requires JS library) or static HTML tables?
- Is the config comparison (b) the most impactful addition?

**Complexity:** M (Medium) — data is already in the results array; need computation + template additions

---

## Item 8: Docs Update

**Current state:**
- `README.md` (root): Quick start, commands table (6 of 8 commands), filtering examples, config docs, adding prompts workflow, repo structure, tagging, roadmap
- `hyoka/README.md`: More detailed version with full flag tables, smart path detection, code review section
- `docs/eval-tool-plan.md`: Architecture/implementation plan (983 lines)
- **No other docs files**

**Gaps identified:**

| Gap | Severity |
|-----|----------|
| `trends` command undocumented in root README | High |
| `report` command undocumented in root README | High |
| No dedicated getting-started guide | Medium |
| No architecture overview for contributors | Medium |
| Root README references `manifest` command but it's not in code's command list | Medium (stale reference) |
| Scoring dimensions (1-10 scale) not explained in root README | Low |
| Reference answers feature not fully explained | Low |
| Optional frontmatter fields not clearly marked | Low |

**What needs to change:**
1. **Add `trends` and `report` to root README commands table** — quick fix
2. **Remove stale `manifest` reference** or clarify if it was renamed
3. **Create `docs/getting-started.md`** — walkthrough from clone to first eval run
4. **Create `docs/architecture.md`** — package structure, data flow, extension points
5. **Sync root README and hyoka README** — they have diverged; consider having root point to hyoka/README for details
6. **Document scoring dimensions** in root README or link to tool README section

**Decisions for Ronnie:**
- Should root README be a thin pointer to hyoka/README, or a standalone doc?
- Is a getting-started guide worth the maintenance burden, or is the current inline quick start sufficient?
- Should docs/ contain rendered examples of reports?

**Complexity:** M (Medium) — several files to update/create; no code changes

---

## Suggested Implementation Order

```
Phase 1 — Quick Wins (1-2 days)
├── Item 2: Flip --analyze default + auto-trends after run     [S]
├── Item 8a: Add trends/report commands to README              [S]
└── Item 1: Add Makefile + go.work at repo root                [S]

Phase 2 — Usability (2-3 days)
├── Item 5: Add --all, --json, command aliases                  [S-M]
├── Item 7: Add duration breakdown + config comparison to       [M]
│           summary reports
└── Item 8b: Create getting-started guide + architecture doc    [M]

Phase 3 — Major Enhancements (1-2 weeks)
├── Item 3: HTML report redesign with timeline view             [L]
└── Item 6: Add history command (or hybrid results layout)      [S-M]

Deferred / As-Needed
└── Item 4: Prompt reorganization (recommend keeping current    [S]
            structure + adding index)
```

**Dependencies:**
- Item 8 (docs) should be updated incrementally as other items land
- Item 2 (auto-trends) should land before Item 7 (summary improvements) since they share report infrastructure
- Item 3 (HTML redesign) is independent but large — can be parallelized with Phase 1-2 work
- Item 6 (results reorg) depends on Ronnie's decision about co-location vs history command

---

## Open Decisions Summary

| # | Decision | Options | Impact |
|---|----------|---------|--------|
| 1 | Invocation style | Makefile / wrapper script / go.work | Item 1 |
| 2 | Auto-trends default | Opt-in `--trends` vs opt-out `--skip-trends` | Item 2 |
| 3 | Report timeline UX | Chronological timeline vs two-panel layout | Item 3 |
| 4 | Verifier scope | Skills-based build checking now vs later | Item 3 |
| 5 | Prompt reorg | Keep current + index vs full reorganization | Item 4 |
| 6 | Safety flag for run | Require `--all` to run everything? | Item 5 |
| 7 | Resume/rerun priority | `--continue`/`--rerun-failed` now or later? | Item 5 |
| 8 | Results co-location | History command vs hybrid vs full move | Item 6 |
| 9 | Summary chart style | Static HTML tables vs interactive JS charts | Item 7 |
| 10 | README structure | Thin root + detailed hyoka/ vs standalone root | Item 8 |

---

## Item 9: Progress Display — Append-on-Start (No Waiting Lines)

**Current state:** Display pre-allocates N lines for all evals (including pending ones showing "⏳ (waiting)"). Uses DECSC/DECRC cursor save/restore to redraw the fixed region.

**What needs to change:**
- Remove `evalPending` state and pre-allocated lines entirely
- Only add a line to the display when an eval actually starts (EventStarting)
- Lines grow from top down — new evals append at the bottom of the active stack
- Keep a summary counter at the bottom: `  3/12 completed  ✅ 2  ❌ 1  2.3m`
- Completed evals stay visible with their final status (no removal)
- The ANSI cursor region grows dynamically as evals start

**Complexity:** M (Medium) — display.go rewrite to dynamic append model

**Phase:** 1 (Quick Wins) — this directly impacts daily UX

---

## Item 10: AI Skill for Writing Prompt Files

**Current state:** No tooling to help users author new prompt files. Contributors must read the frontmatter schema from examples or the plan doc. The 14-field frontmatter is error-prone to write by hand.

**What needs to change:**
- Create a Copilot skill (`skills/prompt-authoring/`) that helps users write new prompt files
- The skill should know: frontmatter schema, valid field values, naming conventions, category types, existing prompt examples
- Could be implemented as: (a) a `new-prompt` CLI command that scaffolds interactively, (b) a Copilot skill SKILL.md for agent-assisted authoring, or (c) both
- The CLI scaffolder would ask: service, language, plane, category → generate the file with correct path and populated frontmatter
- The skill would include prompt-writing best practices: how to write clear instructions, what makes a good eval prompt, reference answer patterns

**Complexity:** M (Medium) — CLI scaffolder + skill doc + validation

**Phase:** 2 (Usability)
