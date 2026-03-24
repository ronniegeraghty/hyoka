---
id: identity-dp-rust-service-principal
service: identity
plane: data-plane
language: rust
category: auth
difficulty: intermediate
description: >
  Can a developer authenticate with a Service Principal (client secret)
  using the Rust SDK?
sdk_package: azure_identity
doc_url: https://docs.rs/azure_identity/latest/azure_identity/
tags:
  - authentication
  - service-principal
  - client-secret
created: 2025-07-28
author: ronniegeraghty
---

# Service Principal Authentication: Azure Identity (Rust)

## Prompt

Show me how to authenticate
to Azure using a Service Principal with client secret in Rust. I need:
1. Required Cargo dependencies
2. How to create a ClientSecretCredential
3. Using it with an Azure SDK client
4. Best practices for secret management in Rust
5. Error handling for authentication failures

Provide a complete Rust example with proper error handling.

## Evaluation Criteria

The generated code should include:
- `azure_identity` crate with `ClientSecretCredential`
- Constructor with tenant_id, client_id, client_secret
- Passing credential to Azure SDK clients
- Environment variable patterns with `std::env::var()`
- `azure_core::Error` handling

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
