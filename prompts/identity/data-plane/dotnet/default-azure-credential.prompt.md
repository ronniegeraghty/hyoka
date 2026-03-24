---
id: identity-dp-dotnet-default-credential
service: identity
plane: data-plane
language: dotnet
category: auth
difficulty: basic
description: >
  Can a developer set up DefaultAzureCredential for Azure SDK clients
  using the .NET SDK?
sdk_package: Azure.Identity
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/identity-readme
tags:
  - authentication
  - default-azure-credential
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# DefaultAzureCredential: Azure Identity (.NET)

## Prompt

Show me how to authenticate
an Azure SDK client using DefaultAzureCredential in C#. Explain:
1. What NuGet packages are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain order (which credentials are tried and in what sequence)
4. How it behaves differently in local development vs deployed Azure environments
5. How to troubleshoot when authentication fails

Provide a complete, runnable example that creates a BlobServiceClient using
DefaultAzureCredential.

## Evaluation Criteria

The generated code should include:
- `Azure.Identity` NuGet package installation
- `DefaultAzureCredential` constructor and options
- Credential chain: Environment → Workload Identity → Managed Identity → Azure CLI → etc.
- Passing credential to any Azure SDK client constructor
- `AuthenticationFailedException` handling and diagnostics

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the generated code demonstrates it clearly enough for first-time users.
