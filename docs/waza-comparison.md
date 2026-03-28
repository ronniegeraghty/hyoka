# Hyoka ↔ Waza Comparison

> Comprehensive comparison between **Hyoka** (Azure SDK prompt eval tool) and
> **[Waza](https://github.com/microsoft/waza)** (general-purpose eval framework).
> Created for issue [#18](https://github.com/ronniegeraghty/hyoka/issues/18) to
> ease potential future migration.

---

## 1. Executive Summary

| Dimension | Hyoka | Waza |
|-----------|-------|------|
| **Purpose** | Azure SDK code-generation quality evaluation | General-purpose AI skill evaluation |
| **Language** | Go | TypeScript / Node.js |
| **Copilot SDK** | v0.2.0 (system CLI) | v0.1.32 (embedded binary) |
| **Scoring** | Criteria-based pass/fail with majority voting | Grader-based 0–1 scores with aggregation |
| **Review model count** | 3-model panel (configurable) | Single LLM judge |
| **Build verification** | 9 language compilers (dotnet, python, js-ts, go, java, rust, cpp, …) | None |
| **Domain focus** | Azure SDK (100+ prompts, service-specific criteria) | Domain-agnostic |
| **Config format** | YAML (`configs/*.yaml`) | YAML (`eval.yaml`) |
| **Maturity** | Production-internal (1 primary contributor) | Rapid OSS iteration (5+ contributors) |

**Strategic position (per team decision 2026-07-26):** Hyoka focuses on Azure SDK
domain expertise. Waza owns general eval infrastructure. Maintain compatibility
where possible; don't compete on infra.

---

## 2. Feature-by-Feature Comparison

### 2.1 Evaluation Pipeline

| Feature | Hyoka | Waza | Gap |
|---------|-------|------|-----|
| Code generation | Copilot session with model + skills + MCP | Copilot session via skill execution | Aligned |
| Verification | Dedicated Copilot session (requirements check) | N/A (grader handles) | Hyoka-only |
| Review / Grading | 3-reviewer panel → majority voting → consolidation | Single `prompt` grader or custom graders | **Migration blocker** |
| Build verification | Language-specific compilers (9 targets) | None | Hyoka-only |
| Tool usage tracking | `ExpectedTools` vs actual `ToolCalls` | Not tracked | Hyoka-only |

### 2.2 Scoring & Criteria

| Feature | Hyoka | Waza | Gap |
|---------|-------|------|-----|
| Scoring model | Pass/fail per criterion → ratio score | 0–1 float per grader → weighted aggregate | Conceptually mappable |
| General criteria | 5 fixed (Code Builds, Latest Packages, Best Practices, Error Handling, Code Quality) | None (user-defined graders) | Different approach |
| Tiered criteria | 3 tiers: general rubric → attribute-matched → prompt-specific | 2 tiers: spec-level + task-level graders | Partially aligned |
| Majority voting | 3 reviewers, first model consolidates | Not supported | **Migration blocker** |
| Criterion granularity | Named criteria with pass/fail + justification | Grader name + float score + optional annotation | Mappable with wrapper |

### 2.3 Configuration & Skills

| Feature | Hyoka | Waza | Gap |
|---------|-------|------|-----|
| Config file | `configs/*.yaml` (multiple per project) | `eval.yaml` (one per skill) | Different scope |
| Skill resolution | `type: local` (glob) / `type: remote` (npx fetch) | `skill_directories` in eval.yaml | Aligned concept |
| MCP servers | `mcp_servers:` on generator/reviewer config | `mcp_servers:` field exists but **not plumbed** in execution | Waza gap |
| Multi-config runs | Yes — run N configs × M prompts | Single eval.yaml per run | Hyoka feature |
| Plugin system | `plugins:` with hooks (pre/post tool use) | Not supported | Hyoka-only |

### 2.4 Reporting & Trends

| Feature | Hyoka | Waza | Gap |
|---------|-------|------|-----|
| Report format | JSON + HTML + Markdown | JSON summary | Hyoka richer |
| Trend analysis | `hyoka trends` — cross-run regression detection | Not built-in | Hyoka-only |
| Report server | `hyoka serve` — local HTTP viewer | Not built-in | Hyoka-only |
| Session transcripts | Full event logs per reviewer | Execution logs | Similar concept |

### 2.5 CLI Interface

| Command | Hyoka equivalent | Waza equivalent |
|---------|-----------------|-----------------|
| Run evals | `hyoka run` | `waza run` / `waza run <name>` |
| List prompts | `hyoka list` | N/A |
| Validate config | `hyoka validate` | `waza check` |
| Show configs | `hyoka configs` | N/A |
| View trends | `hyoka trends` | N/A |
| Serve reports | `hyoka serve` | N/A |
| Check environment | `hyoka check-env` | N/A |
| Manifest | `hyoka manifest` | N/A |

---

## 3. Config Format Comparison

### 3.1 Hyoka Config (`configs/*.yaml`)

```yaml
configs:
  - name: baseline/claude-opus-4.6
    description: "Baseline — raw Copilot with Claude Opus 4.6"

    generator:
      model: "claude-opus-4.6"
      skills:
        - type: remote
          name: azure-keyvault-py
          repo: microsoft/skills
        - type: local
          path: "./skills/generator"
      mcp_servers:
        azure:
          type: local
          command: npx
          args: ["-y", "@azure/mcp@latest"]
          tools: ["*"]

    reviewer:
      models:
        - "claude-opus-4.6"      # consolidator
        - "gemini-3-pro-preview"
        - "gpt-4.1"
      skills:
        - type: local
          path: "./skills/reviewer"
```

### 3.2 Waza eval.yaml

```yaml
skill: keyvault-dp-python
description: "Evaluate KeyVault data-plane Python code generation"

tasks:
  - name: create-secret
    file: tasks/create-secret.yaml

graders:
  - name: code-quality
    type: prompt
    prompt: "Evaluate the generated code for correctness and best practices."
  - name: api-usage
    type: prompt
    prompt: "Check that the correct Azure SDK APIs are used."

metrics:
  aggregate: weighted_average
  weights:
    code-quality: 0.6
    api-usage: 0.4
```

### 3.3 Key Format Differences

| Aspect | Hyoka | Waza | Alignment path |
|--------|-------|------|----------------|
| **Top-level structure** | `configs:` array (multi-config) | Single skill eval | Hyoka could export per-config eval.yaml |
| **Model specification** | `generator.model` + `reviewer.models[]` | Inherited from Copilot SDK config | Different layer |
| **Skill reference** | `skills: [{type, name, repo, path}]` | `skill:` name + `skill_directories` | Could normalize to Waza's `skill_directories` |
| **Graders vs criteria** | Criteria YAML + rubric.md → pass/fail | `graders:` array with type + prompt → float | Criteria → grader mapping possible |
| **Tasks** | Prompts directory with frontmatter | `tasks:` array referencing task YAML files | Could generate task files from prompts |
| **Metrics** | Criterion pass-rate (passed/total) | `metrics.aggregate` with weights | Float conversion straightforward |
| **MCP servers** | Fully plumbed in generator/reviewer | Defined but not executed | Blocked on Waza |

---

## 4. Grading / Criteria Deep Dive

### 4.1 Hyoka's Three-Tier Criteria

```
Tier 1: General rubric (5 criteria — always applied)
  ├── Code Builds
  ├── Latest Packages
  ├── Best Practices
  ├── Error Handling
  └── Code Quality

Tier 2: Attribute-matched criteria (from criteria/*.yaml)
  └── Matched by prompt metadata: language, service, plane, category
      Example: java.yaml (10 criteria), python.yaml (5), key-vault.yaml (3)

Tier 3: Prompt-specific criteria (from prompt frontmatter)
  └── Custom criteria defined per prompt
```

Each criterion produces: `{name, passed: bool, justification: string}`

### 4.2 Waza's Grader System

```
Graders (defined in eval.yaml):
  ├── type: prompt    — LLM judge with custom prompt
  │     └── Calls set_waza_grade_pass / set_waza_grade_fail
  ├── type: exact     — Exact string match
  ├── type: includes  — Substring check
  └── type: custom    — User-defined function

Each grader produces: {name, score: 0.0–1.0, annotation?: string}
```

### 4.3 Mapping Hyoka Criteria → Waza Graders

| Hyoka criterion | Waza grader mapping | Notes |
|----------------|---------------------|-------|
| Code Builds | `type: custom` (run compiler) | Waza has no built-in build step |
| Latest Packages | `type: prompt` with version-check prompt | Requires domain knowledge |
| Best Practices | `type: prompt` with best-practices prompt | Direct mapping |
| Error Handling | `type: prompt` with error-handling prompt | Direct mapping |
| Code Quality | `type: prompt` with quality prompt | Direct mapping |
| Prompt-specific criteria | `type: prompt` per criterion | 1:1 mapping, but no majority voting |

**Conversion formula:** Hyoka pass/fail → Waza score: `passed ? 1.0 : 0.0`

---

## 5. Skill Handling Comparison

| Aspect | Hyoka | Waza |
|--------|-------|------|
| Skill definition | SKILL.md + references/ directory | SKILL.md (agentskills.io spec) |
| Skill discovery | Config-driven (`skills:` list) | `.waza.yaml` workspace matching or `--discover` |
| Skill scoping | Generator-only, reviewer-only, or shared | Single scope per eval |
| Remote skills | `type: remote` → auto-fetch via `npx skills add` | Manual or pre-installed |
| Local skills | `type: local` with glob patterns | `skill_directories` paths |
| Plugin bundling | Plugins bundle skills + MCP + hooks | Not supported |

### Layout Comparison

**Hyoka project layout:**
```
hyoka/
├── configs/          # eval configurations
├── prompts/          # evaluation prompts (with frontmatter)
├── skills/           # skill definitions
│   ├── generator/    # generator-scoped skills
│   └── reviewer/     # reviewer-scoped skills
├── criteria/         # tiered criteria YAML files
├── templates/        # report templates
└── reports/          # evaluation output
```

**Waza project layout (skill-centric):**
```
waza-project/
├── .waza.yaml        # workspace config
├── skills/
│   └── keyvault-dp-python/
│       ├── SKILL.md
│       └── eval.yaml
└── evals/
    └── keyvault-dp-python/
        └── tasks/
            └── create-secret.yaml
```

---

## 6. Migration Blockers

These Hyoka features have **no Waza equivalent** and would block a full migration:

### 6.1 Multi-Model Review Panel (Critical)
Hyoka runs 3 independent reviewer models in parallel; the first model
consolidates results via majority voting per criterion. Waza uses a single
LLM judge per grader.

**Workaround:** Define 3 separate `prompt` graders (one per reviewer model),
but Waza provides no aggregation/voting mechanism. A custom post-processing
step would be needed.

### 6.2 Build Verification (High)
Hyoka's `--verify-build` invokes language-specific compilers for 9 targets.
Waza has no build step.

**Workaround:** Implement as `type: custom` grader that shells out to
compilers. Feasible but must be maintained outside Waza's core.

### 6.3 Dedicated Verification Session (Medium)
Hyoka runs a separate Copilot session to verify generated code meets
requirements before grading. Waza has generation → grading only.

**Workaround:** Could model as a pre-grading `prompt` grader, but loses
the independent session isolation that prevents self-bias.

### 6.4 Multi-Config Matrix Runs (Medium)
Hyoka runs N configs × M prompts in a single invocation. Each Waza
eval.yaml covers a single config.

**Workaround:** Script multiple `waza run` invocations. Loses unified
reporting.

### 6.5 Plugin Lifecycle Hooks (Low)
Hyoka's plugin system supports `pre_tool_use` / `post_tool_use` hooks.
Waza has no hook mechanism.

### 6.6 Resource Monitoring (Low)
`--monitor-resources` tracks CPU/memory per process. No Waza equivalent.

---

## 7. What Already Aligns

These areas are already compatible or trivially mappable:

1. **Skill format** — Both use SKILL.md following agentskills.io conventions.
2. **YAML configuration** — Both use YAML; field-level mapping is straightforward.
3. **Copilot SDK** — Same underlying SDK (version delta: v0.2.0 vs v0.1.32).
4. **Prompt → Task mapping** — Hyoka prompts can be converted to Waza task YAML files.
5. **Criteria → Grader mapping** — Hyoka pass/fail criteria map to `prompt` graders with binary scores.
6. **Local skill directories** — Both support directory-based skill references.
7. **CLI paradigm** — Both use `run` as primary command; `validate`/`check` for verification.

---

## 8. Recommendations

### 8.1 Changes Hyoka Could Adopt Now (Low Risk)

| Change | Effort | Benefit |
|--------|--------|---------|
| Add `hyoka export --format waza` command | Medium | Generate eval.yaml + task files from hyoka configs + prompts |
| Adopt Waza's grader naming in criteria output | Low | Use `score: 0.0/1.0` alongside `passed: bool` in reports |
| Support `eval.yaml` as alternative config input | Medium | Read Waza eval.yaml and map to internal ToolConfig |
| Align task file format with Waza's task YAML | Low | Add optional `tasks/` directory alongside prompts |

### 8.2 Convention Alignment (No Code Changes)

| Convention | Current Hyoka | Waza | Recommendation |
|------------|--------------|------|----------------|
| Score representation | `passed: bool` | `score: float` | Add float field to `CriterionResult` (0.0/1.0) |
| Grader/criterion naming | `CriterionResult` | `GraderResult` | Keep hyoka naming; add type alias for export |
| Config file name | `configs/*.yaml` | `eval.yaml` | Support `eval.yaml` as recognized config name |
| Workspace discovery | CLI flags / project config | `.waza.yaml` | Support `.waza.yaml` as fallback discovery |

### 8.3 What NOT to Change

| Feature | Reason to keep |
|---------|---------------|
| Multi-model review panel | Core differentiator; Waza can't replicate it |
| Build verification | Domain-critical for Azure SDK (compilation correctness) |
| Three-session pipeline | Prevents self-bias in verification |
| Tiered criteria system | Azure SDK domain expertise encoded in criteria |
| Plugin/hook system | Extends eval flexibility beyond what Waza offers |

### 8.4 Future Migration Triggers

Consider migrating to Waza **if and when**:

1. Waza adds multi-model review with aggregation/voting
2. Waza plumbs `mcp_servers` into execution (currently defined but unused)
3. Waza adds build verification or custom pre-grading steps
4. Waza's Copilot SDK version catches up to v0.2.0+

Until then, maintain hyoka as the primary tool with Waza-compatible export
as a bridge for teams that want to use Waza's infrastructure.

---

## Appendix: Quick Reference

### Config Field Mapping

| Hyoka field | Waza equivalent | Notes |
|-------------|----------------|-------|
| `configs[].name` | `skill:` | Waza scopes to skill name |
| `configs[].description` | `description:` | Direct mapping |
| `generator.model` | (SDK config) | Waza inherits model from Copilot SDK |
| `generator.skills` | `skill_directories:` | Hyoka richer (remote + local + glob) |
| `generator.mcp_servers` | `mcp_servers:` | Waza defined but not plumbed |
| `reviewer.models` | N/A | **No Waza equivalent** |
| `reviewer.skills` | N/A | Waza has single skill scope |
| Criteria YAML files | `graders:` | Criteria → grader conversion needed |
| Prompt frontmatter | `tasks/*.yaml` | Frontmatter → task file export |
| `metrics` (implicit ratio) | `metrics.aggregate` | Add explicit metrics block |

### CLI Command Mapping

| Hyoka | Waza | Notes |
|-------|------|-------|
| `hyoka run --config X --prompt Y` | `waza run [name]` | Different scope model |
| `hyoka validate` | `waza check` | Similar purpose |
| `hyoka list` | `waza run --discover` (partial) | Different output |
| `hyoka trends` | N/A | No Waza equivalent |
| `hyoka serve` | N/A | No Waza equivalent |
| `hyoka configs` | N/A | No Waza equivalent |
| `hyoka check-env` | N/A | No Waza equivalent |
| `hyoka manifest` | N/A | No Waza equivalent |
| `hyoka rerender` | N/A | No Waza equivalent |
