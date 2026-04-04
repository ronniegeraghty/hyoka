# Property-based tool filter schema

> Design document for issue [#106](https://github.com/ronniegeraghty/hyoka/issues/106) — task 1.3a

## Problem

Today, `generator.available_tools` and `generator.excluded_tools` are flat string
lists that apply uniformly to every prompt evaluated under a config.  To support
**pairwise testing** (for example, "give Python prompts the Azure MCP tools but
don't give Go prompts `pip-tools`"), we need tool availability that adapts based
on the prompt being evaluated.

## Schema

### New `tools` field on `generator`

```yaml
generator:
  model: claude-opus-4.6
  tools:
    # No condition → always available
    - name: "bash"

    # Include only when the prompt's language is python
    - name: "azure-mcp"
      when:
        language: python

    # Exclude when the prompt's language is go
    - name: "pip-tools"
      exclude_when:
        language: go

    # Multiple conditions are ANDed
    - name: "cosmosdb-mcp"
      when:
        language: python
        service: cosmos-db

  # Legacy flat lists still work (backward compatible)
  available_tools: ["bash", "azure-mcp"]
  excluded_tools: ["dangerous-tool"]
```

### `ToolEntry` definition

| Field          | Type                | Required | Description                                                       |
|----------------|---------------------|----------|-------------------------------------------------------------------|
| `name`         | `string`            | yes      | Tool name (must match SDK tool identifiers)                       |
| `when`         | `map[string]string` | no       | Include this tool only when **all** key-value pairs match         |
| `exclude_when` | `map[string]string` | no       | Exclude this tool when **all** key-value pairs match              |

A `ToolEntry` with **neither** `when` nor `exclude_when` is unconditional (always included).

Setting **both** `when` and `exclude_when` on the same entry is a validation error.

### Matchable prompt properties

These are the `Prompt` struct fields available for matching:

| Key          | Example values                     | Prompt field     |
|--------------|------------------------------------|------------------|
| `language`   | `python`, `dotnet`, `go`, `java`   | `Language`       |
| `service`    | `identity`, `key-vault`, `storage` | `Service`        |
| `plane`      | `data-plane`, `management-plane`   | `Plane`          |
| `category`   | `auth`, `crud`, `pagination`       | `Category`       |
| `difficulty`  | `easy`, `medium`, `hard`           | `Difficulty`     |

All comparisons are case-sensitive, exact string equality.

Multiple conditions within a single `when` or `exclude_when` are **ANDed**:
every key-value pair must match for the condition to apply.

## Resolution logic

`ResolveTools` accepts a `GeneratorConfig` and a set of prompt properties and
returns the final `(availableTools []string, excludedTools []string)` pair that
is passed to the Copilot session.

```
1. Start with the legacy flat lists:
     available = generator.available_tools   (may be nil → "all defaults")
     excluded  = generator.excluded_tools    (may be nil → "exclude nothing")

2. If generator.tools is present, iterate each entry:
     a. Entries with neither when nor exclude_when
        → append name to available
     b. Entries with when:
        → if ALL conditions match the prompt → append name to available
     c. Entries with exclude_when:
        → if ALL conditions match the prompt → append name to excluded

3. Deduplicate both lists.

4. If a tool appears in BOTH available and excluded, excluded wins
   (defense-in-depth: explicit exclusion is always respected).

5. Return (available, excluded). Nil semantics preserved:
     - nil available  = "all default tools"
     - empty available = "zero tools"
```

### Precedence summary

| Scenario                                     | Result           |
|----------------------------------------------|------------------|
| Tool in legacy `available_tools` only        | Available        |
| Tool in legacy `excluded_tools` only         | Excluded         |
| Tool in `tools:` with matching `when:`       | Available        |
| Tool in `tools:` with matching `exclude_when:` | Excluded      |
| Tool in both available and excluded           | **Excluded**     |
| `tools:` not set, legacy lists not set        | All defaults     |

## Backward compatibility

The existing flat `available_tools` and `excluded_tools` fields remain fully
supported.  Configs that don't use the new `tools:` field behave identically to
today.

When both `tools:` and the legacy lists coexist, results are merged (see
resolution logic above).  This allows incremental migration: teams can start
adding conditional entries alongside their existing flat lists.

## Interaction with `always_on: true` (Phase 2)

In Phase 2, pairwise testing will introduce an `always_on: true` flag on
`ToolEntry`.  Tools marked `always_on` bypass `when`/`exclude_when` evaluation
and are always included — useful for baseline tools (like `bash`, `view`,
`edit`) that every prompt needs regardless of properties.

```yaml
tools:
  - name: "bash"
    always_on: true    # Phase 2: never filtered out
  - name: "azure-mcp"
    when:
      language: python
```

The `ResolveTools` function already handles unconditional entries (no
`when`/`exclude_when`), so `always_on` is a semantic alias that also prevents
the tool from being added to the excluded list during pairwise matrix
generation.

## Validation rules

1. Every `ToolEntry` must have a non-empty `name`.
2. `when` and `exclude_when` must not both be set on the same entry.
3. Map keys in `when`/`exclude_when` must be recognized prompt property names
   (`language`, `service`, `plane`, `category`, `difficulty`).
4. Map values must be non-empty strings.

Validation errors are returned from `config.Validate()` during config loading,
consistent with existing validation patterns.

## Go types

See [`hyoka/internal/config/tool_filter.go`](../hyoka/internal/config/tool_filter.go)
for the implementation.

```go
// ToolEntry represents a tool with optional property-based conditions.
type ToolEntry struct {
    Name        string            `yaml:"name"                    json:"name"`
    When        map[string]string `yaml:"when,omitempty"          json:"when,omitempty"`
    ExcludeWhen map[string]string `yaml:"exclude_when,omitempty"  json:"exclude_when,omitempty"`
}

// PromptProperties holds the subset of prompt metadata used for tool filtering.
type PromptProperties struct {
    Language   string
    Service    string
    Plane      string
    Category   string
    Difficulty string
}

// ResolveTools evaluates conditional tool entries against prompt properties and
// merges the result with legacy flat lists. Returns the final available and
// excluded tool slices.
func ResolveTools(gen *GeneratorConfig, props PromptProperties) (available, excluded []string)
```
