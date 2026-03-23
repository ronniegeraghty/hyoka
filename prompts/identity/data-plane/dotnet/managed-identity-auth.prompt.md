---
id: identity-dp-dotnet-managed-identity
service: identity
plane: data-plane
language: dotnet
category: auth
difficulty: intermediate
description: >
  Can a developer use Managed Identity to authenticate Azure SDK clients
  using the .NET SDK documentation?
sdk_package: Azure.Identity
doc_url: https://learn.microsoft.com/en-us/dotnet/azure/sdk/authentication/
tags:
  - authentication
  - managed-identity
  - azure-hosted
created: 2025-07-28
author: ronniegeraghty
---

# Managed Identity Authentication: Azure Identity (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, show me how to use
Managed Identity to authenticate Azure SDK clients in C#. Cover:
1. System-assigned vs user-assigned managed identity differences
2. How to create a ManagedIdentityCredential for each type
3. Using it with an Azure SDK client (e.g., KeyVaultClient or BlobServiceClient)
4. How to test locally when managed identity isn't available
5. Common pitfalls and error handling

Provide examples for both system-assigned and user-assigned identity.

## Evaluation Criteria

The documentation should cover:
- `ManagedIdentityCredential` class and constructors
- System-assigned: no parameters needed
- User-assigned: passing the client ID
- Integration with `DefaultAzureCredential` (managed identity in the chain)
- `CredentialUnavailableException` when not running in Azure
- Combining with `ChainedTokenCredential` for local fallback

## Context

Managed Identity is the recommended auth pattern for code running in Azure.
It eliminates the need for managing secrets entirely. This tests whether the
.NET docs explain both system-assigned and user-assigned identity clearly,
including the critical local development fallback story.
