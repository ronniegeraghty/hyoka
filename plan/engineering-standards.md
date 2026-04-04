# Hyoka Engineering Standards

**Date:** 2026-10-14  
**Author:** Morpheus (Lead/Architect)  
**Source:** Codebase audit findings + established patterns  
**Status:** Active

These standards are based on patterns observed in the existing hyoka codebase. They reflect how the project already works, codified for consistency as the team grows.

---

## 1. Error Handling

### Rules

- **Wrap with `%w`** — All `fmt.Errorf` calls must use `%w` for error wrapping. This preserves the error chain for `errors.Is()` and `errors.As()`.
- **Imperative voice** — Error messages use imperative voice: `"open config file: %w"`, not `"failed to open config file: %w"` or `"could not open config file: %w"`.
- **No log-and-return** — Never log an error and then return it. Either log it (terminal handling) or return it (propagation). Doing both creates duplicate noise.
- **No panics in production code** — `panic()` is forbidden outside of test code. Use `return err` instead.
- **No silent discard** — Don't use `_ = someFunc()` for errors that could affect output quality. At minimum, log with `slog.Warn`.
- **`os.Exit` only in main** — `os.Exit()` calls are acceptable only in `main()`, validation commands, and emergency signal handlers.

### Examples

```go
// ✅ Good
if err := loadConfig(path); err != nil {
    return fmt.Errorf("load config %s: %w", path, err)
}

// ❌ Bad — log-and-return
if err := loadConfig(path); err != nil {
    slog.Error("failed to load config", "error", err)
    return err
}

// ❌ Bad — missing %w
if err := loadConfig(path); err != nil {
    return fmt.Errorf("failed to load config: %v", err)
}

// ❌ Bad — silent discard
_ = os.Remove(tempFile)  // Use slog.Warn if removal matters
```

### Known Exceptions

- `filepath.Abs()` errors are discarded in several places (`main.go:219,263,270,286`, `fetcher.go:68,82,88,120`). These almost never fail on valid paths but should be logged at debug level for defense-in-depth.

---

## 2. Logging

### Rules

- **slog only** — Use `log/slog` (Go 1.21+) for all diagnostic logging. No `log.Printf`, no `fmt.Println` for diagnostics, no third-party logging libraries.
- **Role-based loggers** — When logging inside a subsystem, create a logger with a `role` attribute: `slog.With("role", "evaluator")`. This enables log filtering by component.
- **User-facing output to stdout/stderr** — Progress bars, results summaries, and interactive output go directly to stdout/stderr. These are not logs.
- **Structured attributes** — Use key-value pairs, not string interpolation: `slog.Info("config loaded", "name", cfg.Name, "model", cfg.Model)`.
- **Log levels:**
  - `Debug` — Verbose diagnostic info (enabled with `--log-level debug`)
  - `Info` — Normal operational events
  - `Warn` — Recoverable issues (degraded behavior, fallbacks, discarded non-fatal errors)
  - `Error` — Unrecoverable issues (will cause the operation to fail)

### Examples

```go
// ✅ Good
logger := slog.With("role", "reviewer")
logger.Info("review complete", "model", model, "score", score)

// ❌ Bad — fmt.Println for diagnostics
fmt.Println("Review complete:", model, score)

// ❌ Bad — unstructured
slog.Info(fmt.Sprintf("Review complete: %s scored %d", model, score))
```

---

## 3. Testing

### Rules

- **Table-driven tests** — Use `[]struct{ name string; ... }` with `t.Run(tc.name, ...)` for parameterized tests. This is the Go standard pattern.
- **`t.TempDir()` for temp files** — Never create temp files manually. `t.TempDir()` auto-cleans on test completion.
- **Stubs over mocks** — Prefer simple struct implementations of interfaces over mock frameworks. Hyoka uses no mock libraries.
- **No `time.Sleep` for assertions** — Timing-based assertions are flaky. Use channels, mutexes, or polling with timeouts instead.
- **Test file naming** — `*_test.go` in the same package for unit tests. Use `_test` package suffix only when testing exported API from a consumer's perspective.
- **Test function naming** — `TestFunctionName_Scenario` (e.g., `TestLoadConfig_MissingFile`, `TestEngine_Run_GuardrailExceeded`).

### Coverage Expectations

All new packages must have tests. Current coverage baseline:
- `eval/` — 57 tests (excellent)
- `config/` — comprehensive edge case coverage
- `serve/` — includes path traversal tests
- `pidfile/` — **zero tests** (must be addressed)
- `review/` — 5 tests for 732 lines (thin — needs improvement)

### Examples

```go
// ✅ Good — table-driven
func TestParsePrompt(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Prompt
        wantErr bool
    }{
        {name: "valid prompt", input: validYAML, want: expectedPrompt},
        {name: "missing id", input: noIDYAML, wantErr: true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := ParsePrompt(tc.input)
            if tc.wantErr {
                if err == nil { t.Fatal("expected error") }
                return
            }
            if err != nil { t.Fatal(err) }
            if got != tc.want { t.Errorf("got %v, want %v", got, tc.want) }
        })
    }
}
```

---

## 4. CLI Conventions

### Rules

- **`RunE` not `Run`** — All cobra commands use `RunE` (returns error) not `Run` (returns nothing). This enables proper error propagation.
- **`PersistentPreRunE` for init** — Cross-cutting concerns (logging setup, config loading) belong in `PersistentPreRunE` on the root command.
- **Flag groups** — Related flags are grouped logically. Use cobra's flag grouping when available.
- **Safety boundaries** — Destructive operations require explicit opt-in flags (e.g., `--allow-cloud`). Safe by default.
- **`--dry-run` support** — Commands that perform actions should support `--dry-run` to preview without executing.
- **Help text** — Every flag has a description. Defaults are shown in help output.

### Flag Conventions

- Boolean flags: `--flag` (no value needed)
- String flags with defaults: `--flag value` with default shown in help
- Comma-separated multi-value: Must be quoted on CLI (`--config "a,b,c"`)
- Singular names: `--prompt-id` (one value), `--config` (can be comma-separated)

---

## 5. Configuration

### Rules

- **`Normalize()` is idempotent** — Config normalization can be called multiple times with the same result. This is critical for legacy migration.
- **`Effective*()` getters** — When a config field has defaults or fallbacks, use `Effective*()` methods (e.g., `EffectiveModel()`) that resolve the final value.
- **Validate post-parse** — Validation runs after YAML parsing and normalization, not during. This separates parsing concerns from business rules.
- **`KnownFields(true)`** — Use strict YAML parsing to catch field typos. Unknown fields cause parse errors.
- **No silent shadowing** — Duplicate config names, duplicate prompt IDs, and ambiguous references must produce errors, not silent behavior.

### Config Structure

```yaml
configs:
  - name: "team/model-variant"    # Slash-separated: team/description
    description: "Human-readable"
    generator:
      model: "model-name"          # Required — empty must fail validation
      system_prompt: ""            # Optional — default empty
      skills: [...]
      mcp_servers: {...}
    reviewer:
      models: [...]                # Multi-model panel
      system_prompt: ""            # Optional — default empty
      skills: [...]
```

---

## 6. Dependencies

### Rules

- **Minimal footprint** — New dependencies require explicit justification. The current 3 direct deps (Copilot SDK, Cobra, YAML) set the bar.
- **Standard library preferred** — Use `log/slog`, `net/http`, `text/template`, `os/exec`, `encoding/json`, `path/filepath` before reaching for external packages.
- **No HTTP frameworks** — `net/http` with `http.ServeMux` is sufficient.
- **No ORM** — Hyoka doesn't use databases. If it did, `database/sql` would be the choice.
- **No mock frameworks** — Stubs via interfaces.
- **Vendor or module proxy** — Dependencies fetched via Go module proxy. No vendoring currently.

---

## 7. Code Organization

### Rules

- **Packages by domain** — Each package represents a domain concept: `eval`, `review`, `config`, `prompt`, `report`, `serve`, `criteria`, `skills`, `trends`, `clean`, `progress`.
- **Interfaces at boundaries** — Package boundaries use interfaces (`CopilotEvaluator`, `Reviewer`). Concrete types are internal to packages.
- **Acyclic dependency graph** — Package imports must not form cycles. The dependency direction flows: `main → eval → review → report`.
- **One concern per package** — Don't mix config parsing with CLI flag handling, or report generation with trend analysis.
- **Internal only** — All packages live under `hyoka/internal/`. Nothing is exported outside the module.

### File Size Guidelines

Files over 800 lines should be examined for splitting opportunities:
- `report/html.go` (1374 lines) — candidate for template extraction
- `main.go` (1329 lines) — scheduled for split to `cmd/` package
- `eval/engine.go` (1035 lines) — acceptable as core orchestrator
- `trends/trends.go` (857 lines) — complex but cohesive

---

## 8. CI Requirements

### Rules

- **All PRs must pass** — `go build ./hyoka/...`, `go vet ./hyoka/...`, `go test ./hyoka/...`
- **No merge without green CI** — Branch protection rules enforce passing checks.
- **CI runs on every push and PR** — Not just on main, on all branches.
- **Test timeout** — CI tests have a reasonable timeout (5 minutes). Tests that take longer need optimization.

### Workflow Location

`.github/workflows/ci.yml` — currently does not exist (P0 task to create).

---

## 9. Git Workflow

### Branch Naming

```
{username}/issue-{N}-{short-description}
```

Examples: `ronniegeraghty/issue-84-skill-loading`, `copilot/issue-72-auth-check`

### Commit Messages

- Imperative mood: "Add config validation" not "Added config validation"
- Reference issues: "Fix reviewer model bug (#84)"
- Always include trailer:

```
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```

### PR Workflow

1. Create branch from `main`
2. Implement changes with tests
3. Push and open PR
4. CI must pass (build + vet + test)
5. Review required
6. Squash merge to `main`

### Git Identity

```bash
git config user.name "ronniegeraghty"
git config user.email "ronniegeraghty@users.noreply.github.com"
gh auth switch --user ronniegeraghty  # Before pushing
```

---

## 10. Reference Patterns

### microsoft/waza

The Azure SDK team's production agent evaluation tool. Key patterns to follow:

- **Zero system prompt** — SDK session config handles isolation, not prompt injection
- **`ResourceFile` pattern** — Starter files placed in workspace before session starts
- **Config-driven everything** — System prompts, tools, limits all in config YAML
- **Session config–based isolation** — Working directory, tool availability, permissions set via SDK

### Go Standard Patterns

- **Context propagation** — All long-running operations accept `context.Context` as first parameter
- **Functional options** — For complex constructors, use `With*` option functions
- **Error wrapping chain** — Each layer adds context: `"run eval: create session: open config: %w"`
- **Graceful shutdown** — `signal.NotifyContext` + two-phase SIGTERM/SIGKILL
