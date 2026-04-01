# Configuration Guide

hyoka uses YAML configuration files to define evaluation setups. Each config specifies a generator model, reviewer models, skills, and MCP servers.

## Config Directory

By default, configs are loaded from `./configs/`. Use `--config-dir` to specify a different location.

## Config File Format

A config file contains one or more named configurations:

```yaml
configs:
  - name: baseline/claude-opus-4.6
    generator:
      model: claude-opus-4.6
      skills:
        - type: local
          path: ../skills/generator
    reviewer:
      models:
        - claude-opus-4.6
        - gpt-5.3-codex
        - claude-sonnet-4.5
      skills:
        - type: local
          path: ../skills/reviewer
```

## Configuration Fields

### Top-Level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Unique name for this configuration |
| `description` | string | no | Human-readable description |

### Generator Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | required | Model to use for code generation |
| `skills` | list | [] | Skill references for the generator session |
| `mcp_servers` | map | {} | MCP server configurations |
| `available_tools` | list | [] | Allowlist of tools the agent can use |
| `excluded_tools` | list | [] | Denylist of tools |

### Reviewer Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | — | Single reviewer model |
| `models` | list | — | Multiple reviewer models (panel review) |
| `skills` | list | [] | Skill references for review sessions |

When `models` lists multiple models, hyoka uses a **panel review** where all models review independently and the first model acts as consolidator to produce a consensus result.

### Skill Types

```yaml
skills:
  # Local skill from filesystem
  - type: local
    path: ../skills/generator

  # Remote skill from GitHub
  - type: remote
    repo: github.com/Azure/ai-hub-sdk
    name: azure-sdk-tools
```

### MCP Servers

```yaml
mcp_servers:
  azure:
    type: sse
    command: npx
    args: ["-y", "@azure/mcp@latest"]
```

## Legacy Format

hyoka also supports a flat legacy format for backward compatibility:

```yaml
configs:
  - name: simple-config
    model: claude-opus-4.6
    reviewer_model: claude-sonnet-4.5
    skill_directories:
      - ../skills/generator
```

Legacy fields are normalized to the structured `generator`/`reviewer` format at load time.

## Multiple Config Files

Place multiple `.yaml` files in the config directory. All are loaded automatically. Use `--config <name>` to select specific configs, or `--all-configs` to run all.

## Tiered Evaluation Criteria

Use `--criteria-dir` to point to a directory of attribute-matched criteria YAML files. These are matched against prompt metadata (language, service, plane) and merged with prompt-specific criteria at review time. See [prompt-authoring.md](prompt-authoring.md) for details.
