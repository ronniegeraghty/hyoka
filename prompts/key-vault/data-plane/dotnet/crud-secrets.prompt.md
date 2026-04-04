---
id: key-vault-dp-dotnet-crud
properties:
  service: key-vault
  plane: data-plane
  language: dotnet
  category: crud
  difficulty: basic
  description: 'Can a developer create, read, update, and delete secrets in Azure Key Vault using the .NET SDK?

    '
  sdk_package: Azure.Security.KeyVault.Secrets
  doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/security.keyvault.secrets-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- secrets
- crud
- getting-started
---

# CRUD Secrets: Azure Key Vault (.NET)

## Prompt

Write a C# console application that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Include proper error handling
and show required NuGet packages.

## Evaluation Criteria

The generated code should include:
- Installing `Azure.Security.KeyVault.Secrets` and `Azure.Identity` NuGet packages
- Creating a `SecretClient` with vault URI and credential
- `SetSecret()`, `GetSecret()`, `StartDeleteSecret()`, `PurgeDeletedSecret()`
- Handling soft-delete (polling `DeleteSecretOperation` to completion before purge)
- Exception handling for `RequestFailedException`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the generated code provides a complete, runnable flow
covering the full lifecycle including soft-delete and async polling patterns.
