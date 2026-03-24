---
id: key-vault-dp-dotnet-error-handling
service: key-vault
plane: data-plane
language: dotnet
category: error-handling
difficulty: intermediate
description: >
  Can a developer handle Key Vault errors including
  access denied (403), secret not found (404), and throttling (429) in .NET?
sdk_package: Azure.Security.KeyVault.Secrets
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/security.keyvault.secrets-readme
tags:
  - error-handling
  - exceptions
  - access-policy
  - throttling
created: 2025-07-27
author: ronniegeraghty
---

# Error Handling: Azure Key Vault Secrets (.NET)

## Prompt

How do I properly handle errors when working with Azure Key Vault secrets in .NET?
I need to handle common failures: access denied when RBAC or access policies
aren't configured correctly (403), secret not found (404), secret version
conflicts, and throttling when hitting Key Vault rate limits (429).
Show me try/catch patterns with Azure.Security.KeyVault.Secrets including
how to extract the error code and HTTP status from RequestFailedException.

## Evaluation Criteria

- `RequestFailedException` for all Key Vault errors
- Extracting `Status` and `ErrorCode` properties
- 403 handling: diagnosing RBAC vs. access policy misconfiguration
- 404 handling: secret not found vs. deleted secret
- 429 throttling: Key Vault rate limits and retry behavior
- Soft-delete and purge protection error scenarios
- `SecretClientOptions` retry configuration

## Context

Key Vault errors are often caused by misconfigured access policies or RBAC roles,
which are notoriously difficult to debug. The generated code needs to demonstrate
not just how to catch exceptions but how to diagnose the root cause of 403 errors.
