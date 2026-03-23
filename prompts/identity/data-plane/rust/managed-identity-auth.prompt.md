---
id: identity-dp-rust-managed-identity
service: identity
plane: data-plane
language: rust
category: auth
difficulty: intermediate
description: >
  Can a developer use Managed Identity to authenticate Azure SDK clients
  using the Rust SDK documentation?
sdk_package: azure_identity
doc_url: https://docs.rs/azure_identity/latest/azure_identity/
tags:
  - authentication
  - managed-identity
  - azure-hosted
created: 2025-07-28
author: ronniegeraghty
---

# Managed Identity Authentication: Azure Identity (Rust)

## Prompt

Using only the Azure SDK for Rust documentation, show me how to use
Managed Identity to authenticate Azure SDK clients in Rust. Cover:
1. System-assigned vs user-assigned managed identity
2. How to create credentials for each type
3. Using it with Azure SDK clients
4. Local development fallback strategies
5. Error handling when not running in Azure

Provide Rust examples for both identity types.

## Evaluation Criteria

The documentation should cover:
- Managed identity credential types in `azure_identity`
- System-assigned: default construction
- User-assigned: passing client ID
- Integration with `DefaultAzureCredential` chain
- Error handling with `azure_core::Error`
- Chained credential for local fallback

## Context

Managed Identity is the recommended auth pattern for code running in Azure.
It eliminates the need for managing secrets entirely. This tests whether the
Rust docs explain both system-assigned and user-assigned identity clearly,
including the critical local development fallback story.
