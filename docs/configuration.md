# Configuration Guide

hyoka uses YAML configuration files to define evaluation setups. Each config specifies a generator model, reviewer models, skills, and MCP servers.

## Config Directory

By default, configs are loaded from `./configs/`. Use `--config-dir` to specify a different location.

## Config Names vs Filenames

The `name` field in a config is what you pass to the `--config` CLI flag. It is **not** the filename. For example, a config file called `azure-mcp-opus.yaml` might define `name: azure-mcp/claude-opus-4.6`. You'd run it with: `--config azure-mcp/claude-opus-4.6`.

Current config names and their filenames:

| Filename | Config Name |
|----------|-------------|
| `baseline-sonnet.yaml` | `baseline/claude-sonnet-4.5` |
| `baseline-opus.yaml` | `baseline/claude-opus-4.6` |
| `baseline-opus-skills.yaml` | `baseline-skills/claude-opus-4.6` |
| `baseline-codex.yaml` | `baseline/gpt-5.3-codex` |
| `azure-mcp-sonnet.yaml` | `azure-mcp/claude-sonnet-4.5` |
| `azure-mcp-opus.yaml` | `azure-mcp/claude-opus-4.6` |
| `azure-mcp-codex.yaml` | `azure-mcp/gpt-5.3-codex` |

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

### Skills

Skills are Copilot agent instructions packaged as directories containing a `SKILL.md` file. When attached to a generator or reviewer session, the agent receives the skill's content as additional context that guides its behavior.

Skills can be attached to either the **generator** (code generation agent) or the **reviewer** (grading panel agents), or both. They are configured under the `generator.skills` and `reviewer.skills` fields respectively.

#### Skill Types

There are two ways to load skills: **local** (from the filesystem) and **remote** (fetched from a GitHub repository).

##### Local Skills

Local skills reference a directory on disk. Paths can be absolute or relative to the config file's directory.

```yaml
generator:
  model: claude-opus-4.6
  skills:
    - type: local
      path: ./skills/generator
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Must be `"local"` |
| `path` | yes | Path to a skill directory (absolute or relative) |

**Glob patterns** are supported, letting you load multiple skill directories at once. Only directories are included — files are filtered out.

```yaml
generator:
  skills:
    # Loads every subdirectory under ./skills/generator/
    - type: local
      path: "./skills/generator/*"
```

For example, if `./skills/reviewer/` contains three subdirectories (`code-review-comments/`, `reviewer-build/`, `sdk-version-check/`), the pattern `./skills/reviewer/*` expands to all three.

##### Remote Skills

Remote skills are fetched from a GitHub repository using `npx skills add`. They are cached locally under `.skills-cache/` so subsequent runs don't re-download.

```yaml
generator:
  model: claude-sonnet-4.5
  skills:
    - type: remote
      name: azure-keyvault-py
      repo: microsoft/skills
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Must be `"remote"` |
| `repo` | yes | GitHub repository in `owner/repo` format |
| `name` | no | Specific skill name within the repo |

Under the hood, hyoka runs:

```
npx skills add <repo> --directory .skills-cache/<repo>/<name> [--name <name>]
```

#### Generator vs Reviewer Skills

- **Generator skills** guide the code generation agent — for example, providing SDK usage patterns, coding conventions, or language-specific best practices.
- **Reviewer skills** guide the review panel agents — for example, instructing them to add inline review comments or verify builds.

```yaml
configs:
  - name: my-eval/claude-opus-4.6
    generator:
      model: claude-opus-4.6
      skills:
        - type: local
          path: ./skills/generator
        - type: remote
          name: azure-keyvault-py
          repo: microsoft/skills
    reviewer:
      models:
        - claude-opus-4.6
        - gpt-4.1
      skills:
        - type: local
          path: "./skills/reviewer/*"
```

#### Skill Directory Structure

Each skill is a directory containing at minimum a `SKILL.md` file:

```
skills/
├── generator/
│   └── my-sdk-skill/
│       └── SKILL.md        # Instructions for the generator agent
└── reviewer/
    ├── code-review-comments/
    │   └── SKILL.md        # Adds inline REVIEW: comments
    ├── reviewer-build/
    │   └── SKILL.md        # Verifies generated code builds
    └── sdk-version-check/
        └── SKILL.md        # Checks SDK package versions
```

The `SKILL.md` file contains markdown instructions that the Copilot agent receives as context during its session.

#### Legacy Skill Format

Older configs use flat `skill_directories`, `generator_skill_directories`, or `reviewer_skill_directories` fields. These are automatically normalized to the unified `skills` list at load time:

```yaml
# Legacy format (still supported)
configs:
  - name: old-style
    model: claude-opus-4.6
    skill_directories:              # → generator.skills (type: local)
      - ../skills/generator
    reviewer_skill_directories:     # → reviewer.skills (type: local)
      - ../skills/reviewer
```

The precedence for legacy migration is:

1. `generator_skill_directories` → `generator.skills`
2. If absent, `skill_directories` → `generator.skills`
3. `reviewer_skill_directories` → `reviewer.skills`

New configs should use the structured `generator.skills` / `reviewer.skills` format.

### MCP Servers

```yaml
mcp_servers:
  azure:
    type: local
    command: npx
    args: ["-y", "@azure/mcp@latest", "server", "start"]
    tools: ["*"]
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | `"local"` (stdio) or `"sse"` / `"http"` (remote) |
| `command` | yes | Command to launch the MCP server |
| `args` | no | Arguments passed to the command |
| `tools` | no | Tool filter — `["*"]` for all tools, or a list of specific tool names |

> **Important:** The `tools` field must be set (typically `["*"]`) for the MCP server's tools to be registered with the agent. Without it, the server starts but its tools won't be available.

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

Legacy fields (`model`, `reviewer_model`, `reviewer_models`, `skill_directories`, `generator_skill_directories`, `reviewer_skill_directories`, `mcp_servers`, `available_tools`, `excluded_tools`) are automatically normalized to the structured `generator`/`reviewer` format at load time. See the [Skills > Legacy Skill Format](#legacy-skill-format) section for details on how skill directories are migrated.

## Multiple Config Files

Place multiple `.yaml` files in the config directory. All are loaded automatically. Use `--config <name>` to select specific configs, or `--all-configs` to run all.

## Tiered Evaluation Criteria

Use `--criteria-dir` to point to a directory of attribute-matched criteria YAML files. These are matched against prompt metadata (language, service, plane) and merged with prompt-specific criteria at review time. See [prompt-authoring.md](prompt-authoring.md) for details.
