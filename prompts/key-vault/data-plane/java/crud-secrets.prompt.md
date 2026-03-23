---
id: key-vault-dp-java-crud
service: key-vault
plane: data-plane
language: java
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the Java SDK based on the documentation alone?
sdk_package: azure-security-keyvault-secrets
doc_url: https://learn.microsoft.com/en-us/azure/key-vault/secrets/quick-create-java
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (Java)

## Prompt

Using only the Azure SDK for Java documentation, write a Java application that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Show the Maven dependency
for azure-security-keyvault-secrets and azure-identity. Include proper exception handling.

## Evaluation Criteria

The documentation should cover:
- Maven dependency for `azure-security-keyvault-secrets` and `azure-identity`
- Creating a `SecretClient` with `SecretClientBuilder`
- `setSecret()`, `getSecret()`, `beginDeleteSecret()`, `purgeDeletedSecret()`
- Using `SyncPoller` to wait for delete completion before purge
- Exception handling for `HttpResponseException`

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the Java docs provide a complete, runnable flow
covering the full lifecycle including the SyncPoller pattern for long-running deletes.
