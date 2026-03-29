# Prompt Authoring Guide

Prompts are Markdown files with YAML frontmatter that define evaluation scenarios for AI code generation.

## File Naming

Prompts must use the `.prompt.md` extension:

```
prompts/
  storage/
    data-plane/
      python/
        crud/
          blob-upload.prompt.md
```

## Frontmatter Schema

```yaml
---
id: storage-dp-python-crud           # Unique identifier (required)
service: storage                      # Azure service name (required)
plane: data-plane                     # data-plane or management-plane (required)
language: python                      # Programming language (required)
category: crud                        # Use-case category (required)
difficulty: basic                     # basic, intermediate, advanced (required)
description: "Upload a blob..."       # Short description
sdk_package: azure-storage-blob       # Primary SDK package
doc_url: https://learn.microsoft.com/...  # Reference documentation
tags:
  - blob
  - upload
created: "2026-01-15"
author: your-name
expected_packages:                    # Packages the code should use
  - azure-storage-blob
  - azure-identity
expected_tools:                       # Tools the agent should invoke
  - create_file
  - run_terminal_command
starter_project: ./starter/           # Optional starter project directory
reference_answer: ./reference/        # Optional reference answer directory
timeout: 300                          # Prompt-specific timeout (seconds)
---
```

### Required Fields

- `id` — Unique prompt identifier (used in reports and filtering)
- `service` — Azure service (e.g., `storage`, `key-vault`, `cosmos-db`)
- `plane` — `data-plane` or `management-plane`
- `language` — `python`, `java`, `dotnet`, `go`, `javascript`, `typescript`
- `category` — Use-case category (e.g., `crud`, `auth`, `encryption`)
- `difficulty` — `basic`, `intermediate`, `advanced`

## Prompt Body

After the frontmatter, use Markdown sections:

```markdown
# Storage Blob Upload (Python)

## Prompt

Write a Python script that uploads a file to Azure Blob Storage
using the azure-storage-blob SDK with DefaultAzureCredential.

## Evaluation Criteria

- Uses BlobServiceClient with DefaultAzureCredential
- Creates container if it doesn't exist
- Uploads file with proper content type detection
- Handles BlobAlreadyExistsError gracefully

## Notes

Optional notes for the prompt author (not sent to the agent).
```

### Sections

| Section | Purpose |
|---------|---------|
| `## Prompt` | The actual prompt sent to the AI agent (required) |
| `## Evaluation Criteria` | Prompt-specific criteria for the reviewer (Tier 3) |
| `## Notes` | Author notes (not used in evaluation) |

## Tiered Evaluation Criteria

hyoka supports three tiers of evaluation criteria:

1. **Tier 1 — General** (always applied): 5 general criteria from the rubric (Code Builds, Latest Package Versions, Best Practices, Error Handling, Code Quality)

2. **Tier 2 — Attribute-Matched** (conditional): YAML files in a `criteria/` directory that activate based on prompt language, service, or other metadata. For example, `criteria/language/java.yaml` adds Java-specific criteria to all Java prompts.

3. **Tier 3 — Prompt-Specific** (per-prompt): The `## Evaluation Criteria` section in each prompt file.

All tiers are merged and sent to the reviewer as a unified criteria list. Use `--criteria-dir criteria/` to enable Tier 2 criteria.

### Criteria YAML Format

```yaml
# criteria/language/python.yaml
match:
  language: python
criteria:
  - name: DefaultAzureCredential Usage
    description: Authentication uses DefaultAzureCredential.
  - name: Context Manager for Clients
    description: SDK clients are used with `with` statements.
```

## Near-Miss Detection

If no prompts are found, hyoka scans for common naming mistakes:

- `auth-prompt.md` → should be `auth.prompt.md` (dot, not hyphen)
- `auth.prompt.txt` → should be `auth.prompt.md` (wrong extension)
- `*.md` files with YAML frontmatter → may need `.prompt.md` suffix

## Creating New Prompts

Use the interactive scaffolding command:

```bash
hyoka new-prompt
```

This asks for service, language, plane, category, and difficulty, then generates a properly structured prompt file.
