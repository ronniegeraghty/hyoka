---
id: key-vault-dp-js-ts-crud
service: key-vault
plane: data-plane
language: js-ts
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the JavaScript/TypeScript SDK?
sdk_package: "@azure/keyvault-secrets"
doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/keyvault-secrets-readme
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (JavaScript/TypeScript)

## Prompt

Write a Node.js script
(TypeScript preferred) that performs all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Show required npm packages
and include proper error handling with try/catch.

## Evaluation Criteria

The generated code should include:
- Installing `@azure/keyvault-secrets` and `@azure/identity` npm packages
- Creating a `SecretClient` with vault URL and credential
- `setSecret()`, `getSecret()`, `beginDeleteSecret()`, `purgeDeletedSecret()`
- Awaiting the `DeleteSecretPoller` before purging
- Error handling for `RestError`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the generated code provides a complete, runnable flow
covering the full lifecycle including the async poller pattern for soft-delete.
