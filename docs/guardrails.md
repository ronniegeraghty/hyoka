# Guardrails and Safety

hyoka enforces several guardrails to prevent runaway evaluations and protect the host environment.

## Generator Guardrails

These limits apply per-evaluation and abort the run if exceeded:

| Guardrail | Default | Flag | Description |
|-----------|---------|------|-------------|
| Max Turns | 25 | `--max-turns` | Maximum assistant message turns. Prevents agents from looping indefinitely. |
| Max Files | 50 | `--max-files` | Maximum files generated. Prevents agents from creating excessive output. |
| Max Output Size | 1 MB | `--max-output-size` | Total size of all generated files. Accepts `KB`, `MB`, `GB` suffixes. |
| Max Session Actions | 50 | `--max-session-actions` | Maximum actions (reasoning, response, or tool call) per Copilot session. |

When a guardrail triggers:
- The evaluation is marked as **failed**
- `guardrail_abort_reason` is populated in the report
- The specific limit exceeded is logged at `warn` level

## Safety Boundaries

By default, hyoka enforces safety boundaries preventing generated code from provisioning real Azure resources.

| Flag | Description |
|------|-------------|
| `--allow-cloud` | Disables safety boundaries, allowing real Azure resource provisioning |
| `--sandbox` | (Default) Enforces safety boundaries |

## Fan-Out Confirmation

When running more than 10 evaluations, hyoka prompts for confirmation:

```
⚠️  Large run detected (42 evaluations). Continue? [y/N]
```

Use `-y` or `--yes` to skip the confirmation (for CI). Use `--all-configs` when running multiple configs.

## Process Lifecycle

hyoka tracks all spawned Copilot processes to ensure cleanup:

- **PID registration**: Child Copilot processes are registered with the ProcessTracker
- **Graceful shutdown**: On completion or Ctrl+C, processes receive SIGTERM then SIGKILL
- **Orphan detection**: After cleanup, hyoka scans for orphaned Copilot processes and terminates them
- **Validation**: Use `--strict-cleanup` to fail the run if any orphans are found (CI mode)

## Resource Monitoring

Use `--monitor-resources` to track CPU and memory usage of Copilot sessions:

- Samples all tracked PIDs every 5 seconds
- Records peak CPU % and peak memory MB per evaluation
- Includes aggregate stats in the run summary
- Logs warnings when thresholds are exceeded (e.g., >2 GB RAM per process)

## Troubleshooting

### Orphaned Processes

If Copilot processes persist after a run:

```bash
# Use the built-in clean command (recommended)
hyoka clean

# Or find and kill manually by PID
ps aux | grep 'copilot.*headless'
kill <PID>
```

Use `--strict-cleanup` in CI to detect this automatically.

### Timeouts

Each phase has an independent timeout:

| Phase | Default | Flag |
|-------|---------|------|
| Generation | 600s (10 min) | `--generate-timeout` |
| Build | 300s (5 min) | `--build-timeout` |
| Review | 300s (5 min) | `--review-timeout` |

If a phase times out, the report includes `error_category: "timeout"` with details.

### Zero Prompts Found

If no prompts are found:
1. Check the `--prompts` directory path
2. Ensure files use `.prompt.md` extension
3. Run `hyoka validate` to check for naming issues
4. Check `hyoka list` to see discovered prompts

### Debug Logging

Enable detailed logging to diagnose issues:

```bash
# Log to stderr
hyoka run --config baseline --log-level debug

# Log to a file
hyoka run --config baseline --log-level debug --log-file eval.log
```

When `--log-level` is `debug` or `info` with `--progress auto`, live progress is automatically disabled to prevent output conflicts.
