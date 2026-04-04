---
id: identity-dp-go-managed-identity
properties:
  service: identity
  plane: data-plane
  language: go
  category: auth
  difficulty: intermediate
  description: 'Can a developer use Managed Identity to authenticate Azure SDK clients using the Go SDK?

    '
  sdk_package: github.com/Azure/azure-sdk-for-go/sdk/azidentity
  doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- managed-identity
- azure-hosted
---

# Managed Identity Authentication: Azure Identity (Go)

## Prompt

Show me how to use
Managed Identity to authenticate Azure SDK clients in Go. Cover:
1. System-assigned vs user-assigned managed identity
2. How to create a ManagedIdentityCredential for each type
3. Using it with Azure SDK clients
4. Local development fallback strategies
5. Error handling and troubleshooting

Provide Go examples for both identity types.

## Evaluation Criteria

The generated code should include:
- `azidentity.NewManagedIdentityCredential()` function
- System-assigned: nil options
- User-assigned: `ManagedIdentityCredentialOptions{ID: azidentity.ClientID("...")}`
- Integration with `DefaultAzureCredential` chain
- Error handling when not running in Azure
- `azidentity.NewChainedTokenCredential()` for fallback

## Context

Managed Identity is the recommended auth pattern for code running in Azure.
It eliminates the need for managing secrets entirely. This tests whether the
generated code demonstrates both system-assigned and user-assigned identity clearly,
including the critical local development fallback story.
