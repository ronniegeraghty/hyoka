---
id: identity-dp-cpp-managed-identity
properties:
  service: identity
  plane: data-plane
  language: cpp
  category: auth
  difficulty: intermediate
  description: 'Can a developer use Managed Identity to authenticate Azure SDK clients using the C++ SDK?

    '
  sdk_package: azure-identity-cpp
  doc_url: https://github.com/Azure/azure-sdk-for-cpp/tree/main/sdk/identity/azure-identity
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- managed-identity
- azure-hosted
---

# Managed Identity Authentication: Azure Identity (C++)

## Prompt

Show me how to use
Managed Identity to authenticate Azure SDK clients in C++. Cover:
1. System-assigned vs user-assigned managed identity
2. How to create a ManagedIdentityCredential for each type
3. Using it with Azure SDK clients
4. Local development fallback strategies
5. Error handling and troubleshooting

Provide C++ examples for both identity types.

## Evaluation Criteria

The generated code should include:
- `Azure::Identity::ManagedIdentityCredential` class
- System-assigned: default constructor
- User-assigned: passing client ID
- Integration with `DefaultAzureCredential` chain
- `Azure::Core::Credentials::AuthenticationException` handling
- `ChainedTokenCredential` for local fallback

## Context

Managed Identity is the recommended auth pattern for code running in Azure.
It eliminates the need for managing secrets entirely. This tests whether the
generated code demonstrates both system-assigned and user-assigned identity clearly,
including the critical local development fallback story.
