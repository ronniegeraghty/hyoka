---
id: identity-dp-go-service-principal
service: identity
plane: data-plane
language: go
category: auth
difficulty: intermediate
description: >
  Can a developer authenticate with a Service Principal (client secret)
  using the Go SDK?
sdk_package: github.com/Azure/azure-sdk-for-go/sdk/azidentity
doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity
tags:
  - authentication
  - service-principal
  - client-secret
created: 2025-07-28
author: ronniegeraghty
---

# Service Principal Authentication: Azure Identity (Go)

## Prompt

Show me how to authenticate
to Azure using a Service Principal with client secret in Go. I need:
1. Required Go modules
2. How to create a ClientSecretCredential with NewClientSecretCredential
3. Using it with an Azure SDK client
4. Best practices for secret management in Go
5. Error handling for authentication failures

Provide a complete Go example with proper error handling.

## Evaluation Criteria

The generated code should include:
- `azidentity` module with `NewClientSecretCredential()` function
- Parameters: tenantID, clientID, clientSecret, options
- Passing credential to Azure SDK client constructors
- Environment variable patterns with `os.Getenv()`
- Error handling patterns

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
