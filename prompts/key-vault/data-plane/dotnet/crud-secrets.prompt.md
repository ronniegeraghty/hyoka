---
id: key-vault-dp-dotnet-crud
service: key-vault
plane: data-plane
language: dotnet
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the .NET SDK based on the documentation alone?
sdk_package: Azure.Security.KeyVault.Secrets
doc_url: https://learn.microsoft.com/en-us/azure/key-vault/secrets/quick-create-net
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, write a C# console application that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Include proper error handling
and show required NuGet packages.

## Evaluation Criteria

The documentation should cover:
- Installing `Azure.Security.KeyVault.Secrets` and `Azure.Identity` NuGet packages
- Creating a `SecretClient` with vault URI and credential
- `SetSecret()`, `GetSecret()`, `StartDeleteSecret()`, `PurgeDeletedSecret()`
- Handling soft-delete (polling `DeleteSecretOperation` to completion before purge)
- Exception handling for `RequestFailedException`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the .NET docs provide a complete, runnable flow
covering the full lifecycle including soft-delete and async polling patterns.
