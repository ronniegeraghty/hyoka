---
id: storage-dp-python-mcp-best-practices
service: storage
plane: data-plane
language: python
category: best-practices
difficulty: basic
description: >
  Can the agent use the Azure MCP best practices tool
  to generate Azure Blob Storage code following official guidance?
sdk_package: azure-storage-blob
tags:
  - mcp
  - best-practices
  - getting-started
created: 2026-04-02
author: ronniegeraghty
---

# Azure Blob Storage Best Practices (Python)

## Prompt

Before writing any code, you MUST use the Azure MCP tools available in your
environment to look up Azure best practices for code generation. Look for a
tool related to Azure best practices (it may be named something like
`get_azure_bestpractices_get` or similar) and call it first.

Then, following the best practices returned by that tool, write a Python script
that uploads a file to Azure Blob Storage and downloads it back.

Requirements:
- You MUST call an Azure best practices MCP tool first to get guidance before generating code. Do not skip this step.
- Use `DefaultAzureCredential` from `azure-identity` for authentication.
- Use the `azure-storage-blob` SDK package.
- Include proper error handling with `HttpResponseError`.
- Include a `requirements.txt` with pinned package versions.

## Evaluation Criteria

- Agent called an Azure best practices MCP tool before generating code
- Uses `DefaultAzureCredential` for authentication
- Uses `BlobServiceClient` from `azure.storage.blob`
- Includes upload and download operations
- Includes error handling with `HttpResponseError`
- Includes `requirements.txt`

## Context

This prompt tests whether the agent properly utilizes the Azure MCP server's
best practices tools when explicitly instructed to do so, ensuring MCP tool
integration is working end-to-end.
