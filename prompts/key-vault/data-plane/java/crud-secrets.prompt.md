---
id: key-vault-dp-java-crud
service: key-vault
plane: data-plane
language: java
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the Java SDK?
sdk_package: azure-security-keyvault-secrets
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/security-keyvault-secrets-readme
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (Java)

## Prompt

Write a Java application that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Show the Maven dependency
for azure-security-keyvault-secrets and azure-identity. Include proper exception handling.

## Evaluation Criteria

The generated code should include:
- Maven dependency for `azure-security-keyvault-secrets` and `azure-identity`
- Creating a `SecretClient` with `SecretClientBuilder`
- `setSecret()`, `getSecret()`, `beginDeleteSecret()`, `purgeDeletedSecret()`
- Using `SyncPoller` to wait for delete completion before purge
- Exception handling for `HttpResponseException`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the generated code provides a complete, runnable flow
covering the full lifecycle including the SyncPoller pattern for long-running deletes.
