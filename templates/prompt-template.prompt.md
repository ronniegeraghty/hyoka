---
id: <service>-<dp|mp>-<language>-<category-slug>
service: # storage | key-vault | cosmos-db | event-hubs | app-configuration | purview | digital-twins | identity | resource-manager | service-bus
plane: # data-plane | management-plane
language: # dotnet | java | js-ts | python | go | rust | cpp
category: # authentication | pagination | polling | retries | error-handling | crud | batch | streaming | auth | provisioning
difficulty: # basic | intermediate | advanced
description: >
  One to three sentences describing what this prompt tests.
sdk_package: # e.g., Azure.Storage.Blobs
api_version: # e.g., "2024-11-04"
doc_url: # https://learn.microsoft.com/...
tags: []
created: # YYYY-MM-DD
author: # GitHub username
# expected_tools: []  # Optional: tool names the generation session should use (e.g., create_file, run_terminal_command, azure_mcp)
# starter_project: "" # Optional: path to starter project directory (relative to prompt file). Used with project_context: existing.
# project_context:    # Optional: "blank" (default) starts from empty workspace; "existing" copies starter_project first.
---

# <Title>: <Service> (<Language>)

## Prompt

Write the exact prompt text here. Be specific about what you're asking the agent
to accomplish using the Azure SDK.

## Expected Coverage

The generated code should demonstrate:
- Key API usage the prompt tests
- Expected packages or imports
- Configuration or setup steps
- Error handling patterns

## Context

Why this prompt matters and what code generation quality aspect it evaluates.
