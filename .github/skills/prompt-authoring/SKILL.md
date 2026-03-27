# Prompt Authoring Skill

You are an expert at authoring evaluation prompts for the **Azure SDK Prompt Evaluation Tool** (`hyoka`). You help users create well-structured `.prompt.md` files that test AI agent code generation quality.

## Prompt File Format

Each prompt is a Markdown file with YAML frontmatter. Files live at:
```
prompts/{service}/{plane}/{language}/{slug}.prompt.md
```

### Frontmatter Schema

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `id` | ✅ | string | Unique ID: `{service}-{dp|mp}-{language}-{slug}` |
| `service` | ✅ | string | Azure service name |
| `plane` | ✅ | string | `data-plane` or `management-plane` |
| `language` | ✅ | string | Target programming language |
| `category` | ✅ | string | Use-case category |
| `difficulty` | ✅ | string | `basic`, `intermediate`, or `advanced` |
| `description` | ✅ | string | 1–3 sentences: what this prompt tests |
| `created` | ✅ | string | Date in `YYYY-MM-DD` format |
| `author` | ✅ | string | GitHub username |
| `sdk_package` | ❌ | string | Expected SDK package (e.g., `Azure.Storage.Blobs`) |
| `doc_url` | ❌ | string | Library reference docs URL (see doc_url convention below) |
| `tags` | ❌ | list | Free-form tags for filtering (e.g., `[identity, getting-started]`) |
| `expected_tools` | ❌ | list | Tool names the agent should use (e.g., `[create_file, run_terminal_command]`) |
| `expected_packages` | ❌ | list | SDK packages the generated code should import |
| `starter_project` | ❌ | string | Path to starter project dir (relative to prompt file) |
| `project_context` | ❌ | string | `blank` (default) or `existing` (copies starter_project first) |
| `reference_answer` | ❌ | string | Inline reference code for LLM-as-judge scoring |
| `timeout` | ❌ | int | Per-prompt timeout in seconds (overrides config default) |

### Valid Values

**Services:** `storage`, `key-vault`, `cosmos-db`, `event-hubs`, `app-configuration`, `purview`, `digital-twins`, `identity`, `resource-manager`, `service-bus`

**Languages:** `dotnet`, `java`, `js-ts`, `python`, `go`, `rust`, `cpp`

**Planes:** `data-plane`, `management-plane`

**Categories:** `authentication`, `pagination`, `polling`, `retries`, `error-handling`, `crud`, `batch`, `streaming`, `auth`, `provisioning`

**Difficulties:** `basic`, `intermediate`, `advanced`

### doc_url Convention

The `doc_url` field should point to the **library's API reference docs**, not quickstarts or tutorials:

| Language | URL Pattern | Example |
|----------|-------------|---------|
| Python | `learn.microsoft.com/en-us/python/api/overview/azure/{pkg}-readme` | `azure/identity-readme` |
| .NET | `learn.microsoft.com/en-us/dotnet/api/overview/azure/{pkg}-readme` | `azure/storage.blobs-readme` |
| Java | `learn.microsoft.com/en-us/java/api/overview/azure/{pkg}-readme` | `azure/cosmos-readme` |
| JS/TS | `learn.microsoft.com/en-us/javascript/api/overview/azure/{pkg}-readme` | `azure/identity-readme` |
| Go | `pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/...` | `sdk/azidentity` |
| Rust | `docs.rs/{crate}/latest/{crate}/` | `azure_identity` |
| C++ | `github.com/Azure/azure-sdk-for-cpp/tree/main/sdk/...` | `sdk/identity/azure-identity` |

### ID Convention

The `id` field follows the pattern: `{service}-{dp|mp}-{language}-{slug}`

- Use `dp` for `data-plane`, `mp` for `management-plane`
- The slug should be a concise kebab-case descriptor of what the prompt tests
- Example: `storage-dp-dotnet-auth`, `cosmos-db-dp-python-crud-items`

### File Naming

The filename should match the slug portion of the ID: `{slug}.prompt.md`

Examples:
- `authentication.prompt.md`
- `pagination-list-blobs.prompt.md`
- `crud-items.prompt.md`

## Prompt Structure

Every `.prompt.md` file should have these sections after the frontmatter:

```markdown
# Title: Service (Language)

## Prompt

The exact prompt text sent to the AI agent. Be specific and actionable.

## Evaluation Criteria

Bullet list of what the generated code should demonstrate.
The review agent uses this (along with the general rubric) to score generated code.

## Context

Why this prompt matters and what quality aspect it evaluates.
(Human-readable only — not used by the eval tool.)
```

**How sections flow through the eval tool:**
- **Generator agent** receives ONLY the `## Prompt` text — no frontmatter, no evaluation criteria
- **Review agent** receives the prompt text, generated code, the general rubric (`hyoka/internal/review/rubric.md`), AND the `## Evaluation Criteria` section
- **`## Context`** is for human readers only — the eval tool ignores it entirely

## Writing Good Prompts

### DO ✅

- **Be specific about the task**: "Create a BlobServiceClient using DefaultAzureCredential" not "Use storage"
- **Specify the SDK package**: "Using the `Azure.Storage.Blobs` NuGet package"
- **Ask for complete, runnable code**: "Show the complete setup including required packages"
- **Include realistic constraints**: "The code should work in a console app targeting .NET 8"
- **Test one concept per prompt**: Focus on authentication, or pagination, or error handling — not all three
- **Set clear expectations in Evaluation Criteria**: List specific APIs, patterns, and imports

### DON'T ❌

- **Don't be vague**: "Write some Azure code" — too broad to evaluate meaningfully
- **Don't test multiple concepts**: A prompt testing auth + pagination + retries is hard to score
- **Don't assume context**: The agent starts from a blank workspace unless `project_context: existing` is set
- **Don't skip the description**: It's used in reports and filtering
- **Don't duplicate existing prompts**: Check `hyoka list` first

### Difficulty Guidelines

| Difficulty | What It Means | Example |
|-----------|---------------|---------|
| `basic` | Single API call, straightforward setup | Authenticate and list blobs |
| `intermediate` | Multiple API calls, error handling, configuration | Upload with retry policy and progress tracking |
| `advanced` | Complex workflows, multiple services, edge cases | Event-driven pipeline with dead-letter handling |

## Writing Good Reference Answers

The `reference_answer` field is optional but valuable — it enables LLM-as-judge scoring with a reference comparison.

- Write the reference as actual working code, not pseudocode
- Include imports, package references, and error handling
- The reference should represent a "good" answer, not a perfect one
- Keep it focused on the prompt's specific task

## Example: Good Prompt

```yaml
---
id: storage-dp-dotnet-auth
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: basic
description: >
  Can the docs help a developer authenticate to Azure Blob Storage
  using DefaultAzureCredential in .NET?
sdk_package: Azure.Storage.Blobs
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/storage.blobs-readme
tags:
  - identity
  - default-azure-credential
  - getting-started
created: 2025-07-27
author: ronniegeraghty
---

# Authentication: Azure Blob Storage (.NET)

## Prompt

How do I authenticate to Azure Blob Storage using DefaultAzureCredential in C#?
I need to create a BlobServiceClient that uses managed identity in production
but falls back to Azure CLI credentials during local development.
Show me the complete setup including required NuGet packages.

## Evaluation Criteria

The generated code should demonstrate:
- Azure.Identity and Azure.Storage.Blobs package setup
- DefaultAzureCredential initialization
- BlobServiceClient creation with token credential
- Basic container/blob operation to verify auth works

## Context

Authentication is the first step for any Azure SDK interaction.
This tests whether the agent produces correct, current auth patterns.
```

## Example: Bad Prompt

```yaml
---
id: storage-dp-dotnet-stuff
service: storage
plane: data-plane
language: dotnet
category: crud
difficulty: basic
description: "test storage"
created: 2025-07-27
author: someone
---

# Storage

## Prompt

Write Azure storage code in C#.
```

**Problems:**
- Vague prompt — "storage code" could mean anything
- No specific SDK package mentioned
- Description is meaningless for reports
- Missing `sdk_package`, `doc_url`, `tags`
- No Evaluation Criteria section
- No Context section

## CLI Scaffolding

You can also use the CLI to scaffold a new prompt interactively:

```bash
go run ./hyoka new-prompt
```

This asks for service, language, plane, category, and difficulty, then generates the file with populated frontmatter at the correct path.

## Validation

After creating a prompt, always validate:

```bash
go run ./hyoka validate
```

This checks frontmatter schema compliance, required fields, and naming conventions.
