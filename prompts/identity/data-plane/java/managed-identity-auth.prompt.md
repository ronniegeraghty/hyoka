---
id: identity-dp-java-managed-identity
service: identity
plane: data-plane
language: java
category: auth
difficulty: intermediate
description: >
  Can a developer use Managed Identity to authenticate Azure SDK clients
  using the Java SDK?
sdk_package: azure-identity
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/identity-readme
tags:
  - authentication
  - managed-identity
  - azure-hosted
created: 2025-07-28
author: ronniegeraghty
---

# Managed Identity Authentication: Azure Identity (Java)

## Prompt

Show me how to use
Managed Identity to authenticate Azure SDK clients in Java. Cover:
1. System-assigned vs user-assigned managed identity
2. How to create a ManagedIdentityCredential for each type
3. Using it with Azure SDK clients
4. Local development fallback strategies
5. Error handling and troubleshooting

Provide examples for both system-assigned and user-assigned identity.

## Evaluation Criteria

The generated code should include:
- `ManagedIdentityCredentialBuilder` class
- System-assigned: default builder with no client ID
- User-assigned: `.clientId()` on the builder
- Integration with `DefaultAzureCredential` chain
- `CredentialUnavailableException` when not in Azure
- `ChainedTokenCredentialBuilder` for local fallback

## Context

Managed Identity is the recommended auth pattern for code running in Azure.
It eliminates the need for managing secrets entirely. This tests whether the
generated code demonstrates both system-assigned and user-assigned identity clearly,
including the critical local development fallback story.
