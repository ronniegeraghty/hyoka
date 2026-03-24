---
id: identity-dp-python-service-principal
service: identity
plane: data-plane
language: python
category: auth
difficulty: intermediate
description: >
  Can a developer authenticate with a Service Principal (client secret)
  using the Python SDK?
sdk_package: azure-identity
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme
tags:
  - authentication
  - service-principal
  - client-secret
created: 2025-07-28
author: ronniegeraghty
---

# Service Principal Authentication: Azure Identity (Python)

## Prompt

Show me how to authenticate
to Azure using a Service Principal with client secret in Python. I need:
1. Required pip packages
2. How to create a ClientSecretCredential with tenant_id, client_id, and client_secret
3. Using it with an Azure SDK client
4. Best practices for secret management (environment variables, .env files)
5. Error handling for authentication failures

Provide a complete example with proper exception handling.

## Evaluation Criteria

The generated code should include:
- `azure-identity` package with `ClientSecretCredential` class
- Constructor keyword arguments: tenant_id, client_id, client_secret
- Passing credential to Azure SDK clients
- `os.environ` or python-dotenv patterns
- `ClientAuthenticationError` exception handling

## Context

Service Principal authentication with client secrets is the most common pattern
for application-to-application auth in Azure. This tests whether the generated code
covers the full setup including credential creation, usage, and secret management best practices.
