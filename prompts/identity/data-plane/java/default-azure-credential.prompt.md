---
id: identity-dp-java-default-credential
properties:
  service: identity
  plane: data-plane
  language: java
  category: auth
  difficulty: basic
  description: 'Can a developer set up DefaultAzureCredential for Azure SDK clients using the Java SDK?

    '
  sdk_package: azure-identity
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/identity-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- authentication
- default-azure-credential
- getting-started
---

# DefaultAzureCredential: Azure Identity (Java)

## Prompt

Show me how to authenticate
an Azure SDK client using DefaultAzureCredential in Java. Explain:
1. What Maven dependencies are needed
2. How to create and use a DefaultAzureCredential instance
3. The credential chain order and which credentials are tried
4. How it behaves differently in local development vs Azure environments
5. How to troubleshoot authentication failures

Provide a complete example that creates a SecretClient using DefaultAzureCredential.

## Evaluation Criteria

The generated code should include:
- Maven dependency for `azure-identity`
- `DefaultAzureCredentialBuilder` pattern
- Credential chain order in Java SDK
- Passing credential to client builders (e.g., `SecretClientBuilder`)
- Logging configuration for authentication troubleshooting

## Context

DefaultAzureCredential is the recommended starting point for Azure SDK authentication.
It abstracts away the complexity of credential selection and works across environments.
This tests whether the generated code demonstrates it clearly enough for first-time users.
