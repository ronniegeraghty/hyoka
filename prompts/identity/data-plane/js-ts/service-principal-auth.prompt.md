---
id: identity-dp-js-ts-service-principal
properties:
  service: identity
  plane: data-plane
  language: js-ts
  category: auth
  difficulty: intermediate
  description: 'Can a developer authenticate with a Service Principal (client secret) using the JavaScript/TypeScript SDK?

    '
  sdk_package: '@azure/identity'
  doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/identity-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- service-principal
- client-secret
---

# Service Principal Authentication: Azure Identity (JavaScript/TypeScript)

## Prompt

Show me how to
authenticate to Azure using a Service Principal with client secret. I need:
1. Required npm packages
2. How to create a ClientSecretCredential with tenant ID, client ID, and secret
3. Using it with an Azure SDK client
4. Best practices for secret management in Node.js
5. Error handling for authentication failures

Provide a complete TypeScript example.

## Evaluation Criteria

The generated code should include:
- `@azure/identity` package with `ClientSecretCredential` class
- Constructor parameters: tenantId, clientId, clientSecret
- Passing credential to Azure SDK clients
- dotenv or environment variable patterns
- `AuthenticationError` handling

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
