---
id: identity-dp-python-default-credential
properties:
  service: identity
  plane: data-plane
  language: python
  category: auth
  difficulty: basic
  description: 'Can a developer set up DefaultAzureCredential for Azure SDK clients using the Python SDK?

    '
  sdk_package: azure-identity
  doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- default-azure-credential
- getting-started
---

# DefaultAzureCredential: Azure Identity (Python)

## Prompt

Show me how to authenticate
an Azure SDK client using DefaultAzureCredential. Explain:
1. What pip packages are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain order and which credentials are tried
4. How it works in local development (VS Code, Azure CLI) vs Azure deployments
5. How to troubleshoot authentication failures with logging

Provide a complete example that creates a SecretClient using DefaultAzureCredential.

## Evaluation Criteria

The generated code should include:
- `azure-identity` pip package installation
- `DefaultAzureCredential()` constructor and keyword arguments
- Credential chain: Environment → Workload Identity → Managed Identity → Azure CLI → etc.
- Passing credential to Azure SDK clients
- `ClientAuthenticationError` handling and `logging` module configuration

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the generated code demonstrates it clearly enough for first-time users.
