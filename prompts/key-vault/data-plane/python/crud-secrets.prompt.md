---
id: key-vault-dp-python-crud
service: key-vault
plane: data-plane
language: python
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the Python SDK based on the documentation alone?
sdk_package: azure-keyvault-secrets
doc_url: https://learn.microsoft.com/en-us/azure/key-vault/secrets/quick-create-python
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-27
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a script that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Include proper error handling
and show required pip packages.

## Evaluation Criteria

The documentation should cover:
- Installing `azure-keyvault-secrets` and `azure-identity` packages
- Creating a `SecretClient` with vault URL and credential
- `set_secret()`, `get_secret()`, `begin_delete_secret()`, `purge_deleted_secret()`
- Handling soft-delete (waiting for delete to complete before purge)
- Exception handling for `ResourceNotFoundError`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the Python docs provide a complete, runnable flow
covering the full lifecycle including soft-delete behavior.
