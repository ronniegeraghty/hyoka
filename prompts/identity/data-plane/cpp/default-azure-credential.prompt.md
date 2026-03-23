---
id: identity-dp-cpp-default-credential
service: identity
plane: data-plane
language: cpp
category: auth
difficulty: basic
description: >
  Can a developer set up DefaultAzureCredential for Azure SDK clients
  using the C++ SDK documentation?
sdk_package: azure-identity-cpp
doc_url: https://github.com/Azure/azure-sdk-for-cpp/tree/main/sdk/identity/azure-identity
tags:
  - authentication
  - default-azure-credential
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# DefaultAzureCredential: Azure Identity (C++)

## Prompt

Using only the Azure SDK for C++ documentation, show me how to authenticate
an Azure SDK client using DefaultAzureCredential. Explain:
1. What vcpkg/CMake dependencies are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain order and which credentials are tried
4. How it works in local development vs Azure-hosted environments
5. How to handle authentication errors

Provide a complete C++ example that creates a Key Vault SecretClient using
DefaultAzureCredential with proper exception handling.

## Evaluation Criteria

The documentation should cover:
- vcpkg/CMake setup for `azure-identity-cpp`
- `Azure::Identity::DefaultAzureCredential` class
- Credential chain order in C++ SDK
- Passing `std::shared_ptr<TokenCredential>` to client constructors
- Exception handling with `Azure::Core::RequestFailedException`

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the C++ docs explain it clearly enough for first-time users.
