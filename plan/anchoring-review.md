# Anchoring Bias Review — Evolution Plan

**Date:** 2026-10-15  
**Reviewer:** Morpheus (Lead/Architect)  
**Scope:** All 5 plan documents reviewed against current codebase and Waza reference architecture  
**Requested by:** Ronnie Geraghty

---

## Summary

The evolution plan has **three significant anchoring biases** that should be corrected, **two moderate ones** worth discussing, and **several areas where the plan genuinely proposes the right approach**. The biggest bias is treating hyoka's review system as something to iterate on rather than something to fundamentally rethink using Waza's grader model.

---

## 1. Review System: LLM-Monolith → Grader Architecture

### What the plan proposes

Keep the current `Reviewer` interface (single Copilot session that reads all files, scores all criteria, returns one JSON blob). Add transparency (FR-05: show per-reviewer reasoning), reviewer tools (FR-17), and configurable reviewer system prompts (FR-18). The `PanelReviewer` consolidation pattern stays.

### How it's anchored to current implementation

The entire review pipeline assumes review = "send everything to an LLM and parse JSON." The `BuildReviewPrompt()` function stuffs the original prompt, all generated files, all reference files, and all criteria into one massive string. The reviewer's system prompt says "Respond with ONLY valid JSON." The panel runs multiple models doing the *same monolithic review* and then consolidates.

This is anchored to the initial "LLM-as-judge" implementation. It works for simple cases but fundamentally limits what review can do.

### Alternative approach: Waza's Grader Architecture

Waza has **12 grader types** — `code`, `prompt`, `text`, `file`, `json_schema`, `program`, `behavior`, `action_sequence`, `skill_invocation`, `trigger`, `diff`, `tool_constraint`. Each grader is:

- **Focused** — one concern per grader (does this file exist? does the code build? did the agent use the right tools?)
- **Typed** — a `program` grader runs a real linter, a `file` grader checks file existence, a `prompt` grader uses LLM-as-judge
- **Composable** — evaluation criteria are a list of graders, each with a weight
- **Deterministic where possible** — file existence, JSON schema validation, and build checks don't need LLMs

Hyoka should adopt this. The current `Reviewer` interface becomes one grader type (`prompt` grader). But you also get:

```yaml
graders:
  - kind: file
    name: "main_file_exists"
    config:
      path: "main.py"
      
  - kind: program
    name: "builds_successfully"
    config:
      command: "python -m py_compile main.py"
      
  - kind: prompt
    name: "code_quality"
    model: "claude-opus-4.6"
    rubric: "Evaluate the code for..."
    weight: 0.5
    
  - kind: behavior
    name: "used_mcp_tools"
    config:
      required_tools: ["azure-mcp"]
```

### Recommendation: **Pivot**

Replace the `Reviewer`/`PanelReviewer` with a `Grader` interface and pluggable grader types. The current LLM review becomes a `prompt` grader. This is the single highest-impact architectural change because it:

1. Makes review deterministic where possible (file checks, build verification)
2. Enables FR-17 (reviewer tools) naturally — a `program` grader IS a tool
3. Makes criteria composable and weighted (Waza's `Weight` field)
4. Eliminates the fragile "parse JSON from LLM output" pattern for non-LLM checks
5. Enables FR-15 (pairwise) analysis more cleanly — graders produce structured scores

### Impact

- `hyoka/internal/review/` → redesign as `hyoka/internal/graders/` with a `Grader` interface and per-type implementations
- `criteria/` format changes — criteria YAML becomes grader config YAML
- `report/types.go` — `ReviewResult` becomes `GraderResults` (simpler, typed)
- Evolution plan Phase 3 (Transparency) simplifies — grader results ARE transparent by design
- FR-05, FR-14, FR-17 merge into the grader architecture rather than being separate features
- Plan sections 1.2, 3.2, and Phase 3 need rewriting

---

## 2. Config System: Normalize() Legacy Dance → Clean Schema

### What the plan proposes

Keep the dual-struct `ToolConfig` with 10+ legacy fields, `Normalize()` migration, and `Effective*()` getters. Decision D6 says "big-bang migration" for prompts but the config system keeps backward compat.

### How it's anchored to current implementation

The plan says big-bang for prompts (87 files, manageable) but doesn't apply the same logic to configs (8 files, even more manageable). The `Normalize()` pattern exists because configs were once flat and got refactored to nested. With only 8 config files, there's no reason to maintain both formats.

Lines 59-71 of `config.go` define 10 legacy fields (`Model`, `ReviewerModel`, `ReviewerModels`, `MCPServers`, `SkillDirectories`, `GeneratorSkillDirectories`, `ReviewerSkillDirectories`, `AvailableTools`, `ExcludedTools`, `Skills`, `Plugins`). Lines 76-128 implement `Normalize()` to migrate them. Lines 131-190 implement 7 `Effective*()` getters to handle both paths.

That's ~130 lines of code (35% of config.go) dedicated to backward compatibility for 8 files.

### Alternative approach

Big-bang migrate the 8 config files to the new `Generator`/`Reviewer` sub-struct format. Delete all legacy fields, `Normalize()`, and `Effective*()` getters. Access `tc.Generator.Model` directly instead of `tc.EffectiveModel()`.

While we're at it, add the new fields the plan needs:
- `generator.system_prompt` (FR-18)
- `reviewer.system_prompt` (FR-18)
- `reviewer.tools` (FR-17)  
- `session_limits` block (FR-08)
- `graders` list (from Finding #1 above)

### Recommendation: **Pivot**

Delete legacy config support in the same Phase 0 sprint as CI. It's 8 files. Write a migration script, update the files, delete ~130 lines of dead code. The plan already decided big-bang for prompts — apply the same logic to configs.

### Impact

- `config.go` shrinks by ~35%
- All `Effective*()` calls in the codebase become direct field access
- `Normalize()` deleted
- Engineering standards §5 (Configuration) needs updating — remove "Normalize() is idempotent" rule
- Plan Phase 1 simplifies — no need to maintain dual formats during transition

---

## 3. Prompt Properties: `map[string]string` Bolt-On → First-Class Struct

### What the plan proposes

Add `Properties map[string]string` to the `Prompt` struct. Keep existing fields (`Service`, `Language`, etc.) as "convenience accessors that read from the properties map." Filter flags query the map.

### How it's anchored to current implementation

This is the classic "preserve the old shape, add a generic escape hatch" pattern. It creates two ways to access the same data — `p.Language` and `p.Properties["language"]` — with the implicit requirement that they stay in sync. The `KnownFields(true)` strictness must be relaxed for the properties section, which undermines the typo-catching benefit.

### Alternative approach

Since we're doing big-bang migration anyway, make properties THE representation. Drop the typed fields entirely:

```go
type Prompt struct {
    ID         string            `yaml:"id"`
    Properties map[string]string `yaml:"properties"`
    PromptText string            `yaml:"-"`
    // ...metadata fields (FilePath, EvaluationCriteria, etc.)
}

// Convenience accessors read from Properties
func (p *Prompt) Language() string { return p.Properties["language"] }
func (p *Prompt) Service() string  { return p.Properties["service"] }
```

Prompt frontmatter becomes:

```yaml
id: key-vault-dp-python-crud
properties:
  service: key-vault
  plane: data-plane
  language: python
  category: crud
  difficulty: medium
  sdk_package: azure-keyvault-secrets
```

This is cleaner than having both `Service string` AND `Properties["service"]`. The convenience methods are still there for ergonomics.

### Recommendation: **Pivot, but moderate**

Use `Properties` as the single source of truth. Add convenience getter methods (not struct fields) for commonly-used properties. This eliminates the sync problem and makes the prompt format truly generic.

The `Filter` struct should also become property-based: `Filters map[string]string` instead of `Service`, `Plane`, `Language`, `Category` fields.

### Impact

- `prompt/types.go` redesigned — fewer struct fields, more methods
- `prompt/parser.go` — frontmatter parsing changes to nested `properties:` key
- All 87 prompts reformatted (already planned, just with different target format)
- `criteria/criteria.go` — `PromptAttrs` and `MatchCondition` become `map[string]string`
- Filter flag handling simplified — generic `--filter key=value` alongside legacy aliases

---

## 4. CLI Split: main.go → cmd/ Package (Moderate Bias)

### What the plan proposes

Split `main.go` into `hyoka/cmd/` package with per-command files. Standard Cobra pattern.

### How it's anchored

The plan treats this as a refactoring task — move code from one file to many files. This is fine mechanically, but it misses an opportunity to rethink the command structure itself.

The current `runCmd()` function (starting at line 124 of main.go) is a 40+ flag monster. Many of those flags exist because the command does too much — it's `run` + `filter` + `configure` + `limit` + `monitor` + `cleanup` in one command.

### Alternative approach

Consider whether the command surface area is right:

- `hyoka run` with 40+ flags → could `hyoka run` take a run specification file instead? Waza uses `eval.yaml` files that encode all this. `hyoka run eval.yaml` is cleaner than `hyoka run --prompt-id X --config Y --max-session-actions 50 --monitor-resources --strict-cleanup --criteria-dir ...`
- The `--config` flag (comma-separated names that reference `name:` inside YAML files, not filenames) is a confusing UX. Waza just points to an eval spec file.

### Recommendation: **Keep the plan, but add a note**

The `cmd/` split is correct and necessary. But add a Phase 2+ task to explore a `hyoka run <spec-file>` pattern that absorbs most flags into a declarative run spec. Don't block the split on this — but don't treat the split as the end of CLI evolution.

### Impact

- Plan task 1.5 stays as-is
- Add a future-looking note about run spec files
- No immediate code changes beyond what's planned

---

## 5. Criteria Tiers → Grader Config (Moderate Bias)

### What the plan proposes

Remove Tier 1 (default criteria). Generalize Tier 2 (attribute-matched) to use property-based matching. Keep Tier 3 (prompt-specific criteria in `## Evaluation Criteria` section).

### How it's anchored

The "tier" concept itself is anchored to the current system where criteria are text blobs that get merged and injected into a review prompt. With a grader architecture (Finding #1), criteria become grader configs — there's no need for tiers because graders are naturally composable.

"Attribute-matched criteria" becomes "graders defined in criteria YAML files with property-based `when:` conditions." "Prompt-specific criteria" becomes "graders defined inline in the prompt file." The merging logic in `MergeCriteria()` disappears — you just collect all applicable graders into a list.

### Recommendation: **Absorb into grader pivot**

If Finding #1 is accepted, the criteria system redesign follows naturally. Don't design criteria as a separate tier system — design it as grader configuration with property-based applicability rules.

### Impact

- `criteria/` package simplifies dramatically or merges into `graders/`
- `MergeCriteria()` and `FormatCriteria()` deleted
- Plan sections 1.2a-c rewritten in terms of grader config
- `MatchCondition` → `when: map[string]string` on grader configs

---

## 6. Report Types — 25+ Fields (Moderate Concern)

### What the plan proposes

Phase 3 adds action timeline and per-reviewer data to reports. Phase 4 builds comparison engine on top of report data.

### How it's anchored

`EvalReport` (report/types.go:77-109) already has 25+ fields including guardrail data, error categories, resource stats, and environment info. Adding action timeline and reviewer transparency will push it past 30 fields. The plan treats the report as an ever-growing struct.

### Alternative approach

With a grader architecture, the report structure simplifies:

```go
type EvalReport struct {
    PromptID     string
    ConfigName   string
    Timestamp    string
    Duration     float64
    Environment  EnvironmentInfo
    Events       []SessionEvent   // Full action timeline
    GraderResults []GraderResult  // Each grader's output
    AggregateScore float64
    Success      bool
    Error        string
}
```

Grader results carry their own typed details — a `program` grader result has build output, a `prompt` grader result has LLM reasoning, a `file` grader result has existence checks. No need for 25+ top-level fields because the details live in the grader results.

### Recommendation: **Redesign alongside grader pivot**

If Finding #1 is accepted, redesign reports to center on `[]GraderResult` instead of spreading review data across top-level fields. This also simplifies the serve site — the API returns structured grader results that the frontend can render by type.

### Impact

- `report/types.go` simplified
- `report/html.go` (1374 lines!) can render grader results generically instead of hardcoding field layouts
- Serve API becomes cleaner — one endpoint shape for all grader types
- Backward compat for existing JSON reports breaks (acceptable for early project)

---

## Areas Where the Plan is NOT Anchored (Good Calls)

### ✅ Zero System Prompt (Plan §1.6)

The plan correctly identifies that the 15 hardcoded rules in `copilot.go:628-655` are bias, not operational necessity. The proposed SDK hooks for file path validation (keeping the `OnPreToolUse` hook but removing the system prompt rules) is exactly right. Waza does this identically. No anchoring bias here.

### ✅ Big-Bang Property Migration (Decision D6)

Rejecting backward compatibility for 87 prompts is the right call. The plan avoids the trap of maintaining dual formats indefinitely. The only refinement (Finding #3) is about the shape of the target format, not the migration strategy.

### ✅ Starter Files / ResourceFile (Plan §1.7)

The plan correctly adopts Waza's `ResourceFile` pattern. The existing `StarterProject` field on `Prompt` is minimal and buggy (`copilot.go:83` silently discards errors). The plan proposes to fix and extend it properly. No anchoring bias — this is genuinely the right approach.

### ✅ `.hyoka` Project Directory (Plan §5.1)

Project-scoped only, structured subdirectories, auto-discovery. This is clean design not anchored to current state (which has no equivalent). Good call.

### ✅ Pairwise Testing (Plan §2.1)

The `--pairwise` flag with `always_on` exemptions is well-designed. This is entirely new functionality with no existing implementation to anchor to. The approach of generating N+1 config variants (baseline + one-tool-removed) is sound.

### ✅ Isolated Workspaces (Plan §2.3)

The plan proposes clean workspace directories per session. This aligns with Waza's `os.MkdirTemp` + cleanup pattern. Current hyoka uses the reports directory as workspace, which leaks state. The proposed change is the right one.

---

## Prioritized Recommendations

| # | Finding | Severity | Action |
|---|---------|----------|--------|
| 1 | Review → Grader architecture | **High** | Redesign review system around pluggable graders |
| 2 | Config legacy cleanup | **Medium** | Big-bang migrate 8 config files, delete Normalize() |
| 3 | Prompt properties as sole representation | **Medium** | Properties map as source of truth, typed fields → methods |
| 5 | Criteria tiers → grader config | **Medium** | Absorbed by Finding #1 |
| 6 | Report struct bloat | **Low** | Absorbed by Finding #1 |
| 4 | CLI run spec file | **Low** | Future enhancement note, don't block current work |

---

## Plan Sections That Need Updating If Recommendations Accepted

### If Finding #1 (Grader Architecture) accepted:
- **Phase 1, §1.2** (Criteria Filters) → Rewrite as grader config format
- **Phase 3, §3.2** (Transparent Review Panel) → Replaced by grader result transparency
- **FR-05** (Transparent Review Panel) → Merge into grader architecture
- **FR-14** (Criteria Filters) → Merge into grader config
- **FR-17** (Reviewer Tools) → Becomes `program` grader type
- **Engineering Standards §5** → Update config schema to include graders
- New package: `hyoka/internal/graders/` with interface + per-type implementations

### If Finding #2 (Config Cleanup) accepted:
- **Phase 0** → Add task: "Big-bang migrate config files, delete legacy fields"
- **Engineering Standards §5** → Remove Normalize() rule
- **Phase 1** → Remove implicit dependency on legacy config support

### If Finding #3 (Prompt Properties) accepted:
- **Phase 1, §1.1** → Revise target format (properties as map, not bolt-on)
- Migration script targets different output format
- `prompt/types.go` redesign more aggressive

---

## Open Questions for Ronnie

1. **Grader architecture:** Does the Waza grader model resonate? Are there grader types hyoka needs that Waza doesn't have? (e.g., SDK version checking, dependency validation)

2. **Config cleanup timing:** Should config migration happen in Phase 0 (alongside CI) or Phase 1 (alongside prompt migration)? Phase 0 is cleaner but adds scope.

3. **Run spec files:** Is the `hyoka run eval.yaml` pattern interesting for the future? Or is the current flag-based approach sufficient for hyoka's use cases?
