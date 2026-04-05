---
id: key-vault-dp-dotnet-pagination
properties:
  service: key-vault
  plane: data-plane
  language: dotnet
  category: pagination
  difficulty: intermediate
  description: 'Can a developer paginate through a large list of Key Vault secrets using the .NET SDK?

    '
  sdk_package: Azure.Security.KeyVault.Secrets
  doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/security.keyvault.secrets-readme
  created: '2025-07-27'
  author: ronniegeraghty
tags:
- pagination
- list-secrets
- async
---

# Pagination: List Key Vault Secrets (.NET)

## Prompt

Write a C# program that lists all
secrets in an Azure Key Vault that contains hundreds of secrets. The program should:
1. Use SecretClient with DefaultAzureCredential
2. Iterate through secrets page-by-page using AsyncPageable
3. Print the name, content type, and enabled status of each secret
4. Handle the case where some secrets are disabled
5. Show both sync and async iteration patterns

I want to understand how Azure.Page<T> and AsyncPageable<T> work
for large result sets. Show required NuGet packages.

## Evaluation Criteria

- `SecretClient.GetPropertiesOfSecretsAsync()` returning `AsyncPageable<SecretProperties>`
- `await foreach` pattern for async iteration
- `AsPages()` for explicit page-by-page control
- Page size hints via `GetPropertiesOfSecretsAsync(cancellationToken)`
- Accessing `SecretProperties` fields (Name, ContentType, Enabled, CreatedOn)
- Sync alternative using `Pageable<SecretProperties>`
- Error handling during pagination

## Context

Key Vaults in enterprise environments often contain hundreds of secrets.
Developers need to understand the AsyncPageable pattern to efficiently
enumerate them without loading all results into memory at once.
