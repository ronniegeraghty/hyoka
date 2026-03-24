---
id: identity-dp-cpp-service-principal
service: identity
plane: data-plane
language: cpp
category: auth
difficulty: intermediate
description: >
  Can a developer authenticate with a Service Principal (client secret)
  using the C++ SDK?
sdk_package: azure-identity-cpp
doc_url: https://github.com/Azure/azure-sdk-for-cpp/tree/main/sdk/identity/azure-identity
tags:
  - authentication
  - service-principal
  - client-secret
created: 2025-07-28
author: ronniegeraghty
---

# Service Principal Authentication: Azure Identity (C++)

## Prompt

Show me how to authenticate
to Azure using a Service Principal with client secret in C++. I need:
1. Required vcpkg/CMake packages
2. How to create a ClientSecretCredential with tenant ID, client ID, and secret
3. Using it with an Azure SDK client
4. Best practices for secret management in C++
5. Error handling for authentication failures

Provide a complete C++ example with proper exception handling.

## Evaluation Criteria

The generated code should include:
- vcpkg/CMake setup for `azure-identity-cpp`
- `Azure::Identity::ClientSecretCredential` class
- Constructor parameters: tenantId, clientId, clientSecret
- Passing credential to Azure SDK clients
- `Azure::Core::RequestFailedException` handling

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
