---
id: identity-dp-dotnet-service-principal
properties:
  service: identity
  plane: data-plane
  language: dotnet
  category: auth
  difficulty: intermediate
  description: 'Can a developer authenticate with a Service Principal (client secret) using the .NET SDK?

    '
  sdk_package: Azure.Identity
  doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/identity-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- service-principal
- client-secret
---

# Service Principal Authentication: Azure Identity (.NET)

## Prompt

Show me how to authenticate
to Azure using a Service Principal with client secret in C#. I need:
1. Required NuGet packages
2. How to create a ClientSecretCredential with tenant ID, client ID, and client secret
3. How to use it with an Azure SDK client (e.g., BlobServiceClient)
4. Best practices for storing the secret (environment variables vs configuration)
5. Error handling when credentials are invalid

Provide a complete example with proper error handling.

## Evaluation Criteria

The generated code should include:
- `Azure.Identity` package with `ClientSecretCredential` class
- Constructor parameters: tenantId, clientId, clientSecret
- Passing credential to Azure SDK clients
- Environment variable patterns for secret storage
- `AuthenticationFailedException` for invalid credentials

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
