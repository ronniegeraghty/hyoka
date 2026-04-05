# Grader Config YAML Schema

**Status:** DRAFT  
**Issue:** #103 (Task 1.2a)  
**Decisions:** DM3 (gate semantics), DM4 (typed result fields), DM19 (one model per prompt grader)

## Overview

The grader config schema replaces the three-tier criteria system with a flat,
composable list of typed graders. Each grader is a single concern — file
existence, build success, LLM review — evaluated independently and aggregated
by weighted scoring. Hard constraints use `gate: true` to override weighted
averages (DM3).

### Design Principles

1. **Deterministic where possible** — Use file checks, build commands, and
   schema validators instead of LLMs when the answer is objective.
2. **One concern per grader** — Each grader checks exactly one thing.
3. **Composable via weights** — The framework aggregates; graders don't know
   about each other.
4. **Explicit over implicit** — No tier merging, no hidden precedence rules.

## Root Schema

```yaml
graders:
  - kind: <string>              # Required: grader type (see Grader Kinds below)
    name: <string>              # Required: human-readable name, unique within config
    config: <object>            # Required: kind-specific configuration (see below)
    weight: <float>             # Optional: scoring weight 0.0–1.0 (default: 1.0)
    gate: <bool>                # Optional: hard pass/fail (default: false) — DM3
    when: <map[string]string>   # Optional: property-based applicability conditions
```

### Field Reference

| Field    | Type                | Required | Default | Description                                                                |
|----------|---------------------|----------|---------|----------------------------------------------------------------------------|
| `kind`   | string              | yes      | —       | Grader type identifier. One of the six supported kinds.                    |
| `name`   | string              | yes      | —       | Human-readable name. Must be unique within the grader list.                |
| `config` | object              | yes      | —       | Kind-specific configuration. Schema varies by `kind`.                      |
| `weight` | float64             | no       | 1.0     | Scoring weight for weighted average. Normalized across all graders.        |
| `gate`   | bool                | no       | false   | When true, failure causes the entire evaluation to fail (DM3).             |
| `when`   | map\[string\]string | no       | —       | Property conditions for applicability. All entries must match (AND logic).  |

### Applicability (`when`)

The `when` field uses exact case-insensitive string matching against prompt
properties. All specified keys must match for the grader to apply. If `when`
is omitted, the grader applies to all prompts.

```yaml
when:
  language: python        # Only apply to Python prompts
  service: key-vault      # AND only Key Vault service
```

This replaces the `MatchCondition` struct from the criteria system. Property
keys are not restricted to a fixed set — any prompt property can be used.

### Gate Semantics (DM3)

When `gate: true`, the grader acts as a hard constraint. If a gate grader
fails, the entire evaluation fails regardless of the weighted average from
other graders. Use gates for objective requirements:

- File must exist → `gate: true`
- Code must compile → `gate: true`
- No forbidden tools used → `gate: true`

Non-gate graders contribute to the weighted score only.

## Grader Kinds

### `file` — File Existence and Content Check

Checks whether a generated file exists and optionally matches a content
pattern. Fully deterministic — no LLM involved.

```yaml
- kind: file
  name: "main_file_exists"
  config:
    path: "main.py"              # Required: file path relative to work dir
    pattern: "def main"          # Optional: regex pattern to match in content
    must_exist: true             # Optional: default true
  weight: 1.0
  gate: true
```

| Config Field  | Type   | Required | Default | Description                                       |
|---------------|--------|----------|---------|---------------------------------------------------|
| `path`        | string | yes      | —       | File path relative to the generated output dir.    |
| `pattern`     | string | no       | —       | Regex pattern that must match file content.         |
| `must_exist`  | bool   | no       | true    | Whether the file must exist. False inverts check.   |

**Result:** Pass if file exists (or doesn't, per `must_exist`) and content
matches `pattern` (if specified).

---

### `program` — External Command Execution

Runs an external command (compiler, linter, test suite) and grades based on
exit code. Deterministic for build/lint checks.

```yaml
- kind: program
  name: "builds_successfully"
  config:
    command: "python"                # Required: executable to run
    args: ["-m", "py_compile", "main.py"]  # Optional: command arguments
    timeout: 30                      # Optional: seconds (default: 30)
  weight: 1.0
  gate: true
```

| Config Field | Type     | Required | Default | Description                                    |
|--------------|----------|----------|---------|------------------------------------------------|
| `command`    | string   | yes      | —       | Executable to run.                              |
| `args`       | []string | no       | —       | Command-line arguments.                         |
| `timeout`    | int      | no       | 30      | Maximum execution time in seconds.              |

**Result:** Pass if exit code is 0. Stdout/stderr captured in result details.

---

### `prompt` — LLM-as-Judge Review

Sends generated code to a single LLM model for rubric-based review. Each
`prompt` grader uses exactly one model (DM19). For multi-model review,
configure multiple `prompt` graders with different models and weights.

```yaml
# Multi-model review via separate grader instances (DM19)
- kind: prompt
  name: "code_quality_opus"
  config:
    model: "claude-opus-4.6"         # Required: single model per instance
    rubric: |                        # Required: evaluation instructions
      Evaluate the code for correctness, SDK best practices,
      error handling, and documentation quality.
  weight: 0.7

- kind: prompt
  name: "code_quality_sonnet"
  config:
    model: "claude-sonnet-4.5"
    rubric: |
      Evaluate the code for correctness, SDK best practices,
      error handling, and documentation quality.
  weight: 0.3
```

| Config Field | Type   | Required | Default | Description                                    |
|--------------|--------|----------|---------|------------------------------------------------|
| `model`      | string | yes      | —       | LLM model to use for review.                   |
| `rubric`     | string | yes      | —       | Evaluation instructions sent to the model.      |

**Result:** Score from LLM response, plus per-criterion pass/fail breakdown,
summary, issues, and strengths.

---

### `behavior` — Session Behavior Constraints

Validates the generator's session behavior — which tools were used, how many
turns were taken. Checks against the session event log.

```yaml
- kind: behavior
  name: "uses_azure_mcp"
  config:
    required_tools: ["azure-mcp"]    # Optional: tools that must be called
    forbidden_tools: ["rm", "sudo"]  # Optional: tools that must NOT be called
    max_turns: 25                    # Optional: max generation turns allowed
  weight: 1.0
  gate: true
  when:
    service: key-vault
```

| Config Field      | Type     | Required | Default | Description                                   |
|-------------------|----------|----------|---------|-----------------------------------------------|
| `required_tools`  | []string | no       | —       | Tools that must appear in session events.      |
| `forbidden_tools` | []string | no       | —       | Tools that must NOT appear in session events.  |
| `max_turns`       | int      | no       | —       | Maximum number of generation turns allowed.    |

**Result:** Pass if all constraints are satisfied. Details include which
tools were found/missing.

---

### `action_sequence` — Ordered Action Verification

Verifies that the generator performed actions in a specific order. Checks
the session event log for ordered subsequence matching.

```yaml
- kind: action_sequence
  name: "read_before_write"
  config:
    expected_actions:                # Required: ordered list of expected actions
      - "read_file"
      - "edit_file"
  weight: 0.5
```

| Config Field       | Type     | Required | Default | Description                                     |
|--------------------|----------|----------|---------|-------------------------------------------------|
| `expected_actions` | []string | yes      | —       | Ordered list of actions (subsequence match).     |

**Result:** Pass if all expected actions appear in order in the session log.
Other actions may appear between expected actions (subsequence, not exact).

---

### `tool_constraint` — Fine-Grained Tool Usage Rules

Enforces specific tool usage constraints — required tools, forbidden tools,
and call count bounds. More granular than `behavior` for tool-specific rules.

```yaml
- kind: tool_constraint
  name: "mcp_tool_limits"
  config:
    required: ["azure-mcp"]          # Optional: tools that must be called
    forbidden: ["dangerous-tool"]    # Optional: tools that must NOT be called
    min_calls: 1                     # Optional: min total tool calls
    max_calls: 50                    # Optional: max total tool calls
  weight: 1.0
```

| Config Field | Type     | Required | Default | Description                                     |
|--------------|----------|----------|---------|-------------------------------------------------|
| `required`   | []string | no       | —       | Tools that must be called at least once.         |
| `forbidden`  | []string | no       | —       | Tools that must never be called.                 |
| `min_calls`  | int      | no       | —       | Minimum total tool invocations required.         |
| `max_calls`  | int      | no       | —       | Maximum total tool invocations allowed.          |

**Result:** Pass if all constraints are satisfied. Details include observed
tool call counts.

## Complete Example

```yaml
graders:
  # Gate: file must exist
  - kind: file
    name: "main_file_exists"
    config:
      path: "main.py"
    weight: 1.0
    gate: true
    when:
      language: python

  # Gate: code must compile
  - kind: program
    name: "compiles_successfully"
    config:
      command: "python"
      args: ["-m", "py_compile", "main.py"]
      timeout: 30
    weight: 1.0
    gate: true
    when:
      language: python

  # Multi-model LLM review (DM19)
  - kind: prompt
    name: "code_review_opus"
    config:
      model: "claude-opus-4.6"
      rubric: |
        Evaluate the generated code for:
        1. Correctness — does it solve the prompt requirements?
        2. SDK best practices — proper use of Azure SDK patterns
        3. Error handling — appropriate try/except with specific exceptions
        4. Documentation — clear docstrings and comments where needed
    weight: 0.7

  - kind: prompt
    name: "code_review_sonnet"
    config:
      model: "claude-sonnet-4.5"
      rubric: |
        Evaluate the generated code for:
        1. Correctness — does it solve the prompt requirements?
        2. SDK best practices — proper use of Azure SDK patterns
        3. Error handling — appropriate try/except with specific exceptions
        4. Documentation — clear docstrings and comments where needed
    weight: 0.3

  # Behavioral: must use Azure MCP, must not exceed turn limit
  - kind: behavior
    name: "azure_tool_usage"
    config:
      required_tools: ["azure-mcp"]
      max_turns: 25
    weight: 1.0
    gate: true
    when:
      service: key-vault

  # Sequence: should read prompt context before editing
  - kind: action_sequence
    name: "reads_before_writes"
    config:
      expected_actions:
        - "read_file"
        - "edit_file"
    weight: 0.5

  # Tool constraint: bounded tool usage
  - kind: tool_constraint
    name: "reasonable_tool_usage"
    config:
      forbidden: ["rm"]
      max_calls: 50
    weight: 0.5
```

## Scoring Algorithm

1. **Filter** — Select graders where all `when` conditions match the prompt properties.
2. **Execute** — Run each applicable grader independently.
3. **Gate check** — If any grader with `gate: true` fails, the evaluation fails (score = 0).
4. **Weighted average** — For non-gate graders (and passing gates):
   `final_score = Σ(grader_score × weight) / Σ(weight)`

## Migration from Criteria System

| Criteria System              | Grader System                              |
|------------------------------|--------------------------------------------|
| `MatchCondition` struct      | `when: map[string]string`                  |
| `Criterion{Name,Description}`| `prompt` grader with `rubric`              |
| Tier 1 (general rubric)      | `prompt` grader with no `when` conditions  |
| Tier 2 (attribute-matched)   | `prompt` grader with `when` conditions     |
| Tier 3 (prompt-specific)     | `prompt` grader or inline rubric extension |
| `MergeCriteria()`            | Removed — no merging needed                |
| `FormatCriteria()`           | Removed — rubric is grader-native          |

## Go Types

See `hyoka/internal/graders/types.go` for the Go type definitions that
implement this schema. The types use dual `yaml`/`json` struct tags
consistent with the existing `config` package patterns.
