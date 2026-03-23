---
id: identity-dp-go-default-credential
service: identity
plane: data-plane
language: go
category: auth
difficulty: basic
description: >
  Can a developer set up DefaultAzureCredential for Azure SDK clients
  using the Go SDK documentation?
sdk_package: github.com/Azure/azure-sdk-for-go/sdk/azidentity
doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity
tags:
  - authentication
  - default-azure-credential
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# DefaultAzureCredential: Azure Identity (Go)

## Prompt

Using only the Azure SDK for Go documentation, show me how to authenticate
an Azure SDK client using DefaultAzureCredential. Explain:
1. What Go modules are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain order and which credentials are tried
4. How it works in local development vs Azure-hosted environments
5. How to troubleshoot authentication failures

Provide a complete Go example that creates a Key Vault SecretClient using
DefaultAzureCredential.

## Evaluation Criteria

The documentation should cover:
- `azidentity` module import
- `azidentity.NewDefaultAzureCredential()` function
- Credential chain order in Go SDK
- Passing credential to client constructors
- Error handling with `*azcore.ResponseError`

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the Go docs explain it clearly enough for first-time users.
