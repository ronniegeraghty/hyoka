---
id: identity-dp-java-service-principal
properties:
  service: identity
  plane: data-plane
  language: java
  category: auth
  difficulty: intermediate
  description: 'Can a developer authenticate with a Service Principal (client secret) using the Java SDK?

    '
  sdk_package: azure-identity
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/identity-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- service-principal
- client-secret
---

# Service Principal Authentication: Azure Identity (Java)

## Prompt

Show me how to authenticate
to Azure using a Service Principal with client secret in Java. I need:
1. Required Maven dependencies
2. How to create a ClientSecretCredential with ClientSecretCredentialBuilder
3. Using it with an Azure SDK client
4. Best practices for secret management
5. Error handling for invalid credentials

Provide a complete example with proper exception handling.

## Evaluation Criteria

The generated code should include:
- Maven dependency for `azure-identity`
- `ClientSecretCredentialBuilder` with tenantId, clientId, clientSecret
- Passing credential to Azure SDK client builders
- Environment variable patterns for secret storage
- `AuthenticationException` handling

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
