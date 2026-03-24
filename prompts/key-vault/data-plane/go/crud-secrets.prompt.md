---
id: key-vault-dp-go-crud
service: key-vault
plane: data-plane
language: go
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the Go SDK?
sdk_package: github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets
doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (Go)

## Prompt

Write a Go program that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential from azidentity for authentication.
Show the go module imports and include proper error handling.

## Evaluation Criteria

The generated code should include:
- Importing `github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets`
- Importing `github.com/Azure/azure-sdk-for-go/sdk/azidentity`
- Creating a client with `azsecrets.NewClient()`
- `SetSecret()`, `GetSecret()`, `DeleteSecret()`, `PurgeDeletedSecret()`
- Handling the `*azcore.ResponseError` type for errors

## Context

CRUD operations on secrets are the most fundamental Key Vault use case.
This tests whether the generated code provides a complete, runnable flow covering
the full lifecycle. Go's error handling pattern (checking returned errors)
makes this a good test of code generation completeness.
