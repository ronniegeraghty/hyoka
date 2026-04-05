---
id: identity-dp-rust-default-credential
properties:
  service: identity
  plane: data-plane
  language: rust
  category: auth
  difficulty: basic
  description: 'Can a developer set up DefaultAzureCredential for Azure SDK clients using the Rust SDK?

    '
  sdk_package: azure_identity
  doc_url: https://docs.rs/azure_identity/latest/azure_identity/
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- default-azure-credential
- getting-started
---

# DefaultAzureCredential: Azure Identity (Rust)

## Prompt

Show me how to authenticate
an Azure SDK client using DefaultAzureCredential. Explain:
1. What Cargo dependencies are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain and which credentials are tried
4. How it works in local development vs Azure-hosted environments
5. How to handle authentication errors

Provide a complete Rust example that creates a Key Vault SecretClient using
DefaultAzureCredential with proper error handling.

## Evaluation Criteria

The generated code should include:
- Cargo.toml dependencies for `azure_identity`
- `DefaultAzureCredential::new()` or builder pattern
- Credential chain order in Rust SDK
- Passing credential via `Arc<dyn TokenCredential>`
- Error handling with `azure_core::Error`

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the generated code demonstrates it clearly enough for first-time users.
