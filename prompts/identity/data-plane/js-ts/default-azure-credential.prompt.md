---
id: identity-dp-js-ts-default-credential
service: identity
plane: data-plane
language: js-ts
category: auth
difficulty: basic
description: >
  Can a developer set up DefaultAzureCredential for Azure SDK clients
  using the JavaScript/TypeScript SDK?
sdk_package: "@azure/identity"
doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/identity-readme
tags:
  - authentication
  - default-azure-credential
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# DefaultAzureCredential: Azure Identity (JavaScript/TypeScript)

## Prompt

Show me how to
authenticate an Azure SDK client using DefaultAzureCredential. Explain:
1. What npm packages are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain order and which credentials are tried
4. How it works in local dev (VS Code, Azure CLI) vs Azure-hosted environments
5. How to troubleshoot authentication failures

Provide a complete TypeScript example that creates a SecretClient using
DefaultAzureCredential.

## Evaluation Criteria

The generated code should include:
- `@azure/identity` npm package installation
- `DefaultAzureCredential` constructor and options
- Credential chain: Environment → Workload Identity → Managed Identity → Azure CLI → etc.
- Passing credential to Azure SDK clients
- `AuthenticationError` handling and logging

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the generated code demonstrates it clearly enough for first-time users.
