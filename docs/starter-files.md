# Starter File Reference Format

**Status:** DRAFT
**Date:** 2025-07-25
**Owner:** Morpheus (Architect)
**Issue:** [#117](https://github.com/ronniegeraghty/hyoka/issues/117)

## 1. Overview

Starter files enable "fix this broken code" and "extend this existing project" style
prompts. Before a Copilot session begins, hyoka copies referenced files into the
agent's workspace so the agent works against a pre-existing codebase rather than
starting from scratch.

This design follows **Waza's `ResourceFile` pattern** — prompt authors declare
resources alongside the prompt, and the evaluation harness materializes them in the
workspace before generation begins.

## 2. Frontmatter Format

### 2.1 Option A: Directory Reference (Recommended)

A single `starter_project` field points to a directory whose entire contents are
copied into the workspace root.

```yaml
---
id: storage-dp-dotnet-fix-retry-logic
service: storage
plane: data-plane
language: dotnet
category: debugging
difficulty: intermediate
description: Fix broken retry logic in an existing Azure Storage client
starter_project: ./starters/
---
```

The directory tree is copied as-is:

```
prompts/storage/data-plane/dotnet/
├── fix-retry-logic.prompt.md
└── starters/
    ├── Program.cs
    ├── StorageApp.csproj
    └── appsettings.json
```

After copy, the agent workspace contains:

```
<workspace>/
├── Program.cs
├── StorageApp.csproj
└── appsettings.json
```

> **Why recommend directory reference?** Most starter projects involve multiple
> interdependent files (source, project file, config). A single directory keeps them
> co-located, avoids long file lists in frontmatter, and mirrors how developers think
> about projects.

### 2.2 Option B: Explicit File List

For prompts that need only one or two files, an explicit `starter_files` list avoids
creating a dedicated directory.

```yaml
---
id: identity-dp-python-fix-auth
service: identity
plane: data-plane
language: python
category: debugging
difficulty: beginner
description: Fix a broken DefaultAzureCredential setup
starter_files:
  - ./main.py
  - ./requirements.txt
---
```

Each path is resolved relative to the prompt file's directory. Files are copied into
the workspace root preserving only the filename (no subdirectory nesting).

### 2.3 Precedence Rules

| Condition | Behavior |
|-----------|----------|
| Both `starter_project` and `starter_files` set | Validation error — use one or the other |
| `starter_project` set | Copy entire directory contents to workspace root |
| `starter_files` set | Copy each listed file to workspace root |
| Neither set | Empty workspace (current default behavior) |

## 3. Path Resolution

All paths are resolved **relative to the prompt file's directory**, matching the
existing implementation in `eval/copilot.go`:

```go
starterDir := p.StarterProject
if !filepath.IsAbs(starterDir) && p.FilePath != "" {
    starterDir = filepath.Join(filepath.Dir(p.FilePath), starterDir)
}
```

Absolute paths are supported but discouraged — they break portability across machines.

> **Security:** `copyDir` already skips symlinks and logs a warning. Validation
> should additionally reject paths that escape the `prompts/` tree via `../`
> traversal.

## 4. File Placement

Files are copied into the workspace **before** the Copilot session is created. The
sequence is:

```
1. Create temp workspace directory
2. Copy starter files/project into workspace root
3. List copied files (for metadata)
4. Create Copilot client with workspace as cwd
5. Start session — agent sees files immediately
```

This matches the current implementation. The `starterFiles` list returned from step 3
can be included in `EvalResult` metadata so reviewers know which files were
pre-existing vs. agent-generated.

## 5. Recommended Directory Layout

### Convention

Place starter files in a `starters/` subdirectory next to the prompt file. This keeps
the prompt directory tidy and makes the relationship explicit.

```
prompts/
└── key-vault/
    └── data-plane/
        └── python/
            ├── crud-secrets.prompt.md           # blank project prompt
            ├── fix-error-handling.prompt.md      # starter project prompt
            └── fix-error-handling.starters/
                ├── main.py
                ├── requirements.txt
                └── README.md
```

Naming the directory `<prompt-short-name>.starters/` prevents collisions when
multiple prompts in the same directory each have their own starter files.

### Interaction with `project_context`

| `project_context` | `starter_project` | Meaning |
|--------------------|--------------------|---------|
| `blank` (default)  | not set            | Agent starts from scratch |
| `existing`         | set                | Agent works on pre-existing code |
| `existing`         | not set            | Validation warning — existing context declared but no files provided |
| `blank`            | set                | Validation warning — starter files provided but context says blank |

## 6. Validation

The `hyoka validate` command should check starter file references. Add these checks
to `internal/validate/validate.go`:

### 6.1 Checks for `starter_project`

| Check | Severity | Message |
|-------|----------|---------|
| Path exists and is a directory | Error | `starter_project path does not exist or is not a directory: %s` |
| Path does not escape `prompts/` tree | Error | `starter_project must not reference paths outside prompts/ directory` |
| Directory is not empty | Warning | `starter_project directory is empty: %s` |
| No conflict with `starter_files` | Error | `cannot set both starter_project and starter_files` |

### 6.2 Checks for `starter_files`

| Check | Severity | Message |
|-------|----------|---------|
| Each path exists and is a regular file | Error | `starter file does not exist: %s` |
| Path does not escape `prompts/` tree | Error | `starter file must not reference paths outside prompts/ directory` |
| No duplicate filenames after resolution | Warning | `duplicate filename in starter_files after path resolution: %s` |
| No conflict with `starter_project` | Error | `cannot set both starter_project and starter_files` |

### 6.3 Cross-field consistency

| Check | Severity | Message |
|-------|----------|---------|
| `project_context: existing` without starter files | Warning | `project_context is 'existing' but no starter_project or starter_files provided` |
| `project_context: blank` with starter files | Warning | `project_context is 'blank' but starter files are configured` |

## 7. Struct Changes

The `Prompt` struct in `internal/prompt/types.go` already has `StarterProject`. Add
`StarterFiles` for Option B:

```go
type Prompt struct {
    // ... existing fields ...
    StarterProject  string   `yaml:"starter_project" json:"starter_project,omitempty"`
    StarterFiles    []string `yaml:"starter_files" json:"starter_files,omitempty"`
    // ...
}
```

## 8. Complete Example Prompt

```markdown
---
id: key-vault-dp-python-fix-error-handling
service: key-vault
plane: data-plane
language: python
category: debugging
difficulty: intermediate
description: Fix missing error handling in an Azure Key Vault secrets client
sdk_package: azure-keyvault-secrets
doc_url: https://learn.microsoft.com/python/api/azure-keyvault-secrets/
tags: [error-handling, secrets, existing-project]
created: "2025-07-25"
author: morpheus
project_context:
  type: existing
starter_project: ./fix-error-handling.starters/
timeout: 300
expected_packages:
  - azure-keyvault-secrets
  - azure-identity
---

## Prompt

The project in your workspace contains an Azure Key Vault client that retrieves
secrets but has no error handling. The `main.py` file will crash on network errors
or missing secrets.

Add proper error handling:
1. Wrap Key Vault operations in try/except blocks
2. Handle `ResourceNotFoundError` for missing secrets
3. Handle `HttpResponseError` for service errors
4. Add retry logic using the SDK's built-in retry policy
5. Ensure the client is properly closed in a finally block

## Evaluation Criteria

- Error handling covers all Key Vault operations
- Uses SDK-specific exception types, not bare `except`
- Retry policy is configured on the client
- Client cleanup is guaranteed (context manager or finally)
- Existing functionality is preserved
```

## 9. Implementation Roadmap

| Step | Task | Blocked by |
|------|------|------------|
| 1 | Add `StarterFiles []string` field to `Prompt` struct | — |
| 2 | Update `eval/copilot.go` to handle `starter_files` (copy individual files) | Step 1 |
| 3 | Add validation checks to `validate/validate.go` | Step 1 |
| 4 | Add tests for new validation rules | Step 3 |
| 5 | Create first starter-project prompt as proof of concept | Steps 1–3 |
| 6 | Update `docs/prompt-authoring.md` with starter file guidance | Step 5 |
