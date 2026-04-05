---
id: key-vault-dp-python-crud
properties:
  service: key-vault
  plane: data-plane
  language: python
  category: crud
  difficulty: basic
  description: 'Can a developer create, read, update, and delete secrets in Azure Key Vault using the Python SDK?

    '
  sdk_package: azure-keyvault-secrets
  doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
  created: '2025-07-27'
  author: ronniegeraghty
tags:
- secrets
- crud
- getting-started
---

# CRUD Secrets: Azure Key Vault (Python)

## Prompt

Write a script that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Include proper error handling
and show required pip packages.

## Evaluation Criteria

The generated code should include:
- Installing `azure-keyvault-secrets` and `azure-identity` packages
- Creating a `SecretClient` with vault URL and credential
- `set_secret()`, `get_secret()`, `begin_delete_secret()`, `purge_deleted_secret()`
- Handling soft-delete (waiting for delete to complete before purge)
- Exception handling for `ResourceNotFoundError`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the generated code provides a complete, runnable flow
covering the full lifecycle including soft-delete behavior.
