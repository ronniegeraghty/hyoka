---
id: key-vault-dp-cpp-crud
service: key-vault
plane: data-plane
language: cpp
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the C++ SDK?
sdk_package: azure-security-keyvault-secrets-cpp
doc_url: https://github.com/Azure/azure-sdk-for-cpp/tree/main/sdk/keyvault/azure-security-keyvault-secrets
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (C++)

## Prompt

Write a C++ program that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential from Azure Identity for authentication.
Show the required CMake configuration and vcpkg dependencies.
Include proper exception handling.

## Evaluation Criteria

The generated code should include:
- vcpkg/CMake setup for `azure-security-keyvault-secrets-cpp` and `azure-identity-cpp`
- Creating a `SecretClient` with vault URL and credential
- `SetSecret()`, `GetSecret()`, `StartDeleteSecret()`, `PurgeDeletedSecret()`
- Polling the delete operation to completion
- Exception handling with `Azure::Core::RequestFailedException`

## Context

The C++ SDK requires more setup (CMake, vcpkg) than other languages.
Testing CRUD coverage validates that the generated code covers both the build system
configuration and the API usage for the most fundamental Key Vault scenario.
