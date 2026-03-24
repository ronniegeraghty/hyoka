---
id: key-vault-dp-rust-crud
service: key-vault
plane: data-plane
language: rust
category: crud
difficulty: basic
description: >
  Can a developer create, read, update, and delete secrets in Azure Key Vault
  using the Rust SDK?
sdk_package: azure_security_keyvault_secrets
doc_url: https://docs.rs/azure_security_keyvault_secrets/latest/azure_security_keyvault_secrets/
tags:
  - secrets
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Secrets: Azure Key Vault (Rust)

## Prompt

Write a Rust program that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential from azure_identity for authentication.
Show the Cargo.toml dependencies and include proper error handling with Result types.

## Evaluation Criteria

The generated code should include:
- Cargo.toml dependencies for `azure_security_keyvault_secrets` and `azure_identity`
- Creating a `SecretClient` with vault URL and credential
- Methods for set, get, delete, and purge operations
- Async/await patterns using tokio runtime
- Error handling with `azure_core::Error` and Result types

## Context

The Rust SDK is newer and still evolving. Testing CRUD coverage validates
that the fundamental Key Vault operations are documented for Rust developers.
This also tests whether async patterns and error handling are clearly explained.
